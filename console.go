package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// ConsoleWriter parses the JSON input and writes it in an
// (optionally) colorized, human-friendly format to os.Stderr
type ConsoleWriter struct {
	// ANSIColor determines if used colorized output.
	ANSIColor bool
}

func (w *ConsoleWriter) write(p []byte) (n int, err error) {
	const (
		Reset    = "\x1b[0m"
		Red      = "\x1b[31m"
		Green    = "\x1b[32m"
		Yellow   = "\x1b[33m"
		Cyan     = "\x1b[36m"
		DarkGray = "\x1b[90m"
	)

	var m map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		n, err = os.Stderr.Write(p)
		return
	}

	b := bbpool.Get().(*bb)
	b.Reset()
	defer bbpool.Put(b)

	if v, ok := m["time"]; ok {
		if w.ANSIColor {
			fmt.Fprintf(b, "%s%s%s ", DarkGray, v, Reset)
		} else {
			fmt.Fprintf(b, "%s ", v)
		}
	}

	if v, ok := m["level"]; ok {
		var c, s string
		switch s, _ = v.(string); ParseLevel(s) {
		case DebugLevel:
			c, s = Yellow, "DBG"
		case InfoLevel:
			c, s = Green, "INF"
		case WarnLevel:
			c, s = Red, "WRN"
		case ErrorLevel:
			c, s = Red, "ERR"
		case FatalLevel:
			c, s = Red, "FTL"
		default:
			c, s = Red, "???"
		}
		if w.ANSIColor {
			fmt.Fprintf(b, "%s%s%s ", c, s, Reset)
		} else {
			fmt.Fprintf(b, "%s ", s)
		}
	}

	if v, ok := m["caller"]; ok {
		fmt.Fprintf(b, "%s ", v)
	}

	if v, ok := m["message"]; ok {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		if w.ANSIColor {
			fmt.Fprintf(b, "%s>%s %s", Cyan, Reset, v)
		} else {
			fmt.Fprintf(b, "> %s", v)
		}
	}

	for k, v := range m {
		switch k {
		case "time", "level", "caller", "message":
			continue
		}
		if w.ANSIColor {
			if k == "error" && v != nil {
				fmt.Fprintf(b, " %s%s=%v%s", Red, k, v, Reset)
			} else {
				fmt.Fprintf(b, " %s%s=%s%v%s", Cyan, k, DarkGray, v, Reset)
			}
		} else {
			fmt.Fprintf(b, " %s=%v", k, v)
		}
	}

	b.B = append(b.B, '\n')

	return os.Stderr.Write(b.B)
}
