package log

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestAsyncWriterSize(t *testing.T) {
	w := &AsyncWriter{
		BufferSize:   1000,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       os.Stderr,
	}
	for i := 0; i < 100; i++ {
		fmt.Fprintf(w, "%s, %d during buffer writer 1k buff size\n", timeNow(), i)
	}
	time.Sleep(time.Second)
}

func TestAsyncWriterSizeZero(t *testing.T) {
	w := &AsyncWriter{
		BufferSize: 0,
		Writer:     os.Stderr,
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
