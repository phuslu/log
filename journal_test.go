// +build linux

package log

import (
	"fmt"
	"testing"
)

// journalctl -o verbose -f
func TestJournalWriter(t *testing.T) {
	w := &JournalWriter{}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json journal writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json journal writer error: %+v", err)
		}
	}
}
