package log

import (
	"net"
	"sync"
	"time"
)

// SyslogWriter is an io.WriteCloser that writes logs to journald.
type SyslogWriter struct {
	Network  string
	Address  string
	Hostname string
	Tag      string

	mu    sync.Mutex
	conn  net.Conn
	local bool
}

// Close closes a connection to the syslog daemon.
func (w *SyslogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		err := w.conn.Close()
		w.conn = nil
		return err
	}
	return nil
}

// connect makes a connection to the syslog server.
// It must be called with w.mu held.
func (w *SyslogWriter) connect() (err error) {
	if w.conn != nil {
		w.conn.Close()
		w.conn = nil
	}

	if w.Network == "" {
		for _, network := range []string{"unixgram", "unix"} {
			for _, path := range []string{"/dev/log", "/var/run/syslog", "/var/run/log"} {
				w.conn, err = net.Dial(network, path)
				if err == nil {
					break
				}
			}
		}
		if w.Hostname == "" {
			w.Hostname = hostname
		}
	} else {
		w.conn, err = net.Dial(w.Network, w.Address)
		if err == nil && w.Hostname == "" {
			w.Hostname = w.conn.LocalAddr().String()
		}
	}
	return
}

// Write implements io.Writer.
func (w *SyslogWriter) Write(p []byte) (n int, err error) {
	if w.conn == nil {
		w.mu.Lock()
		if w.conn == nil {
			err = w.connect()
			if err != nil {
				w.mu.Unlock()
				return
			}
		}
		w.mu.Unlock()
	}

	var t dot
	err = jsonToDot(p, &t)
	if err != nil {
		return
	}

	var pr byte
	switch t.Level {
	case TraceLevel:
		pr = '7' // LOG_DEBUG
	case DebugLevel:
		pr = '7' // LOG_DEBUG
	case InfoLevel:
		pr = '6' // LOG_INFO
	case WarnLevel:
		pr = '4' // LOG_WARNING
	case ErrorLevel:
		pr = '3' // LOG_ERR
	case FatalLevel:
		pr = '2' // LOG_CRIT
	case PanicLevel:
		pr = '1' // LOG_ALERT
	default:
		pr = '6' // LOG_INFO
	}

	b := bbpool.Get().(*bb)
	b.Reset()
	defer bbpool.Put(b)

	b.B = append(b.B, '<', pr, '>')
	if w.Network == "" {
		// Compared to the network form below, the changes are:
		//	1. Use time.Stamp instead of time.RFC3339.
		//	2. Drop the hostname field.
		b.B = timeNow().AppendFormat(b.B, time.Stamp)
	} else {
		b.B = timeNow().AppendFormat(b.B, time.RFC3339)
		b.B = append(b.B, ' ')
		b.B = append(b.B, w.Hostname...)
	}
	b.B = append(b.B, ' ')
	b.B = append(b.B, w.Tag...)
	b.B = append(b.B, '[')
	b.B = append(b.B, pid...)
	b.B = append(b.B, ']', ':', ' ')
	b.B = append(b.B, p...)

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		if n, err := w.conn.Write(b.B); err == nil {
			return n, err
		}
	}
	if err := w.connect(); err != nil {
		return 0, err
	}
	return w.conn.Write(b.B)
}
