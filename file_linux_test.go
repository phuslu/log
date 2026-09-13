//go:build linux

package log

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// iovecs builds an iovec list over the given byte slices, mirroring what
// writever hands to FileWriter.
func iovecs(bss ...[]byte) []syscall.Iovec {
	iovs := make([]syscall.Iovec, len(bss))
	for i := range bss {
		if len(bss[i]) > 0 {
			iovs[i].Base = &bss[i][0]
		}
		iovs[i].SetLen(len(bss[i]))
	}
	return iovs
}

func iovecsLenOf(iovs []syscall.Iovec) []int {
	lens := make([]int, len(iovs))
	for i := range iovs {
		lens[i] = int(iovs[i].Len)
	}
	return lens
}

// writevScript replays a scripted sequence of writev results so the iovec
// consumption logic can be tested without a real short write.
type writevScript struct {
	steps []struct {
		n   uintptr
		err error
	}
	calls int
	reqs  []int
}

func (s *writevScript) write(fd int, iovs []syscall.Iovec) (uintptr, error) {
	total := 0
	for i := range iovs {
		total += int(iovs[i].Len)
	}
	s.reqs = append(s.reqs, total)
	if s.calls >= len(s.steps) {
		return 0, errors.New("writevScript: unexpected writev call")
	}
	step := s.steps[s.calls]
	s.calls++
	return step.n, step.err
}

