package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"text/template"
)

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	return isTerminal(fd, runtime.GOOS, runtime.GOARCH)
}

// ConsoleWriter parses the JSON input and writes it in an
// (optionally) colorized, human-friendly format to Out.
type ConsoleWriter struct {
	// ColorOutput determines if used colorized output.
	ColorOutput bool

	// Deprecated: Use ColorOutput instead.
	ANSIColor bool

	// QuoteString determines if quoting string values.
	QuoteString bool

	// EndWithMessage determines if output message in the end.
	EndWithMessage bool

	// TimeField specifies the field name for time in output.
	TimeField string

	// Template determines console output template if not empty.
	// see https://github.com/phuslu/log/blob/master/console.go#L328
	// for available Template Arguments.
	Template *template.Template

	// Out is the output destination. using os.Stderr if empty.
	Out io.Writer
}

func (w *ConsoleWriter) write(out io.Writer, p []byte) (n int, err error) {
	if w.Template != nil {
		return w.writet(out, p)
	}

	var m map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		n, err = out.Write(p)
		return
	}

	b := bbpool.Get().(*bb)
	b.Reset()
	defer bbpool.Put(b)

	const (
		Reset         = "\x1b[0m"
		Black         = "\x1b[30m"
		Red           = "\x1b[31m"
		Green         = "\x1b[32m"
		Yellow        = "\x1b[33m"
		Blue          = "\x1b[34m"
		Magenta       = "\x1b[35m"
		Cyan          = "\x1b[36m"
		White         = "\x1b[37m"
		Gray          = "\x1b[90m"
		BrightRed     = "\x1b[91m"
		BrightGreen   = "\x1b[92m"
		BrightYellow  = "\x1b[93m"
		BrightBlue    = "\x1b[94m"
		BrightMagenta = "\x1b[95m"
		BrightCyan    = "\x1b[96m"
		BrightWhite   = "\x1b[97m"
	)

	var timeField = w.TimeField
	if timeField == "" {
		timeField = "time"
	}
	if v, ok := m[timeField]; ok {
		if w.ColorOutput || w.ANSIColor {
			fmt.Fprintf(b, "%s%s%s ", Gray, v, Reset)
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
			c, s = Yellow, "???"
		}
		if w.ColorOutput || w.ANSIColor {
			fmt.Fprintf(b, "%s%s%s ", c, s, Reset)
		} else {
			fmt.Fprintf(b, "%s ", s)
		}
	}

	if v, ok := m["caller"]; ok {
		fmt.Fprintf(b, "%s ", v)
	}

	var msgField = "message"
	if _, ok := m[msgField]; !ok {
		if _, ok := m["msg"]; ok {
			msgField = "msg"
		}
	}

	if v, ok := m[msgField]; ok && !w.EndWithMessage {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		if w.ColorOutput || w.ANSIColor {
			fmt.Fprintf(b, "%s>%s %s", Cyan, Reset, v)
		} else {
			fmt.Fprintf(b, "> %s", v)
		}
	} else {
		if w.ColorOutput || w.ANSIColor {
			fmt.Fprintf(b, "%s>%s", Cyan, Reset)
		} else {
			fmt.Fprint(b, ">")
		}
	}

	for _, k := range jsonKeys(p) {
		switch k {
		case timeField, msgField, "level", "caller", "stack":
			continue
		}
		v := m[k]
		if w.QuoteString {
			if s, ok := v.(string); ok {
				v = strconv.Quote(s)
			}
		}
		if w.ColorOutput || w.ANSIColor {
			if k == "error" && v != nil {
				fmt.Fprintf(b, " %s%s=%v%s", Red, k, v, Reset)
			} else {
				fmt.Fprintf(b, " %s%s=%s%v%s", Cyan, k, Gray, v, Reset)
			}
		} else {
			fmt.Fprintf(b, " %s=%v", k, v)
		}
	}

	if w.EndWithMessage {
		if v, ok := m[msgField]; ok {
			if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
				v = s[:len(s)-1]
			}
			if w.ColorOutput || w.ANSIColor {
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
			data, _ := json.MarshalIndent(v, "", "  ")
			b.B = append(b.B, data...)
		}
	}

	b.B = append(b.B, '\n')

	return out.Write(b.B)
}

