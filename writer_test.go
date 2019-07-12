package log

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestWriter(t *testing.T) {
	w := &Writer{}
	fmt.Fprintf(w, "hello writer!\n")
	w.Rotate()
}

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

func TestConsoleWriter(t *testing.T) {
	w := &ConsoleWriter{
		ANSIColor: runtime.GOOS != "windows",
	}

	for _, level := range []string{"debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json console writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json console writer error: %+v", err)
		}
	}
}
