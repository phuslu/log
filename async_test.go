package log

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestAsyncWriterSmallSize(t *testing.T) {
	w := &AsyncWriter{
		BufferSize:   1000,
		SyncDuration: 1100 * time.Millisecond,
		Writer:       os.Stderr,
	}
	for i := 0; i < 20; i++ {
		fmt.Fprintf(w, "%s, %d during buffer writer 1k buff size\n", timeNow(), i)
	}
	time.Sleep(time.Second)
	fmt.Fprintf(os.Stderr, "%s, sync to writer\n", timeNow())
	w.Close()
}

func TestAsyncWriterZeroSize(t *testing.T) {
	w := &AsyncWriter{
		BufferSize:   0,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       os.Stderr,
	}
	fmt.Fprintf(w, "%s, before buffer writer zero size\n", timeNow())
	time.Sleep(1100 * time.Millisecond)
	fmt.Fprintf(os.Stderr, "%s, after buffer writer zero size\n", timeNow())
}

func TestAsyncWriterDuration(t *testing.T) {
	w := &AsyncWriter{
		BufferSize:   1000,
		SyncDuration: 10 * time.Millisecond,
		Writer:       os.Stderr,
	}
	fmt.Fprintf(w, "%s, during buffer writer tiny duration\n", timeNow())
	time.Sleep(200 * time.Millisecond)
}

func TestAsyncWriterSyncAuto(t *testing.T) {
	w := &AsyncWriter{
		BufferSize:   8192,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       os.Stderr,
	}
	fmt.Fprintf(w, "%s, before buffer writer auto sync\n", timeNow())
	time.Sleep(1100 * time.Millisecond)
	fmt.Fprintf(os.Stderr, "%s, after buffer writer auto sync\n", timeNow())
}

func TestAsyncWriterSyncCall(t *testing.T) {
	w := &AsyncWriter{
		BufferSize:   8192,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       os.Stderr,
	}
	fmt.Fprintf(w, "%s, before buffer writer sync\n", timeNow())
	w.Sync()
	fmt.Fprintf(os.Stderr, "%s, after buffer writer sync\n", timeNow())
}

func TestAsyncWriterSyncer(t *testing.T) {
	w := &AsyncWriter{
		BufferSize:   8192,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       os.Stderr,
	}
	fmt.Fprintf(w, "%s, before buffer writer sync\n", timeNow())

	w.Sync()
}

func TestAsyncWriterClose(t *testing.T) {
	w := &AsyncWriter{
		BufferSize:   8192,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       func() *os.File { f, _ := os.Open(os.DevNull); return f }(),
	}
	fmt.Fprintf(w, "%s, before buffer writer sync\n", timeNow())
	w.Close()
	fmt.Fprintf(os.Stderr, "%s, after buffer writer sync\n", timeNow())
}

func BenchmarkAsyncWriter(b *testing.B) {
	file, err := os.Open(os.DevNull)
	if err != nil {
		b.Errorf("open null device error: %+v", err)
	}

	w := &AsyncWriter{
		Writer: file,
	}

	p := []byte(`{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1], "b":{}},"message":"hello json console color writer"}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Write(p)
	}
}
