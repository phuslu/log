// +build windows

package log

import (
	"testing"
)

func TestEventlogWriter(t *testing.T) {
	w := &EventlogWriter{
		Source: ".NET Runtime",
		ID:     1000,
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := wprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json eventlog writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json eventlog writer error: %+v", err)
		}
	}
}