func TestWritevFullPartialConsumption(t *testing.T) {
	tests := []struct {
		name  string
		lens  []int
		steps []struct {
			n   uintptr
			err error
		}
		wantWritten   uintptr
		wantRemaining []int
		wantReqs      []int
		wantErr       error
	}{
		{
			name: "single iovec partial then complete",
			lens: []int{5},
			steps: []struct {
				n   uintptr
				err error
			}{{3, nil}, {2, nil}},
			wantWritten: 5,
			wantReqs:    []int{5, 2},
		},
		{
			name: "one syscall spanning multiple iovecs",
			lens: []int{3, 4, 5},
			steps: []struct {
				n   uintptr
				err error
			}{{5, nil}, {7, nil}},
			wantWritten: 12,
			wantReqs:    []int{12, 7},
		},
		{
			name: "exact iovec boundary",
			lens: []int{3, 4},
			steps: []struct {
				n   uintptr
				err error
			}{{3, nil}, {4, nil}},
			wantWritten: 7,
			wantReqs:    []int{7, 4},
		},
		{
			name: "partial then error",
			lens: []int{3, 4},
			steps: []struct {
				n   uintptr
				err error
			}{{3, nil}, {0, syscall.ENOSPC}},
			wantWritten:   3,
			wantRemaining: []int{4},
			wantReqs:      []int{7, 4},
			wantErr:       syscall.ENOSPC,
		},
		{
			name:        "no iovecs",
			lens:        nil,
			wantWritten: 0,
		},
		{
			name:        "all iovecs empty",
			lens:        []int{0, 0, 0},
			wantWritten: 0,
		},
		{
			name: "leading empty iovec",
			lens: []int{0, 4},
			steps: []struct {
				n   uintptr
				err error
			}{{4, nil}},
			wantWritten: 4,
			wantReqs:    []int{4},
		},
		{
			name: "zero progress",
			lens: []int{5},
			steps: []struct {
				n   uintptr
				err error
			}{{0, nil}},
			wantWritten:   0,
			wantRemaining: []int{5},
			wantReqs:      []int{5},
			wantErr:       io.ErrShortWrite,
		},
		{
			name: "minus one means aborted",
			lens: []int{5},
			steps: []struct {
				n   uintptr
				err error
			}{{^uintptr(0), nil}},
			wantWritten:   0,
			wantRemaining: []int{5},
			wantReqs:      []int{5},
			wantErr:       io.ErrShortWrite,
		},
		{
			name: "illegal byte count",
			lens: []int{5},
			steps: []struct {
				n   uintptr
				err error
			}{{9, nil}},
			wantWritten:   0,
			wantRemaining: []int{5},
			wantReqs:      []int{5},
			wantErr:       errWritevOverflow,
		},
		{
			name: "EINTR is retried",
			lens: []int{5},
			steps: []struct {
				n   uintptr
				err error
			}{{0, syscall.EINTR}, {5, nil}},
			wantWritten: 5,
			wantReqs:    []int{5, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := make([][]byte, len(tt.lens))
			for i, l := range tt.lens {
				bss[i] = bytes.Repeat([]byte("a"), l)
			}
			iovs := iovecs(bss...)

			script := &writevScript{steps: tt.steps}
			written, remaining, err := writevFullWith(1, iovs, script.write)

			if written != tt.wantWritten {
				t.Errorf("written = %d, want %d", written, tt.wantWritten)
			}
			if got := iovecsLenOf(remaining); !equalInts(got, tt.wantRemaining) {
				t.Errorf("remaining lens = %v, want %v", got, tt.wantRemaining)
			}
			if !equalInts(script.reqs, tt.wantReqs) {
				t.Errorf("syscall request sizes = %v, want %v", script.reqs, tt.wantReqs)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
			if script.calls != len(tt.steps) {
				t.Errorf("writev called %d times, want %d", script.calls, len(tt.steps))
			}
		})
	}
}

func TestWritevFullAdvancesBaseInPlace(t *testing.T) {
	buf := []byte("0123456789")
	iovs := iovecs(buf)

	script := &writevScript{steps: []struct {
		n   uintptr
		err error
	}{{4, nil}, {0, syscall.ENOSPC}}}
	written, remaining, err := writevFullWith(1, iovs, script.write)

	if written != 4 {
		t.Fatalf("written = %d, want 4", written)
	}
	if !errors.Is(err, syscall.ENOSPC) {
		t.Fatalf("err = %v, want ENOSPC", err)
	}
	if len(remaining) != 1 || remaining[0].Len != 6 {
		t.Fatalf("remaining = %v, want one iovec of 6 bytes", iovecsLenOf(remaining))
	}
	if remaining[0].Base != &buf[4] {
		t.Fatalf("remaining base %p, want %p", remaining[0].Base, &buf[4])
	}
}

func TestConsumeIovecs(t *testing.T) {
	tests := []struct {
		lens []int
		n    uintptr
		want []int
	}{
		{[]int{3, 4}, 0, []int{3, 4}},
		{[]int{3, 4}, 2, []int{1, 4}},
		{[]int{3, 4}, 3, []int{4}},
		{[]int{3, 4}, 5, []int{2}},
		{[]int{3, 4}, 7, nil},
		{[]int{0, 0, 4}, 0, []int{4}},
		{[]int{0, 4}, 4, nil},
		{[]int{5}, 10, nil},
	}
	for _, tt := range tests {
		bss := make([][]byte, len(tt.lens))
		for i, l := range tt.lens {
			bss[i] = bytes.Repeat([]byte("x"), l)
		}
		got := iovecsLenOf(consumeIovecs(iovecs(bss...), tt.n))
		if !equalInts(got, tt.want) {
			t.Errorf("consumeIovecs(%v, %d) = %v, want %v", tt.lens, tt.n, got, tt.want)
		}
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// setFileSizeLimit caps the size of newly written files so a real short write
// can be produced, and restores the limit when the test ends.
func setFileSizeLimit(t *testing.T, limit uint64) {
	t.Helper()
	signal.Ignore(syscall.SIGXFSZ)
	var old syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &old); err != nil {
		t.Skipf("getrlimit(RLIMIT_FSIZE): %v", err)
	}
	lim := old
	lim.Cur = limit
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &lim); err != nil {
		t.Skipf("setrlimit(RLIMIT_FSIZE): %v", err)
	}
	t.Cleanup(func() {
		_ = syscall.Setrlimit(syscall.RLIMIT_FSIZE, &old)
		signal.Reset(syscall.SIGXFSZ)
	})
}

// tempLogDir returns a temporary directory whose removal tolerates the
// background goroutine that FileWriter.rotate spawns to refresh the symlink.
func tempLogDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "log-writev-*")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}
	t.Cleanup(func() {
		for i := 0; i < 20; i++ {
			if err := os.RemoveAll(dir); err == nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Logf("giving up removing %s", dir)
	})
	return dir
}

// largestLog returns the content of the biggest rotated log file, since a
// rotation may leave an empty sibling behind if the timestamp ticks over.
func largestLog(t *testing.T, dir, base string) []byte {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, base+".*.log"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	var best []byte
	for _, name := range matches {
		data, err := os.ReadFile(name)
		if err != nil {
			continue
		}
		if len(data) > len(best) {
			best = data
		}
	}
	return best
}

