//go:build windows

package log

import (
	"testing"
)

func TestEventlogWriter(t *testing.T) {
	w := MustNewEventlogWriter(".NET Runtime", 1000, "")

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := wlprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json eventlog writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json eventlog writer error: %+v", err)
		}
	}
}

func TestNewEventlogWriter(t *testing.T) {
	t.Run("test init err", func(t *testing.T) {
		_, err := NewEventlogWriter("", 0, "")
		if err == nil {
			t.Errorf("expect init err, got nil")
		}
	})
}
