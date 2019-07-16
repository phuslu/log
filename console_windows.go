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
	colorBlue   = 1
	colorGreen  = 2
	colorAqua   = 3
	colorRed    = 4
	colorPurple = 5
	colorYellow = 6
	colorWhite  = 7
	colorGray   = 8
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

	var printf = func(color uintptr, format string, args ...interface{}) {
		if color != colorWhite {
			SetConsoleTextAttribute(uintptr(syscall.Stderr), color)
		}
		var i int
		i, err = fmt.Fprintf(os.Stderr, format, args...)
		n += i
		if color != colorWhite {
			SetConsoleTextAttribute(uintptr(syscall.Stderr), colorWhite)
		}
	}

	if v, ok := m["time"]; ok {
		if w.ANSIColor {
			printf(colorGray, "%s ", v)
		} else {
			printf(colorWhite, "%s ", v)
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
			printf(c, "%s ", s)
		} else {
			printf(c, "%s ", s)
		}
	}

	if v, ok := m["caller"]; ok {
		printf(colorWhite, "%s ", v)
	}

	if v, ok := m["message"]; ok {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		if w.ANSIColor {
			printf(colorAqua, ">")
		} else {
			printf(colorWhite, ">")
		}
		printf(colorWhite, " %s", v)
	}

	for k, v := range m {
		switch k {
		case "time", "level", "caller", "message":
			continue
		}
		if w.ANSIColor {
			switch k {
			case "error":
				printf(colorRed, " %s=%v", k, v)
			default:
				printf(colorAqua, " %s=", k)
				printf(colorGray, "%v", v)
			}
		} else {
			printf(colorWhite, " %s=%v", k, v)
		}
	}

	printf(colorWhite, " \n")

	return n, err
}
