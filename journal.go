// +build linux

package log

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// JournalWriter is an io.WriteCloser that writes logs to journald.
type JournalWriter struct {
	// JournalSocket specifies socket name, using `/run/systemd/journal/socket` if empty.
	JournalSocket string

	// TimeField specifies an optional field name for time parsing in output.
	TimeField string

	once sync.Once
	addr *net.UnixAddr
	conn *net.UnixConn
}

// Close implements io.Closer.
func (w *JournalWriter) Close() (err error) {
	if w.conn != nil {
		err = w.conn.Close()
	}
	return
}

// Write implements io.Writer.
func (w *JournalWriter) Write(p []byte) (n int, err error) {
	w.once.Do(func() {
		// unix addr
		w.addr = &net.UnixAddr{
			Net:  "unixgram",
			Name: w.JournalSocket,
		}
		if w.addr.Name == "" {
			w.addr.Name = "/run/systemd/journal/socket"
		}
		// unix conn
		var autobind *net.UnixAddr
		autobind, err = net.ResolveUnixAddr("unixgram", "")
		if err != nil {
			return
		}
		w.conn, err = net.ListenUnixgram("unixgram", autobind)
	})

	if err != nil {
		return
	}

	var m map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		return
	}

	// buffer
	b := bbpool.Get().(*bb)
	b.Reset()
	defer bbpool.Put(b)

	print := func(w io.Writer, name, value string) {
		if strings.ContainsRune(value, '\n') {
			fmt.Fprintln(w, name)
			binary.Write(w, binary.LittleEndian, uint64(len(value)))
			fmt.Fprintln(w, value)
		} else {
			fmt.Fprintf(w, "%s=%s\n", name, value)
		}
	}

	// level
	if v, ok := m["level"]; ok {
		var priority string
		switch s, _ := v.(string); ParseLevel(s) {
		case TraceLevel:
			priority = "7" // Debug
		case DebugLevel:
			priority = "7" // Debug
		case InfoLevel:
			priority = "6" // Informational
		case WarnLevel:
			priority = "4" // Warning
		case ErrorLevel:
			priority = "3" // Error
		case FatalLevel:
			priority = "2" // Critical
		case PanicLevel:
			priority = "0" // Emergency
		default:
			priority = "5" // Notice
		}
		print(b, "PRIORITY", priority)
	}

	// time
	var timeField = w.TimeField
	if timeField == "" {
		timeField = "time"
	}

	// message
	var msgField = "message"
	if _, ok := m[msgField]; !ok {
		if _, ok := m["msg"]; ok {
			msgField = "msg"
		}
	}
	if v, ok := m[msgField]; ok {
		s, _ := v.(string)
		if s != "" && s[len(s)-1] == '\n' {
			s = s[:len(s)-1]
		}
		print(b, "MESSAGE", s)
	}

	// fields
	for _, k := range jsonKeys(p) {
		switch k {
		case "level", msgField, timeField:
			continue
		}
		v := m[k]
		s, ok := v.(string)
		if !ok {
			s = fmt.Sprint(v)
		}
		print(b, strings.ToUpper(k), s)
	}

	print(b, "JSON", *(*string)(unsafe.Pointer(&p)))

	// write
	n, _, err = w.conn.WriteMsgUnix(b.B, nil, w.addr)
	if err == nil {
		return
	}

	opErr, ok := err.(*net.OpError)
	if !ok || opErr == nil {
		return
	}

	sysErr, ok := opErr.Err.(*os.SyscallError)
	if !ok || sysErr == nil {
		return
	}
	if sysErr.Err != syscall.EMSGSIZE && sysErr.Err != syscall.ENOBUFS {
		return
	}

	// Large log entry, send it via tempfile and ancillary-fd.
	var file *os.File
	file, err = ioutil.TempFile("/dev/shm/", "journal.XXXXX")
	if err != nil {
		return
	}
	err = syscall.Unlink(file.Name())
	if err != nil {
		return
	}
	defer file.Close()
	n, err = file.Write(b.B)
	if err != nil {
		return
	}
	rights := syscall.UnixRights(int(file.Fd()))
	_, _, err = w.conn.WriteMsgUnix([]byte{}, rights, w.addr)
	if err == nil {
		n = len(p)
	}

	return
}
