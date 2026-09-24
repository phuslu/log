//go:build linux

package log

import (
	"net"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// journalctl -o verbose -f
func TestJournalWriter(t *testing.T) {
	w := &JournalWriter{}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, _ = wlprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello journal writer"}`+"\n", level)
	}

	_, _ = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"error","msg":"a test message\n"}`+"\n")
	_, _ = wlprintf(w, InfoLevel, "a long long long long message.\n")
	w.Close()
}

func TestJournalWriterError(t *testing.T) {
	sockname := filepath.Join(t.TempDir(), "null.sock")

	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: sockname, Net: "unixgram"})
	if err != nil {
		t.Errorf("listen error: %+v", err)
		return
	}

	// Drain the socket while the writers below run. The goroutine must not call
	// t.Logf and must be gone before the test returns: logging from a goroutine
	// after its test has completed panics the whole test binary.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var data [512]byte
		for {
			buf := data[:]
			if _, _, err := conn.ReadFromUnix(buf); err != nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	defer func() {
		conn.Close() // unblocks ReadFromUnix
		wg.Wait()
	}()

	w := &JournalWriter{
		JournalSocket: sockname,
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, _ = wlprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello journal writer"}`+"\n", level)
	}

	_, _ = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"error","msg":"a test message\n"}`+"\n")
	_, _ = wlprintf(w, InfoLevel, "a long long long long message.\n")
	w.Close()
}
