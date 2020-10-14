// +build linux

package log

import (
	"net"
	"os"
	"testing"
	"time"
)

// journalctl -o verbose -f
func TestJournalWriter(t *testing.T) {
	w := &JournalWriter{}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		wprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello journal writer"}`+"\n", level)
	}

	wprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"error","msg":"a test message\n"}`+"\n")
	wprintf(w, InfoLevel, "a long long long long message.\n")
	w.Close()
}

func TestJournalWriterError(t *testing.T) {
	const sockname = "/tmp/go-tmp-null.sock"

	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: sockname, Net: "unixgram"})
	if err != nil {
		t.Errorf("listen error: %+v", err)
		return
	}
	defer os.Remove(sockname)

	go func() {
		for {
			buf := make([]byte, 2048)
			n, uaddr, err := conn.ReadFromUnix(buf)
			if err != nil {
				t.Logf("listen: error: %v\n", err)
			} else {
				t.Logf("listen: received %v bytes from %+v\n", n, uaddr)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	w := &JournalWriter{
		JournalSocket: sockname,
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		wprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello journal writer"}`+"\n", level)
	}

	wprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"error","msg":"a test message\n"}`+"\n")
	wprintf(w, InfoLevel, "a long long long long message.\n")
	w.Close()
}
