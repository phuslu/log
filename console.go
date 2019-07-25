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
	ansiColorReset    = "\x1b[0m"
	ansiColorRed      = "\x1b[31m"
	ansiColorGreen    = "\x1b[32m"
	ansiColorYellow   = "\x1b[33m"
	ansiColorCyan     = "\x1b[36m"
	ansiColorDarkGray = "\x1b[90m"
)

func (w *ConsoleWriter) vtWrite(p []byte) (n int, err error) {
	var m map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		n, err = os.Stderr.Write(p)
		return
	}

	var b bytes.Buffer

	if v, ok := m["time"]; ok {
		if w.ANSIColor {
			fmt.Fprintf(&b, "%s%s%s ", ansiColorDarkGray, v, ansiColorReset)
		} else {
			fmt.Fprintf(&b, "%s ", v)
		}
	}

	if v, ok := m["level"]; ok {
		var c, s string
		switch s, _ = v.(string); ParseLevel(s) {
		case DebugLevel:
			c, s = ansiColorYellow, "DBG"
		case InfoLevel:
			c, s = ansiColorGreen, "INF"
		case WarnLevel:
			c, s = ansiColorRed, "WRN"
		case ErrorLevel:
			c, s = ansiColorRed, "ERR"
		case FatalLevel:
			c, s = ansiColorRed, "FTL"
		case PanicLevel:
			c, s = ansiColorRed, "PNC"
		default:
			c, s = ansiColorRed, "???"
		}
		if w.ANSIColor {
			fmt.Fprintf(&b, "%s%s%s ", c, s, ansiColorReset)
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
			fmt.Fprintf(&b, "%s>%s %s", ansiColorCyan, ansiColorReset, v)
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
				fmt.Fprintf(&b, " %s%s=%v%s", ansiColorRed, k, v, ansiColorReset)
			} else {
				fmt.Fprintf(&b, " %s%s=%s%v%s", ansiColorCyan, k, ansiColorDarkGray, v, ansiColorReset)
			}
		} else {
			fmt.Fprintf(&b, " %s=%v", k, v)
		}
	}

	b.WriteByte('\n')

	return os.Stderr.Write(b.Bytes())
}
