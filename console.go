// +build !windows

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type ConsoleWriter struct {
	ANSIColor bool
}

const (
	colorReset    = "\x1b[0m"
	colorRed      = "\x1b[31m"
	colorGreen    = "\x1b[32m"
	colorYellow   = "\x1b[33m"
	colorCyan     = "\x1b[36m"
	colorDarkGray = "\x1b[90m"
)

func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
	var m map[string]interface{}

	err = json.Unmarshal(p, &m)
	if err != nil {
		n, err = os.Stderr.Write(p)
		return
	}

	var b bytes.Buffer

	if v, ok := m["time"]; ok {
		if w.ANSIColor {
			fmt.Fprintf(&b, "%s%s%s ", colorDarkGray, v, colorReset)
		} else {
			fmt.Fprintf(&b, "%s ", v)
		}
	}

	if v, ok := m["level"]; ok {
		var c, s string
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
			fmt.Fprintf(&b, "%s%s%s ", c, s, colorReset)
		} else {
			fmt.Fprintf(&b, "%s ", s)
		}
	}

	if v, ok := m["goid"]; ok {
		fmt.Fprintf(&b, "%s ", v)
	}

	if v, ok := m["caller"]; ok {
		fmt.Fprintf(&b, "%s ", v)
	}

	if v, ok := m["message"]; ok {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		if w.ANSIColor {
			fmt.Fprintf(&b, "%s>%s %s", colorCyan, colorReset, v)
		} else {
			fmt.Fprintf(&b, "> %s", v)
		}
	}

	for k, v := range m {
		switch k {
		case "time", "level", "goid", "caller", "message":
			continue
		}
		if w.ANSIColor {
			if k == "error" && v != nil {
				fmt.Fprintf(&b, " %s%s=%v%s", colorRed, k, v, colorReset)
			} else {
				fmt.Fprintf(&b, " %s%s=%s%v%s", colorCyan, k, colorDarkGray, v, colorReset)
			}
		} else {
			fmt.Fprintf(&b, " %s=%v", k, v)
		}
	}

	b.WriteByte('\n')

	return os.Stderr.Write(b.Bytes())
}
