package log

import (
	"fmt"
	"testing"
	"time"
)

func TestBufferWriter(t *testing.T) {
	w := &BufferWriter{
		BufferSize:    8192,
		FlushDuration: 100 * time.Millisecond,
		Writer:        &Writer{},
	}
	fmt.Fprintf(w, "hello buffio writer!\n")
	time.Sleep(110 * time.Millisecond)
}

func TestBufferWriterFlush(t *testing.T) {
	w := &BufferWriter{
		BufferSize:    8192,
		FlushDuration: 100 * time.Millisecond,
		Writer:        &Writer{},
	}
	fmt.Fprintf(w, "hello buffio flushed writer!\n")
	w.Flush()
}