func TestFileWriterWriteVPartialCountsSize(t *testing.T) {
	fw := &FileWriter{Filename: filepath.Join(tempLogDir(t), "out.log")}
	defer fw.Close()

	restore := writevFunc
	writevFunc = func(fd int, iovs []syscall.Iovec) (uintptr, error) {
		return 2, syscall.ENOSPC
	}
	defer func() { writevFunc = restore }()

	n, err := fw.WriteV(iovecs([]byte("hello")))
	if !errors.Is(err, syscall.ENOSPC) {
		t.Fatalf("err = %v, want ENOSPC", err)
	}
	if n != 2 {
		t.Fatalf("n = %d, want 2", n)
	}
	if fw.size != 2 {
		t.Fatalf("size = %d, want 2 (partial bytes must count)", fw.size)
	}

	// an empty batch must not panic nor touch the file
	if n, err := fw.WriteV(nil); n != 0 || err != nil {
		t.Fatalf("WriteV(nil) = (%d, %v), want (0, nil)", n, err)
	}
}

func TestFileWriterWriteVRealShortWrite(t *testing.T) {
	setFileSizeLimit(t, 24)

	dir := tempLogDir(t)
	fw := &FileWriter{Filename: filepath.Join(dir, "out.log")}
	defer fw.Close()

	n, err := fw.WriteV(iovecs(bytes.Repeat([]byte("a"), 64)))
	if n != 24 {
		t.Fatalf("n = %d, want 24", n)
	}
	if !errors.Is(err, syscall.EFBIG) {
		t.Fatalf("err = %v, want EFBIG", err)
	}
	if fw.size != 24 {
		t.Fatalf("size = %d, want 24", fw.size)
	}
	if got := largestLog(t, dir, "out"); len(got) != 24 {
		t.Fatalf("file holds %d bytes, want 24", len(got))
	}
}

// shortWriteFunc writes at most budget bytes in total and then fails with
// ENOSPC, while really writing the bytes it reports so file content stays
// consistent with the returned progress.
func shortWriteFunc(budget int) func(int, []syscall.Iovec) (uintptr, error) {
	var tmp [8]syscall.Iovec
	return func(fd int, iovs []syscall.Iovec) (uintptr, error) {
		if budget <= 0 {
			return 0, syscall.ENOSPC
		}
		left, k := budget, 0
		for i := range iovs {
			if left <= 0 || k == len(tmp) {
				break
			}
			l := int(iovs[i].Len)
			if l > left {
				l = left
			}
			if l == 0 {
				break
			}
			tmp[k] = iovs[i]
			tmp[k].SetLen(l)
			left -= l
			k++
		}
		if k == 0 {
			return 0, syscall.ENOSPC
		}
		n, err := writev(fd, tmp[:k])
		budget -= int(n)
		return n, err
	}
}

