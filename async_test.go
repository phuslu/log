package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestAsyncWriterBatch(t *testing.T) {
	w := &AsyncWriter{
		BatchSize:    11,
		SyncDuration: 1100 * time.Millisecond,
		Writer:       os.Stderr,
	}
	for i := 0; i < 20; i++ {
		fmt.Fprintf(w, "%s, %d during async writer 1k buff size\n", timeNow(), i)
	}
	time.Sleep(time.Second)
	fmt.Fprintf(os.Stderr, "%s, sync to writer\n", timeNow())
	w.Sync()
}

func TestAsyncWriterNoBatch(t *testing.T) {
	w := &AsyncWriter{
		BatchSize:    0,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       os.Stderr,
	}
	fmt.Fprintf(w, "%s, before async writer zero size\n", timeNow())
	time.Sleep(1100 * time.Millisecond)
	w.Sync()
	fmt.Fprintf(os.Stderr, "%s, after async writer zero size\n", timeNow())
}

func TestAsyncWriterDuration(t *testing.T) {
	w := &AsyncWriter{
		BatchSize:    10,
		SyncDuration: 10 * time.Millisecond,
		Writer:       os.Stderr,
	}
	fmt.Fprintf(w, "%s, during async writer tiny duration\n", timeNow())
	time.Sleep(200 * time.Millisecond)
}

func TestAsyncWriterSyncAuto(t *testing.T) {
	w := &AsyncWriter{
		BatchSize:    10,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       os.Stderr,
	}
	fmt.Fprintf(w, "%s, before async writer auto sync\n", timeNow())
	time.Sleep(1100 * time.Millisecond)
	fmt.Fprintf(os.Stderr, "%s, after async writer auto sync\n", timeNow())
}

func TestAsyncWriterSyncCall(t *testing.T) {
	w := &AsyncWriter{
		BatchSize:    10,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       os.Stderr,
	}
	fmt.Fprintf(w, "%s, before async writer sync\n", timeNow())
	w.Sync()
	fmt.Fprintf(os.Stderr, "%s, after async writer sync\n", timeNow())
}

func TestAsyncWriterSyncer(t *testing.T) {
	w := &AsyncWriter{
		BatchSize:    10,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       os.Stderr,
	}
	fmt.Fprintf(w, "%s, before async writer sync\n", timeNow())

	w.Sync()
}

func TestAsyncWriterClose(t *testing.T) {
	w := &AsyncWriter{
		BatchSize:    10,
		SyncDuration: 1000 * time.Millisecond,
		Writer:       func() *os.File { f, _ := os.Open(os.DevNull); return f }(),
	}
	fmt.Fprintf(w, "%s, before async writer close\n", timeNow())
	w.Close()
	fmt.Fprintf(os.Stderr, "%s, after async writer close\n", timeNow())
}

func BenchmarkAsyncWriter(b *testing.B) {
	w := &AsyncWriter{
		Writer: ioutil.Discard,
	}

	b.SetParallelism(1000)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		p := []byte(`{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1], "b":{}},"message":"hello json writer"}`)
		for b.Next() {
			w.Write(p)
		}
	})
}
