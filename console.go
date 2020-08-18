package log

import (
	"errors"
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

	// QuoteString determines if quoting string values.
	QuoteString bool

	// EndWithMessage determines if output message in the end.
	EndWithMessage bool

	// Template specifies an optional text/template for creating a
	// user-defined output format, available arguments are:
	//    type . struct {
	//        Time     string    // "2019-07-10T05:35:54.277Z"
	//        Level    Level     // InfoLevel
	//        Caller   string    // "prog.go:42"
	//        Goid     string    // "123"
	//        Message  string    // "a structure message"
	//        Stack    string    // "<stack string>"
	//        KeyValue []struct {
	//            Key   string       // "foo"
	//            Value string       // "bar"
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
	var t dot

	err = parseJsonDot(p, &t)
	if err != nil {
		n, err = out.Write(p)
		return
	}

	b := bbpool.Get().(*bb)
	b.Reset()
	defer bbpool.Put(b)

	const (
		Reset   = "\x1b[0m"
		Black   = "\x1b[30m"
		Red     = "\x1b[31m"
		Green   = "\x1b[32m"
		Yellow  = "\x1b[33m"
		Blue    = "\x1b[34m"
		Magenta = "\x1b[35m"
		Cyan    = "\x1b[36m"
		White   = "\x1b[37m"
		Gray    = "\x1b[90m"
	)

	// template console writer
	if w.Template != nil {
		w.Template.Execute(b, &t)
		if len(b.B) > 0 && b.B[len(b.B)-1] != '\n' {
			b.B = append(b.B, '\n')
		}
		return out.Write(b.B)
	}

	// pretty console writer
	if w.ColorOutput {
		// colorful level string
		var levelColor = Gray
		switch t.Level {
		case TraceLevel:
			levelColor = Magenta
		case DebugLevel:
			levelColor = Yellow
		case InfoLevel:
			levelColor = Green
		case WarnLevel:
			levelColor = Red
		case ErrorLevel:
			levelColor = Red
		case FatalLevel:
			levelColor = Red
		case PanicLevel:
			levelColor = Red
		}
		// header
		fmt.Fprintf(b, "%s%s%s %s%s%s ", Gray, t.Time, Reset, levelColor, t.Level.Three(), Reset)
		if t.Caller != "" {
			fmt.Fprintf(b, "%s %s ", t.Goid, t.Caller)
		}
		if !w.EndWithMessage {
			fmt.Fprintf(b, "%s>%s %s", Cyan, Reset, t.Message)
		} else {
			fmt.Fprintf(b, "%s>%s", Cyan, Reset)
		}
		// key and values
		for _, kv := range t.KeyValue {
			if w.QuoteString {
				kv.Value = strconv.Quote(kv.Value)
			}
			if kv.Key == "error" {
				fmt.Fprintf(b, " %s%s=%s%s", Red, kv.Key, kv.Value, Reset)
			} else {
				fmt.Fprintf(b, " %s%s=%s%s%s", Cyan, kv.Key, Gray, kv.Value, Reset)
			}
		}
		// message
		if w.EndWithMessage {
			fmt.Fprintf(b, "%s %s", Reset, t.Message)
		}
	} else {
		// header
		fmt.Fprintf(b, "%s %s ", t.Time, t.Level.Three())
		if t.Caller != "" {
			fmt.Fprintf(b, "%s %s >", t.Goid, t.Caller)
		} else {
			fmt.Fprint(b, ">")
		}
		if !w.EndWithMessage {
			fmt.Fprintf(b, " %s", t.Message)
		}
		// key and values
		for _, kv := range t.KeyValue {
			if w.QuoteString {
				fmt.Fprintf(b, " %s=%s", kv.Key, strconv.Quote(kv.Value))
			} else {
				fmt.Fprintf(b, " %s=%s", kv.Key, kv.Value)
			}
		}
		// message
		if w.EndWithMessage {
			fmt.Fprintf(b, " %s", t.Message)
		}
	}

	// stack
	if t.Stack != "" {
		b.B = append(b.B, '\n')
		b.B = append(b.B, t.Stack...)
	}

	b.B = append(b.B, '\n')

	return out.Write(b.B)
}

func parseJsonDot(json []byte, t *dot) error {
	items, err := jsonParse(json)
	if err != nil {
		return err
	}
	if len(items) <= 1 {
		return errors.New("invalid json object")
	}

	t.Time = b2s(items[1].Value)
	for i := 2; i < len(items); i += 2 {
		k, v := items[i].Value, items[i+1].Value
		switch b2s(k) {
		case "level":
			t.Level = noLevel
			if len(v) > 0 {
				switch v[0] {
				case 't':
					t.Level = TraceLevel
				case 'd':
					t.Level = DebugLevel
				case 'i':
					t.Level = InfoLevel
				case 'w':
					t.Level = WarnLevel
				case 'e':
					t.Level = ErrorLevel
				case 'f':
					t.Level = FatalLevel
				case 'p':
					t.Level = PanicLevel
				}
			}
		case "goid":
			t.Goid = b2s(v)
		case "caller":
			t.Caller = b2s(v)
		case "message", "msg":
			if len(v) != 0 && v[len(v)-1] == '\n' {
				t.Message = b2s(v[:len(v)-1])
			} else {
				t.Message = b2s(v)
			}
		case "stack":
			t.Stack = b2s(v)
		default:
			t.KeyValue = append(t.KeyValue, dotkv{b2s(k), b2s(v)})
		}
	}

	return nil
}

type dot struct {
	Time     string
	Level    Level
	Caller   string
	Goid     string
	Message  string
	Stack    string
	KeyValue []dotkv
}

type dotkv struct {
	Key   string
	Value string
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
	"black":      func(s string) string { return "\x1b[30m" + s + "\x1b[0m" },
	"red":        func(s string) string { return "\x1b[31m" + s + "\x1b[0m" },
	"green":      func(s string) string { return "\x1b[32m" + s + "\x1b[0m" },
	"yellow":     func(s string) string { return "\x1b[33m" + s + "\x1b[0m" },
	"blue":       func(s string) string { return "\x1b[34m" + s + "\x1b[0m" },
	"magenta":    func(s string) string { return "\x1b[35m" + s + "\x1b[0m" },
	"cyan":       func(s string) string { return "\x1b[36m" + s + "\x1b[0m" },
	"white":      func(s string) string { return "\x1b[37m" + s + "\x1b[0m" },
	"gray":       func(s string) string { return "\x1b[90m" + s + "\x1b[0m" },
	"contains":   strings.Contains,
	"endswith":   strings.HasSuffix,
	"lower":      strings.ToLower,
	"match":      path.Match,
	"quote":      strconv.Quote,
	"startswith": strings.HasPrefix,
	"title":      strings.ToTitle,
	"upper":      strings.ToUpper,
}
