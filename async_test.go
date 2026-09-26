package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestAsyncWriterZero(t *testing.T) {
	w := &AsyncWriter{
		ChannelSize: 0,
		Writer:      IOWriter{os.Stderr},
	}
	for i := range 10 {
		_, _ = wlprintf(w, InfoLevel, "%s, %d during async writer 1k buff size\n", timeNow(), i)
	}
	if err := w.Close(); err != nil {
		t.Errorf("async close error: %+v", err)
	}
}

func TestAsyncWriterSmall(t *testing.T) {
	w := &AsyncWriter{
		ChannelSize: 5,
		Writer:      IOWriter{os.Stderr},
	}
	for i := range 10 {
		_, _ = wlprintf(w, InfoLevel, "%s, %d during async writer 1k buff size\n", timeNow(), i)
	}
	if err := w.Close(); err != nil {
		t.Errorf("async close error: %+v", err)
	}
}

func TestAsyncWriterSize(t *testing.T) {
	writer1 := &FileWriter{
		Filename: "async_file_test1.log",
	}

	writer2 := &AsyncWriter{
		ChannelSize:   4096,
		DisableWritev: false,
		DiscardOnFull: false,
		Writer: &FileWriter{
			Filename: "async_file_test2.log",
		},
	}

	logger := Logger{
		Writer: &MultiEntryWriter{
			writer1,
			writer2,
		},
	}

	for range 100000 {
		logger.Info().Msg("hello file writer")
	}

	if err := writer1.Close(); err != nil {
		t.Errorf("file writer close error: %+v", err)
	}

	if err := writer2.Close(); err != nil {
		t.Errorf("async file writer close error: %+v", err)
	}

	fi1, err := os.Stat(writer1.Filename)
	if err != nil {
		t.Errorf("file writer stat error: %+v", err)
	}

	fi2, err := os.Stat(writer2.Writer.(*FileWriter).Filename)
	if err != nil {
		t.Errorf("async file writer stat error: %+v", err)
	}

	if fi1.Size() != fi2.Size() {
		t.Errorf("filesize not equal: %v != %v", fi1.Size(), fi2.Size())
	}
}

func BenchmarkSyncFileWriter(b *testing.B) {
	logger := Logger{
		Writer: &FileWriter{
			Filename: "sync_file_test.log",
		},
	}
	defer logger.Writer.(io.Closer).Close()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			logger.Info().Msg("hello file writer")
		}
	})
}

func BenchmarkAsyncFileWriter(b *testing.B) {
	logger := Logger{
		Writer: &AsyncWriter{
			ChannelSize:   4096,
			DisableWritev: false,
			DiscardOnFull: false,
			Writer: &FileWriter{
				Filename: "async_file_test2.log",
			},
		},
	}
	defer logger.Writer.(io.Closer).Close()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			logger.Info().Msg("hello file writer")
		}
	})
}

func BenchmarkAsyncIOWriter(b *testing.B) {
	logger := TSVLogger{
		Writer: &AsyncWriter{
			ChannelSize:   4096,
			DisableWritev: false,
			DiscardOnFull: false,
			Writer: &FileWriter{
				Filename: "async_file_test3.log",
			},
		},
	}
	defer logger.Writer.(io.Closer).Close()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			logger.New().Str("hello file writer").Msg()
		}
	})
}

// holdingWriter blocks every WriteEntry until release is closed, and reports
// each call on started, so a test can pin the background writer in place.
type holdingWriter struct {
	started chan string
	release chan struct{}
	mu      sync.Mutex
	got     []string
}

func newHoldingWriter() *holdingWriter {
	return &holdingWriter{
		started: make(chan string, 64),
		release: make(chan struct{}),
	}
}

func (w *holdingWriter) WriteEntry(e *Entry) (int, error) {
	w.started <- string(e.buf)
	<-w.release
	w.mu.Lock()
	w.got = append(w.got, string(e.buf))
	w.mu.Unlock()
	return len(e.buf), nil
}

func (w *holdingWriter) entries() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return slices.Clone(w.got)
}

// fillAsyncWriter parks the background writer on "a" and then queues
// "b".."z" up to size entries, leaving the queue full.
func fillAsyncWriter(t *testing.T, w *AsyncWriter, hw *holdingWriter, size int) []string {
	t.Helper()
	if _, err := w.Write([]byte("a")); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if got := <-hw.started; got != "a" {
		t.Fatalf("writer started with %q, want a", got)
	}
	want := []string{"a"}
	for i := range size {
		s := string(rune('b' + i))
		if _, err := w.Write([]byte(s)); err != nil {
			t.Fatalf("write %s: %v", s, err)
		}
		want = append(want, s)
	}
	return want
}

