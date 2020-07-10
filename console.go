package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
)

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	return isTerminal(fd, runtime.GOOS, runtime.GOARCH)
}

// ConsoleWriter parses the JSON input and writes it in an
// (optionally) colorized, human-friendly format to os.Stderr
type ConsoleWriter struct {
	// ANSIColor determines if used colorized output.
	ANSIColor bool

	// EndWithMessage determines if output message in the end.
	EndWithMessage bool
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

	if !w.EndWithMessage {
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
	} else {
		if w.ANSIColor {
			fmt.Fprintf(b, "%s>%s", Cyan, Reset)
		} else {
			fmt.Fprint(b, ">")
		}
	}

	for _, k := range jsonKeys(p) {
		switch k {
		case "time", "level", "caller", "stack", "message":
			continue
		}
		v := m[k]
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

	if w.EndWithMessage {
		if v, ok := m["message"]; ok {
			if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
				v = s[:len(s)-1]
			}
			if w.ANSIColor {
				fmt.Fprintf(b, "%s %s", Reset, v)
			} else {
				fmt.Fprintf(b, " %s", v)
			}
		}
	}

	if v, ok := m["stack"]; ok {
		b.B = append(b.B, '\n')
		if s, ok := v.(string); ok {
			b.B = append(b.B, s...)
		} else {
			fmt.Fprint(b, v)
		}
	}

	b.B = append(b.B, '\n')

	return os.Stderr.Write(b.B)
}

func jsonKeys(data []byte) (keys []string) {
	var depth, count int

	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch token.(type) {
		case json.Delim:
			switch token.(json.Delim) {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			}
		case string:
			if depth == 1 {
				if count%2 == 0 {
					keys = append(keys, token.(string))
				}
				count++
			}
		default:
			if depth == 1 {
				count++
			}
		}
	}

	return
}
