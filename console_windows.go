// +build windows

package log

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"syscall"
)

type ConsoleWriter struct {
	ANSIColor bool

	mu sync.Mutex
}

const (
	colorWhite    = 0x07
	colorRed      = 0x04
	colorGreen    = 0x02
	colorYellow   = 0x06
	colorCyan     = 0x03
	colorDarkGray = 0x08
)

var (
	SetConsoleTextAttribute = syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleTextAttribute").Call
)

func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	var m map[string]interface{}

	err = json.Unmarshal(p, &m)
	if err != nil {
		return
	}

	var cprintf = func(color uintptr, format string, args ...interface{}) {
		if color != 0 {
			SetConsoleTextAttribute(uintptr(syscall.Stderr), color)
		}
		var i int
		i, err = fmt.Fprintf(os.Stderr, format, args...)
		n += i
		if color != 0 {
			SetConsoleTextAttribute(uintptr(syscall.Stderr), colorWhite)
		}
	}

	if v, ok := m["time"]; ok {
		if w.ANSIColor {
			cprintf(colorDarkGray, "%s ", v)
		} else {
			cprintf(0, "%s ", v)
		}
	}

	if v, ok := m["level"]; ok {
		var s string
		var c uintptr
		switch s, _ = v.(string); ParseLevel(s) {
		case DebugLevel:
			c, s = colorYellow, "DBG"
		case InfoLevel:
			c, s = colorGreen, "INF"
		case WarnLevel:
			c, s = colorRed, "WRN"
		case ErrorLevel:
			c, s = colorRed, "ERR"
		case FatalLevel:
			c, s = colorRed, "FTL"
		case PanicLevel:
			c, s = colorRed, "PNC"
		default:
			c, s = colorRed, "???"
		}
		if w.ANSIColor {
			cprintf(c, "%s ", v)
		} else {
			cprintf(c, "%s ", v)
		}
	}

	if v, ok := m["caller"]; ok {
		cprintf(0, "%s ", v)
	}

	if v, ok := m["message"]; ok {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		if w.ANSIColor {
			cprintf(colorCyan, ">")
		} else {
			cprintf(0, ">")
		}
		cprintf(0, " %s", v)
	}

	for k, v := range m {
		switch k {
		case "time", "level", "caller", "message":
			continue
		}
		if w.ANSIColor {
			switch k {
			case "error":
				cprintf(colorRed, " %s=%v", k, v)
			default:
				cprintf(colorCyan, " %s=", k)
				cprintf(0, " %v", v)
			}
		} else {
			cprintf(0, " %s=%v", k, v)
		}
	}

	cprintf(0, " \n")

	return n, err
}
