// +build windows

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"syscall"
)

var (
	muLeacy                 sync.Mutex
	SetConsoleTextAttribute = syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleTextAttribute").Call
)

func (w *ConsoleWriter) Write(p []byte) (int, error) {
	return w.leacyWrite(p)
}

func (w *ConsoleWriter) leacyWrite(p []byte) (n int, err error) {
	muLeacy.Lock()
	defer muLeacy.Unlock()

	const (
		winColorBlue   = 1
		winColorGreen  = 2
		winColorAqua   = 3
		winColorRed    = 4
		winColorPurple = 5
		winColorYellow = 6
		winColorWhite  = 7
		winColorGray   = 8
	)

	var m map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		n, err = os.Stderr.Write(p)
		return
	}

	var printf = func(color uintptr, format string, args ...interface{}) {
		if color != winColorWhite {
			SetConsoleTextAttribute(uintptr(syscall.Stderr), color)
		}
		var i int
		i, err = fmt.Fprintf(os.Stderr, format, args...)
		n += i
		if color != winColorWhite {
			SetConsoleTextAttribute(uintptr(syscall.Stderr), winColorWhite)
		}
	}

	if v, ok := m["time"]; ok {
		if w.ANSIColor {
			printf(winColorGray, "%s ", v)
		} else {
			printf(winColorWhite, "%s ", v)
		}
	}

	if v, ok := m["level"]; ok {
		var s string
		var c uintptr
		switch s, _ = v.(string); ParseLevel(s) {
		case DebugLevel:
			c, s = winColorYellow, "DBG"
		case InfoLevel:
			c, s = winColorGreen, "INF"
		case WarnLevel:
			c, s = winColorRed, "WRN"
		case ErrorLevel:
			c, s = winColorRed, "ERR"
		case FatalLevel:
			c, s = winColorRed, "FTL"
		case PanicLevel:
			c, s = winColorRed, "PNC"
		default:
			c, s = winColorRed, "???"
		}
		if w.ANSIColor {
			printf(c, "%s ", s)
		} else {
			printf(winColorWhite, "%s ", s)
		}
	}

	if v, ok := m["goid"]; ok {
		printf(winColorWhite, "%s ", v)
	}

	if v, ok := m["caller"]; ok {
		printf(winColorWhite, "%s ", v)
	}

	if v, ok := m["message"]; ok {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		if w.ANSIColor {
			printf(winColorAqua, ">")
		} else {
			printf(winColorWhite, ">")
		}
		printf(winColorWhite, " %s", v)
	}

	for k, v := range m {
		switch k {
		case "time", "level", "goid", "caller", "message":
			continue
		}
		if w.ANSIColor {
			if k == "error" && v != nil {
				printf(winColorRed, " %s=%v", k, v)
			} else {
				printf(winColorAqua, " %s=", k)
				printf(winColorGray, "%v", v)
			}
		} else {
			printf(winColorWhite, " %s=%v", k, v)
		}
	}

	printf(winColorWhite, " \n")

	return n, err
}