func TestAsyncWriterBlocksWhenFull(t *testing.T) {
	hw := newHoldingWriter()
	w := &AsyncWriter{ChannelSize: 2, DisableWritev: true, Writer: hw}
	want := fillAsyncWriter(t, w, hw, 2)

	returned := make(chan error, 1)
	go func() {
		_, err := w.Write([]byte("d"))
		returned <- err
	}()
	select {
	case err := <-returned:
		t.Fatalf("Write on a full queue returned early: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(hw.release)
	select {
	case err := <-returned:
		if err != nil {
			t.Fatalf("blocked write: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("blocked write was never released")
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if got, want := hw.entries(), append(want, "d"); !slices.Equal(got, want) {
		t.Fatalf("entries = %q, want %q", got, want)
	}
}

func TestAsyncWriterDiscardOnFull(t *testing.T) {
	// ChannelSize 0 means the documented default of 1
	for _, size := range []uint{0, 1, 3} {
		t.Run(fmt.Sprintf("ChannelSize=%d", size), func(t *testing.T) {
			hw := newHoldingWriter()
			w := &AsyncWriter{ChannelSize: size, DiscardOnFull: true, DisableWritev: true, Writer: hw}
			want := fillAsyncWriter(t, w, hw, max(int(size), 1))

			if n, err := w.Write([]byte("full")); n != 0 || err != ErrAsyncWriterFull {
				t.Fatalf("Write on a full queue = (%d, %v), want (0, %v)", n, err, ErrAsyncWriterFull)
			}

			close(hw.release)
			if err := w.Close(); err != nil {
				t.Fatalf("close: %v", err)
			}
			if got := hw.entries(); !slices.Equal(got, want) {
				t.Fatalf("entries = %q, want %q", got, want)
			}
		})
	}
}

// Close must not leave a producer stuck on a full queue: the producer fails
// with ErrAsyncWriterClosed while the entries accepted earlier still drain.
func TestAsyncWriterCloseReleasesBlockedProducer(t *testing.T) {
	hw := newHoldingWriter()
	w := &AsyncWriter{ChannelSize: 1, DisableWritev: true, Writer: hw}
	want := fillAsyncWriter(t, w, hw, 1)

	returned := make(chan error, 1)
	go func() {
		_, err := w.Write([]byte("late"))
		returned <- err
	}()
	closed := make(chan error, 1)
	go func() {
		closed <- w.Close()
	}()

	select {
	case err := <-returned:
		if err != ErrAsyncWriterClosed {
			t.Fatalf("blocked write = %v, want %v", err, ErrAsyncWriterClosed)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not release the blocked producer")
	}

	close(hw.release)
	select {
	case err := <-closed:
		if err != nil {
			t.Fatalf("close: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not return")
	}
	if got := hw.entries(); !slices.Equal(got, want) {
		t.Fatalf("entries = %q, want %q", got, want)
	}
}

type closeCountWriter struct {
	cerr   error
	closes int
}

func (w *closeCountWriter) WriteEntry(e *Entry) (int, error) { return len(e.buf), nil }

func (w *closeCountWriter) Close() error {
	w.closes++
	return w.cerr
}

func TestAsyncWriterCloseTwice(t *testing.T) {
	cerr := errors.New("close boom")
	cw := &closeCountWriter{cerr: cerr}
	w := &AsyncWriter{DisableWritev: true, Writer: cw}
	if _, err := w.Write([]byte("hello")); err != nil {
		t.Fatalf("write: %v", err)
	}
	for i := range 2 {
		if err := w.Close(); err != cerr {
			t.Fatalf("Close #%d = %v, want %v", i+1, err, cerr)
		}
	}
	if cw.closes != 1 {
		t.Fatalf("underlying Close called %d times, want 1", cw.closes)
	}
}

func TestAsyncWriterWriteAfterClose(t *testing.T) {
	writers := map[string]func(t *testing.T) *AsyncWriter{
		"writer": func(t *testing.T) *AsyncWriter {
			return &AsyncWriter{DisableWritev: true, Writer: IOWriter{io.Discard}}
		},
		"writev": func(t *testing.T) *AsyncWriter {
			return &AsyncWriter{Writer: &FileWriter{Filename: filepath.Join(tempLogDir(t), "out.log")}}
		},
	}
	for name, newWriter := range writers {
		t.Run(name, func(t *testing.T) {
			// a writer that was never used can be closed, twice
			idle := newWriter(t)
			for range 2 {
				if err := idle.Close(); err != nil {
					t.Fatalf("close idle writer: %v", err)
				}
			}

			w := newWriter(t)
			if _, err := w.Write([]byte("hello\n")); err != nil {
				t.Fatalf("write: %v", err)
			}
			if err := w.Close(); err != nil {
				t.Fatalf("close: %v", err)
			}
			if n, err := w.Write([]byte("late\n")); n != 0 || err != ErrAsyncWriterClosed {
				t.Fatalf("Write after Close = (%d, %v), want (0, %v)", n, err, ErrAsyncWriterClosed)
			}
			e := &Entry{buf: []byte("late\n")}
			if n, err := w.WriteEntry(e); n != 0 || err != ErrAsyncWriterClosed {
				t.Fatalf("WriteEntry after Close = (%d, %v), want (0, %v)", n, err, ErrAsyncWriterClosed)
			}
		})
	}
}

// orderWriter records every entry it receives, in order.
type orderWriter struct {
	got []string
}

func (w *orderWriter) WriteEntry(e *Entry) (int, error) {
	w.got = append(w.got, string(e.buf))
	return len(e.buf), nil
}

func TestAsyncWriterGenericConcurrentProducers(t *testing.T) {
	const (
		producers = 8
		per       = 2000
	)
	for _, size := range []uint{1, 16, 4096} {
		t.Run(fmt.Sprintf("ChannelSize=%d", size), func(t *testing.T) {
			ow := &orderWriter{}
			w := &AsyncWriter{ChannelSize: size, DisableWritev: true, Writer: ow}

			var wg sync.WaitGroup
			for p := range producers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := range per {
						if _, err := w.Write([]byte(fmt.Sprintf("p%d-%06d", p, i))); err != nil {
							t.Errorf("producer %d: %v", p, err)
							return
						}
					}
				}()
			}
			wg.Wait()
			if err := w.Close(); err != nil {
				t.Fatalf("close: %v", err)
			}

			// every entry arrives exactly once, and in order per producer
			next := make([]int, producers)
			for _, s := range ow.got {
				var p, i int
				if _, err := fmt.Sscanf(s, "p%d-%06d", &p, &i); err != nil || p < 0 || p >= producers {
					t.Fatalf("corrupt entry %q", s)
				}
				if i != next[p] {
					t.Fatalf("producer %d: got entry %d, want %d", p, i, next[p])
				}
				next[p]++
			}
			if len(ow.got) != producers*per {
				t.Fatalf("got %d entries, want %d", len(ow.got), producers*per)
			}
		})
	}
}

func TestAsyncQueueWrapAround(t *testing.T) {
	var q asyncQueue
	q.init(3)

	entries := make([]*Entry, 16)
	for i := range entries {
		entries[i] = &Entry{}
	}
	put := func(i int) {
		t.Helper()
		if err := q.put(entries[i], true); err != nil {
			t.Fatalf("put %d: %v", i, err)
		}
	}
	get := func(max int, want ...int) {
		t.Helper()
		dst := make([]*Entry, max)
		n, done := q.get(dst, false)
		if done {
			t.Fatalf("get reported done on an open queue")
		}
		if n != len(want) {
			t.Fatalf("get moved %d entries, want %d", n, len(want))
		}
		for k, i := range want {
			if dst[k] != entries[i] {
				t.Fatalf("get slot %d holds the wrong entry, want entry %d", k, i)
			}
		}
		// taken slots must not keep entries alive
		live := 0
		for i := range q.buf {
			if q.buf[i] != nil {
				live++
			}
		}
		if live != q.size {
			t.Fatalf("queue holds %d entries in its slots, want %d", live, q.size)
		}
	}

	put(0)
	put(1)
	put(2)
	if err := q.put(entries[3], true); err != ErrAsyncWriterFull {
		t.Fatalf("put on a full queue = %v, want %v", err, ErrAsyncWriterFull)
	}
	get(2, 0, 1)
	put(3)
	put(4) // wraps around
	get(8, 2, 3, 4)
	get(8)
	for i := 5; i < 16; i++ {
		put(i)
		get(8, i)
	}

	put(0)
	q.close()
	if err := q.put(entries[1], false); err != ErrAsyncWriterClosed {
		t.Fatalf("put on a closed queue = %v, want %v", err, ErrAsyncWriterClosed)
	}
	dst := make([]*Entry, 8)
	if n, done := q.get(dst, true); n != 1 || !done || dst[0] != entries[0] {
		t.Fatalf("get after close = (%d, %v), want the last entry and done", n, done)
	}
	if n, done := q.get(dst, true); n != 0 || !done {
		t.Fatalf("get on a drained closed queue = (%d, %v), want (0, true)", n, done)
	}
	for i := range q.buf {
		if q.buf[i] != nil {
			t.Fatalf("drained queue still holds an entry in slot %d", i)
		}
	}
}
