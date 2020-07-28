package log

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestBufferWriter(t *testing.T) {
	w := &BufferWriter{
		Size:     8192,
		Duration: 1000 * time.Millisecond,
		Writer:   &Writer{},
	}
	fmt.Fprintf(w, "%s, before buffer writer auto flush\n", timeNow())
	time.Sleep(1100 * time.Millisecond)
	fmt.Fprintf(os.Stderr, "%s, after buffer writer auto flush\n", timeNow())
}

func TestBufferWriterFlush(t *testing.T) {
	w := &BufferWriter{
		Size:     8192,
		Duration: 1000 * time.Millisecond,
		Writer:   &Writer{},
	}
	fmt.Fprintf(w, "%s, before buffer writer flush\n", timeNow())
	w.Flush()
	fmt.Fprintf(os.Stderr, "%s, after buffer writer flush\n", timeNow())
}
