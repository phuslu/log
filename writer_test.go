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

func TestJSONConsoleWriter(t *testing.T) {
	w := &JSONConsoleWriter{
		ANSIColor: runtime.GOOS != "windows",
	}
	_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"test.go:42","error":"i am a error","foo":"bar","n":42,"message":"hello json console writer"}`+"\n")
	if err != nil {
		t.Errorf("test json console writer error: %+v", err)
	}
}
