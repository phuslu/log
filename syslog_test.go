package log

import (
	"net"
	"os"
	"testing"
	"time"
)

func TestSyslogWriterTCP(t *testing.T) {
	w := &SyslogWriter{
		Network: "udp",
		Address: "10.0.0.2:543",
		Tag:     "",
		Dial:    net.Dial,
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		wlprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello journal writer"}`+"\n", level)
		wlprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277+08:00","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello journal writer"}`+"\n", level)
		wlprintf(w, ParseLevel(level), `{"ts":1234567890,"level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello journal writer"}`+"\n", level)
	}

	_, err := wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"error","msg":"a test message\n"}`+"\n")
	if err != nil {
		t.Errorf("write syslog writer error: %+v", err)
	}

	_, err = wlprintf(w, InfoLevel, "a long long long long message.\n")
	if err != nil {
		t.Errorf("write syslog writer error: %+v", err)
	}

	w.Close()
	w.Close()

	_, err = wlprintf(w, InfoLevel, "a long long long long message again.\n")
	if err != nil {
		t.Errorf("write syslog writer error: %+v", err)
	}
}

func TestSyslogWriterTCPError(t *testing.T) {
	w := &SyslogWriter{
		Network: "tcp",
		Address: "127.0.0.1:601",
		Tag:     "",
		Dial:    net.Dial,
	}

	wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"error","msg":"a test message\n"}`+"\n")
	w.Close()
	wlprintf(w, InfoLevel, "a long long long long message again.\n")
}

func TestSyslogWriterUnix(t *testing.T) {
	const sockname = "/tmp/go-tmp-null.sock"

	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: sockname, Net: "unixgram"})
	if err != nil {
		t.Errorf("listen error: %+v", err)
		return
	}
	defer os.Remove(sockname)

	go func() {
		var data [512]byte
		for {
			buf := data[:]
			n, uaddr, err := conn.ReadFromUnix(buf)
			if err != nil {
				t.Logf("listen: error: %v\n", err)
			} else {
				t.Logf("listen: received %v bytes from %+v\n", n, uaddr)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	w := &SyslogWriter{
		Network: "unixgram",
		Address: "/tmp/go-tmp-null.sock",
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		wlprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello journal writer"}`+"\n", level)
		wlprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277+08:00","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello journal writer"}`+"\n", level)
		wlprintf(w, ParseLevel(level), `{"ts":1234567890,"level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello journal writer"}`+"\n", level)
	}

	_, err = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"error","msg":"a test message\n"}`+"\n")
	if err != nil {
		t.Errorf("write syslog writer error: %+v", err)
	}

	_, err = wlprintf(w, InfoLevel, "a long long long long message.\n")
	if err != nil {
		t.Errorf("write syslog writer error: %+v", err)
	}

	w.Close()
	w.Close()

	_, err = wlprintf(w, InfoLevel, "a long long long long message again.\n")
	if err != nil {
		t.Errorf("write syslog writer error: %+v", err)
	}

	os.Remove(sockname)

	_, err = wlprintf(w, InfoLevel, "a long long long long message again.\n")
	t.Logf("write syslog writer error: %+v", err)
}