func (w *ConsoleWriter) writet(out io.Writer, p []byte) (n int, err error) {
	type KeyValue struct {
		Key   string
		Value interface{}
	}

	o := struct {
		Time     string
		Level    string
		Caller   string
		Message  string
		Stack    string
		KeyValue []KeyValue
	}{}

	var m map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		n, err = out.Write(p)
		return
	}

	var timeField = w.TimeField
	if timeField == "" {
		timeField = "time"
	}
	if v, ok := m[timeField]; ok {
		o.Time = v.(string)
	}

	if v, ok := m["level"]; ok {
		o.Level = v.(string)
	}

	if v, ok := m["caller"]; ok {
		o.Caller = v.(string)
	}

	var msgField = "message"
	if _, ok := m[msgField]; !ok {
		if _, ok := m["msg"]; ok {
			msgField = "msg"
		}
	}

	if v, ok := m[msgField]; ok {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		o.Message = v.(string)
	}

	if v, ok := m["stack"]; ok {
		if s, ok := v.(string); ok {
			o.Stack = s
		} else {
			b, _ := json.MarshalIndent(v, "", "  ")
			o.Stack = string(b)
		}
	}

	for _, k := range jsonKeys(p) {
		switch k {
		case timeField, msgField, "level", "caller", "stack":
			continue
		}
		v := m[k]
		if w.QuoteString {
			if s, ok := v.(string); ok {
				v = strconv.Quote(s)
			}
		}
		o.KeyValue = append(o.KeyValue, KeyValue{k, fmt.Sprint(v)})
	}

	b := bbpool.Get().(*bb)
	b.Reset()
	defer bbpool.Put(b)

	w.Template.Execute(b, &o)
	if len(b.B) > 0 && b.B[len(b.B)-1] != '\n' {
		b.B = append(b.B, '\n')
	}

	return out.Write(b.B)
}

func jsonKeys(data []byte) (keys []string) {
	var depth, count int

	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		// fmt.Printf("==== %d %d <%T> %v\n", depth, count, token, token)
		switch token.(type) {
		case json.Delim:
			switch token.(json.Delim) {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
				if depth == 1 {
					count++
				}
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

// ColorTemplate provides a pre-defined text/template for console color output
//
//    type . struct {
//        Time     string    // "2019-07-10T05:35:54.277Z"
//        Level    string    // "info"
//        Caller   string    // "prog.go:42"
//        Message  string    // "a structure message"
//        Stack    string    // "<stack string>"
//        KeyValue []struct {
//            Key   string       // "foo"
//            Value interface{}  // "bar"
//        }
//    }
//
// Note: use [sprig](https://github.com/Masterminds/sprig) to provides more template functions.
const ColorTemplate = `{{"\x1b[90m"}}{{.Time}}{{"\x1b[0m " -}}
{{if eq "debug" .Level }}{{"\x1b[33mDBG\x1b[0m " -}}
{{else if eq "info"  .Level}}{{"\x1b[32mDBG\x1b[0m " -}}
{{else if eq "warn"  .Level}}{{"\x1b[31mWRN\x1b[0m " -}}
{{else if eq "error" .Level}}{{"\x1b[31mERR\x1b[0m " -}}
{{else if eq "fatal" .Level}}{{"\x1b[31mFTL\x1b[0m " -}}
{{else}}{{"\x1b[31m???\x1b[0m "}}{{end -}}
{{.Caller}}{{" \x1b[36m>\x1b[0m "}}{{.Message}}
{{range .KeyValue -}}
{{if eq .Key "error" -}}{{"\t\x1b[31m"}}{{.Key}}={{.Value}}{{"\x1b[0m" -}}
{{else}}{{"\t\x1b[36m"}}{{.Key}}={{"\x1b[90m"}}{{.Value}}{{"\x1b[0m"}}{{end}}
{{end}}{{.Stack}}`
