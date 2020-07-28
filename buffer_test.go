package log

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestBufferWriterSize(t *testing.T) {
	w := &BufferWriter{
		MaxSize:       1000,
		FlushDuration: 1000 * time.Millisecond,
		Out:           os.Stderr,
	}
	for i := 0; i < 100; i++ {
		fmt.Fprintf(w, "%s, %d during buffer writer 1k buff size\n", timeNow(), i)
	}
	time.Sleep(time.Second)
}

func TestBufferWriterSizeZero(t *testing.T) {
	w := &BufferWriter{
		MaxSize: 0,
		Out:     os.Stderr,
	}
	fmt.Fprintf(w, "%s, before buffer writer zero size\n", timeNow())
	time.Sleep(1100 * time.Millisecond)
	fmt.Fprintf(os.Stderr, "%s, after buffer writer zero size\n", timeNow())
}

func TestBufferWriterDuration(t *testing.T) {
	w := &BufferWriter{
		MaxSize:       1000,
		FlushDuration: 10 * time.Millisecond,
		Out:           os.Stderr,
	}
	fmt.Fprintf(w, "%s, during buffer writer tiny duration\n", timeNow())
	time.Sleep(200 * time.Millisecond)
}

func TestBufferWriterFlushAuto(t *testing.T) {
	w := &BufferWriter{
		MaxSize:       8192,
		FlushDuration: 1000 * time.Millisecond,
		Out:           os.Stderr,
	}
	fmt.Fprintf(w, "%s, before buffer writer auto flush\n", timeNow())
	time.Sleep(1100 * time.Millisecond)
	fmt.Fprintf(os.Stderr, "%s, after buffer writer auto flush\n", timeNow())
}

func TestBufferWriterFlushCall(t *testing.T) {
	w := &BufferWriter{
		MaxSize:       8192,
		FlushDuration: 1000 * time.Millisecond,
		Out:           os.Stderr,
	}
	fmt.Fprintf(w, "%s, before buffer writer flush\n", timeNow())
	w.Flush()
	fmt.Fprintf(os.Stderr, "%s, after buffer writer flush\n", timeNow())
}

func TestBufferWriterClose(t *testing.T) {
	w := &BufferWriter{
		MaxSize:       8192,
		FlushDuration: 1000 * time.Millisecond,
		Out:           os.Stderr,
	}
	fmt.Fprintf(w, "%s, before buffer writer flush\n", timeNow())
	w.Close()
	fmt.Fprintf(os.Stderr, "%s, after buffer writer flush\n", timeNow())
}
