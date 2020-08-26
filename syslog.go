package log

import (
	"bytes"
	"context"
	"net"
	"sync"
	"time"
)

// SyslogWriter is an io.WriteCloser that writes logs to a syslog server..
type SyslogWriter struct {
	Network  string
	Address  string
	Hostname string
	Tag      string

	// Dial specifies the dial function for creating TCP/TLS connections.
	Dial func(ctx context.Context, network, addr string) (net.Conn, error)

	mu   sync.Mutex
	conn net.Conn
}

// Close closes a connection to the syslog daemon.
func (w *SyslogWriter) Close() (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		err = w.conn.Close()
		w.conn = nil
		return
	}
	return
}

// connect makes a connection to the syslog server.
// It must be called with w.mu held.
func (w *SyslogWriter) connect() (err error) {
	if w.conn != nil {
		w.conn.Close()
		w.conn = nil
	}

	var dial = w.Dial
	if dial != nil {
		dial = (&net.Dialer{}).DialContext
	}

	if w.Network == "" {
		for _, network := range []string{"unixgram", "unix"} {
			for _, path := range []string{"/dev/log", "/var/run/syslog", "/var/run/log"} {
				w.conn, err = dial(context.Background(), network, path)
				if err == nil {
					break
				}
			}
		}
		if w.Hostname == "" {
			w.Hostname = hostname
		}
	} else {
		w.conn, err = dial(context.Background(), w.Network, w.Address)
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

	var level byte
	// guess level by fixed offset
	lp := len(p)
	if lp > 49 {
		_ = p[49]
		switch {
		case p[32] == 'Z' && p[42] == ':' && p[43] == '"':
			level = p[44]
		case p[32] == '+' && p[47] == ':' && p[48] == '"':
			level = p[49]
		}
	}
	// guess level by "level":" beginning
	if level == 0 {
		if i := bytes.Index(p, levelBegin); i > 0 && i+len(levelBegin)+1 < lp {
			level = p[i+len(levelBegin)]
		}
	}

	// convert level to syslog priority
	var priority byte
	switch level {
	case 't':
		priority = '7' // LOG_DEBUG
	case 'd':
		priority = '7' // LOG_DEBUG
	case 'i':
		priority = '6' // LOG_INFO
	case 'w':
		priority = '4' // LOG_WARNING
	case 'e':
		priority = '3' // LOG_ERR
	case 'f':
		priority = '2' // LOG_CRIT
	case 'p':
		priority = '1' // LOG_ALERT
	default:
		priority = '6' // LOG_INFO
	}

	b := b1kpool.Get().([]byte)
	defer b1kpool.Put(b)

	b = append(b[:0], '<', priority, '>')
	if w.Network == "" {
		// Compared to the network form below, the changes are:
		//	1. Use time.Stamp instead of time.RFC3339.
		//	2. Drop the hostname field.
		b = timeNow().AppendFormat(b, time.Stamp)
	} else {
		b = timeNow().AppendFormat(b, time.RFC3339)
		b = append(b, ' ')
		b = append(b, w.Hostname...)
	}
	b = append(b, ' ')
	b = append(b, w.Tag...)
	b = append(b, '[')
	b = append(b, pid...)
	b = append(b, ']', ':', ' ')
	b = append(b, p...)

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		if n, err := w.conn.Write(b); err == nil {
			return n, err
		}
	}
	if err := w.connect(); err != nil {
		return 0, err
	}
	return w.conn.Write(b)
}
