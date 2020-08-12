package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
	"text/template"
)

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	return isTerminal(fd, runtime.GOOS, runtime.GOARCH)
}

// ConsoleWriter parses the JSON input and writes it in an
// (optionally) colorized, human-friendly format to Writer.
//
// Default output format:
//     {Time} {Level} {Goid} {Caller} > {Message} {Key}={Value} {Key}={Value}
//
// Note: ConsoleWriter performance is not good, it will parses JSON input into
// structured records, then outputs them in a specific order.
type ConsoleWriter struct {
	// ColorOutput determines if used colorized output.
	ColorOutput bool

	// Deprecated: Use ColorOutput instead.
	ANSIColor bool

	// QuoteString determines if quoting string values.
	QuoteString bool

	// EndWithMessage determines if output message in the end.
	EndWithMessage bool

	// Template specifies an optional text/template for creating a
	// user-defined output format, available arguments are:
	//    type . struct {
	//        Time     string    // "2019-07-10T05:35:54.277Z"
	//        Level    string    // "info"
	//        Caller   string    // "prog.go:42"
	//        Goid     string    // "123"
	//        Message  string    // "a structure message"
	//        Stack    string    // "<stack string>"
	//        KeyValue []struct {
	//            Key   string       // "foo"
	//            Value interface{}  // "bar"
	//        }
	//    }
	// See https://github.com/phuslu/log#template-console-writer for example.
	//
	// If Template is not nil, ColorOutput, QuoteString and EndWithMessage are override.
	Template *template.Template

	// Writer is the output destination. using os.Stderr if empty.
	Writer io.Writer
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

	keys := jsonKeys(p)
	if len(keys) < 2 {
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

	// time
	var timeField = "time"
	v, ok := m[timeField]
	if !ok {
		timeField = keys[0]
		v = m[timeField]
	}
	if w.ColorOutput || w.ANSIColor {
		fmt.Fprintf(b, "%s%s%s ", Gray, v, Reset)
	} else {
		fmt.Fprintf(b, "%s ", v)
	}

	// level
	if v, ok := m["level"]; ok {
		var c, s string
		switch s, _ = v.(string); ParseLevel(s) {
		case TraceLevel:
			c, s = Magenta, "TRC"
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
		case PanicLevel:
			c, s = Red, "PNC"
		default:
			c, s = Yellow, "???"
		}
		if w.ColorOutput || w.ANSIColor {
			fmt.Fprintf(b, "%s%s%s ", c, s, Reset)
		} else {
			fmt.Fprintf(b, "%s ", s)
		}
	}

	// goid
	if v, ok := m["goid"]; ok {
		fmt.Fprintf(b, "%s ", v)
	}

	// caller
	if v, ok := m["caller"]; ok {
		fmt.Fprintf(b, "%s ", v)
	}

	// message
	var msgField = "message"
	if _, ok := m[msgField]; !ok {
		if _, ok := m["msg"]; ok {
			msgField = "msg"
		}
	}

	// >
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

	// key and values
	for _, k := range keys {
		switch k {
		case timeField, msgField, "level", "goid", "caller", "stack":
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

	// message
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

	// stack
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

	dot := struct {
		Time     string
		Level    Level
		Caller   string
		Goid     string
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

	keys := jsonKeys(p)
	if len(keys) < 2 {
		n, err = out.Write(p)
		return
	}

	// time
	var timeField = "time"
	v, ok := m[timeField]
	if !ok {
		timeField = keys[0]
		v = m[timeField]
	}
	dot.Time, ok = v.(string)
	if !ok {
		dot.Time = fmt.Sprint(v)
	}

	// level
	if v, ok := m["level"]; ok {
		dot.Level = ParseLevel(v.(string))
	}

	// caller
	if v, ok := m["caller"]; ok {
		dot.Caller = v.(string)
	}

	// goid
	if v, ok := m["goid"]; ok {
		dot.Goid = v.(json.Number).String()
	}

	// message
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
		dot.Message = v.(string)
	}

	// stack
	if v, ok := m["stack"]; ok {
		if s, ok := v.(string); ok {
			dot.Stack = s
		} else {
			b, _ := json.MarshalIndent(v, "", "  ")
			dot.Stack = string(b)
		}
	}

	// key and values
	for _, k := range keys {
		switch k {
		case timeField, msgField, "level", "caller", "goid", "stack":
			continue
		}
		v := m[k]
		if w.QuoteString {
			if s, ok := v.(string); ok {
				v = strconv.Quote(s)
			}
		}
		dot.KeyValue = append(dot.KeyValue, KeyValue{k, fmt.Sprint(v)})
	}

	b := bbpool.Get().(*bb)
	b.Reset()
	defer bbpool.Put(b)

	w.Template.Execute(b, &dot)
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
// Note: use [sprig](https://github.com/Masterminds/sprig) to introduce more template functions.
const ColorTemplate = `{{gray .Time -}}
{{if eq .Level -1 }}{{magenta " TRC " -}}
{{else if eq .Level 0 }}{{yellow " DBG " -}}
{{else if eq .Level 1}}{{green " INF " -}}
{{else if eq .Level 2}}{{red " WRN " -}}
{{else if eq .Level 3}}{{red " ERR " -}}
{{else if eq .Level 4}}{{red " FTL " -}}
{{else if eq .Level 5}}{{red " PNC " -}}
{{else}}{{red " ??? "}}{{end -}}
{{.Goid}} {{.Caller}}{{cyan " >" -}}
{{range .KeyValue -}}
{{if eq .Key "error" }} {{red (printf "%s%s%s" .Key "=" .Value) -}}
{{else}} {{cyan .Key}}={{gray .Value}}{{end -}}
{{end}} {{.Message}}
{{.Stack}}`

// ColorFuncMap provides a pre-defined template functions for color string
var ColorFuncMap = template.FuncMap{
	"black":     func(s string) string { return "\x1b[30m" + s + "\x1b[0m" },
	"red":       func(s string) string { return "\x1b[31m" + s + "\x1b[0m" },
	"green":     func(s string) string { return "\x1b[32m" + s + "\x1b[0m" },
	"yellow":    func(s string) string { return "\x1b[33m" + s + "\x1b[0m" },
	"blue":      func(s string) string { return "\x1b[34m" + s + "\x1b[0m" },
	"magenta":   func(s string) string { return "\x1b[35m" + s + "\x1b[0m" },
	"cyan":      func(s string) string { return "\x1b[36m" + s + "\x1b[0m" },
	"white":     func(s string) string { return "\x1b[37m" + s + "\x1b[0m" },
	"gray":      func(s string) string { return "\x1b[90m" + s + "\x1b[0m" },
	"contains":  strings.Contains,
	"endsswith": strings.HasSuffix,
	"low":       strings.ToLower,
	"match":     path.Match,
	"quote":     strconv.Quote,
	"statswith": strings.HasPrefix,
	"title":     strings.ToTitle,
	"upper":     strings.ToUpper,
}