func TestAsyncWriterWriteNil(t *testing.T) {
	// writev fast path
	fw := &FileWriter{Filename: filepath.Join(tempLogDir(t), "nil.log")}
	w1 := &AsyncWriter{ChannelSize: 4, Writer: fw}
	if n, err := w1.Write(nil); n != 0 || err != nil {
		t.Fatalf("Write(nil) = (%d, %v), want (0, nil)", n, err)
	}
	if err := w1.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// generic writer path
	w2 := &AsyncWriter{ChannelSize: 4, DisableWritev: true, Writer: IOWriter{io.Discard}}
	if n, err := w2.Write(nil); n != 0 || err != nil {
		t.Fatalf("Write(nil) = (%d, %v), want (0, nil)", n, err)
	}
	if err := w2.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestAsyncWriterPartialWriteKeepsProgress(t *testing.T) {
	dir := tempLogDir(t)
	fw := &FileWriter{Filename: filepath.Join(dir, "out.log")}

	restore := writevFunc
	writevFunc = shortWriteFunc(10)
	defer func() { writevFunc = restore }()

	w := &AsyncWriter{ChannelSize: 0, Writer: fw}
	if _, err := w.Write(bytes.Repeat([]byte("a"), 25)); err != nil {
		t.Fatalf("write: %v", err)
	}

	err := w.Close()
	if !errors.Is(err, syscall.ENOSPC) {
		t.Fatalf("Close() = %v, want ENOSPC", err)
	}
	if got := largestLog(t, dir, "out"); len(got) != 10 {
		t.Fatalf("file holds %d bytes, want exactly the 10 bytes reported as written", len(got))
	}
}

// A retryable error that goes away must be retried, not latched: Close stays
// clean and the entry is still written.
func TestAsyncWriterRetriesTransientWritevError(t *testing.T) {
	dir := tempLogDir(t)
	fw := &FileWriter{Filename: filepath.Join(dir, "out.log")}

	restore := writevFunc
	calls := 0
	writevFunc = func(fd int, iovs []syscall.Iovec) (uintptr, error) {
		calls++
		if calls == 1 {
			return 0, syscall.EAGAIN
		}
		return writev(fd, iovs)
	}
	defer func() { writevFunc = restore }()

	w := &AsyncWriter{ChannelSize: 0, Writer: fw}
	if _, err := w.Write([]byte("hello\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil: a recovered EAGAIN must not be latched", err)
	}
	if got := string(largestLog(t, dir, "out")); got != "hello\n" {
		t.Fatalf("log = %q, want %q", got, "hello\n")
	}
	if calls != 2 {
		t.Fatalf("writev called %d times, want 2 (one EAGAIN plus one success)", calls)
	}
}

// A retryable error that never goes away must be retried a bounded number of
// times, latched, and reported by Close.
func TestAsyncWriterGivesUpAfterRetriesExhausted(t *testing.T) {
	dir := tempLogDir(t)
	fw := &FileWriter{Filename: filepath.Join(dir, "out.log")}

	restore := writevFunc
	calls := 0
	writevFunc = func(fd int, iovs []syscall.Iovec) (uintptr, error) {
		calls++
		return 0, syscall.EAGAIN
	}
	defer func() { writevFunc = restore }()

	w := &AsyncWriter{ChannelSize: 4, Writer: fw}
	if _, err := w.Write([]byte("hello\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); !errors.Is(err, syscall.EAGAIN) {
		t.Fatalf("Close() = %v, want EAGAIN", err)
	}
	if calls != maxWritevRetries+1 {
		t.Fatalf("writev called %d times, want %d (one attempt plus %d retries)", calls, maxWritevRetries+1, maxWritevRetries)
	}
	if got := largestLog(t, dir, "out"); len(got) != 0 {
		t.Fatalf("log holds %d bytes, want none", len(got))
	}
}

// Once the writer aborted, entries queued afterwards must be discarded
// without another writev attempt, and the latched error is still reported.
func TestAsyncWriterAbortedDropsLaterEntries(t *testing.T) {
	dir := tempLogDir(t)
	fw := &FileWriter{Filename: filepath.Join(dir, "out.log")}

	restore := writevFunc
	var calls int32
	writevFunc = func(fd int, iovs []syscall.Iovec) (uintptr, error) {
		atomic.AddInt32(&calls, 1)
		return 0, syscall.ENOBUFS
	}
	defer func() { writevFunc = restore }()

	w := &AsyncWriter{ChannelSize: 64, Writer: fw}
	if _, err := w.Write([]byte("first\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	// wait for the final allowed attempt, after which the writer aborts
	deadline := time.Now().Add(5 * time.Second)
	for atomic.LoadInt32(&calls) <= maxWritevRetries && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := atomic.LoadInt32(&calls); got != maxWritevRetries+1 {
		t.Fatalf("writev called %d times, want %d", got, maxWritevRetries+1)
	}

	// these arrive after the abort and must be dropped
	for i := 0; i < 3; i++ {
		if _, err := w.Write([]byte("later\n")); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := w.Close(); !errors.Is(err, syscall.ENOBUFS) {
		t.Fatalf("Close() = %v, want ENOBUFS", err)
	}
	if got := atomic.LoadInt32(&calls); got != maxWritevRetries+1 {
		t.Fatalf("writev called %d times, want no attempt for dropped entries", got)
	}
	if got := largestLog(t, dir, "out"); len(got) != 0 {
		t.Fatalf("log holds %d bytes, want none", len(got))
	}
}

type failingWriter struct {
	werr, cerr error
	calls      int
}

func (w *failingWriter) WriteEntry(e *Entry) (int, error) {
	w.calls++
	if w.calls == 1 {
		return 0, w.werr
	}
	return len(e.buf), nil
}

func (w *failingWriter) Close() error { return w.cerr }

func TestAsyncWriterCloseReturnsFirstError(t *testing.T) {
	werr := errors.New("write boom")
	cerr := errors.New("close boom")
	w := &AsyncWriter{
		ChannelSize:   0,
		DisableWritev: true,
		Writer:        &failingWriter{werr: werr, cerr: cerr},
	}
	if _, err := w.Write([]byte("first")); err != nil {
		t.Fatalf("first write: %v", err)
	}
	// this write would succeed, it must not clear the latched error
	if _, err := w.Write([]byte("second")); err != nil {
		t.Fatalf("second write: %v", err)
	}
	if err := w.Close(); !errors.Is(err, werr) {
		t.Fatalf("Close() = %v, want the first write error %v", err, werr)
	}
}

type gatedWriter struct {
	gate chan struct{}
	seen chan string
}

func (w *gatedWriter) WriteEntry(e *Entry) (int, error) {
	<-w.gate
	w.seen <- string(e.buf)
	return len(e.buf), nil
}

func TestAsyncWriterWriteCopiesPayload(t *testing.T) {
	bw := &gatedWriter{gate: make(chan struct{}), seen: make(chan string, 1)}
	w := &AsyncWriter{ChannelSize: 4, DisableWritev: true, Writer: bw}

	buf := []byte("AAAA")
	if _, err := w.Write(buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	buf[0] = 'Z' // caller is allowed to reuse its buffer after Write returns
	close(bw.gate)

	if got := <-bw.seen; got == "ZAAA" {
		t.Fatalf("writer observed the caller's mutated buffer %q", got)
	} else if got != "AAAA" {
		t.Fatalf("writer observed %q, want AAAA", got)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestAsyncWriterWritevIntegrity(t *testing.T) {
	const lines = 2500

	writevDir := tempLogDir(t)
	ioDir := tempLogDir(t)
	w1 := &AsyncWriter{
		ChannelSize: 64,
		Writer:      &FileWriter{Filename: filepath.Join(writevDir, "out.log")},
	}
	w2 := &AsyncWriter{
		ChannelSize:   64,
		DisableWritev: true,
		Writer:        &FileWriter{Filename: filepath.Join(ioDir, "out.log")},
	}

	// reuse one buffer on purpose: AsyncWriter must copy it
	buf := make([]byte, 0, 32)
	for i := 0; i < lines; i++ {
		buf = append(buf[:0], fmt.Sprintf("line-%06d\n", i)...)
		if _, err := w1.Write(buf); err != nil {
			t.Fatalf("writev write: %v", err)
		}
		if _, err := w2.Write(buf); err != nil {
			t.Fatalf("plain write: %v", err)
		}
	}
	if err := w1.Close(); err != nil {
		t.Fatalf("writev close: %v", err)
	}
	if err := w2.Close(); err != nil {
		t.Fatalf("plain close: %v", err)
	}

	d1 := largestLog(t, writevDir, "out")
	d2 := largestLog(t, ioDir, "out")
	if !bytes.Equal(d1, d2) {
		t.Fatalf("writev and plain write differ: %d vs %d bytes", len(d1), len(d2))
	}

	got := strings.Split(strings.TrimSuffix(string(d1), "\n"), "\n")
	if len(got) != lines {
		t.Fatalf("got %d lines, want %d", len(got), lines)
	}
	for i, line := range got {
		if want := fmt.Sprintf("line-%06d", i); line != want {
			t.Fatalf("line %d = %q, want %q", i, line, want)
		}
	}
}

func TestAsyncWriterConcurrentProducers(t *testing.T) {
	const (
		producers = 4
		per       = 500
	)
	dir := tempLogDir(t)
	w := &AsyncWriter{
		ChannelSize: 128,
		Writer:      &FileWriter{Filename: filepath.Join(dir, "out.log")},
	}

	var wg sync.WaitGroup
	for p := 0; p < producers; p++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			for i := 0; i < per; i++ {
				if _, err := w.Write([]byte(fmt.Sprintf("p%d-%06d\n", p, i))); err != nil {
					t.Errorf("producer %d: %v", p, err)
					return
				}
			}
		}(p)
	}
	wg.Wait()

	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	seen := make(map[string]bool, producers*per)
	for _, line := range strings.Split(strings.TrimSuffix(string(largestLog(t, dir, "out")), "\n"), "\n") {
		if line == "" {
			continue
		}
		var p, i int
		if _, err := fmt.Sscanf(line, "p%d-%06d", &p, &i); err != nil {
			t.Fatalf("corrupt line %q", line)
		}
		if seen[line] {
			t.Fatalf("duplicate line %q", line)
		}
		seen[line] = true
	}
	if len(seen) != producers*per {
		t.Fatalf("got %d lines, want %d", len(seen), producers*per)
	}
}
