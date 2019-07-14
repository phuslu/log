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

func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
	var m map[string]interface{}

	err = json.Unmarshal(p, &m)
	if err != nil {
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
		var s string
		var c color
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
		case "time", "level", "caller", "message":
			continue
		}
		if w.ANSIColor {
			switch k {
			case "error":
				fmt.Fprintf(&b, " %s%s=%v%s", colorRed, k, v, colorReset)
			default:
				fmt.Fprintf(&b, " %s%s=%s%v", colorCyan, k, colorReset, v)
			}
		} else {
			fmt.Fprintf(&b, " %s=%v", k, v)
		}
	}

	b.WriteByte('\n')

	return os.Stderr.Write(b.Bytes())
}
