package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"text/template"
)

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	return isTerminal(fd, runtime.GOOS, runtime.GOARCH)
}

// ConsoleWriter parses the JSON input and writes it in an
// (optionally) colorized, human-friendly format to os.Stderr
type ConsoleWriter struct {
	// ColorOutput determines if used colorized output.
	ColorOutput bool

	// Deprecated: Use ColorOutput instead.
	ANSIColor bool

	// QuoteString determines if quoting string values.
	QuoteString bool

	// EndWithMessage determines if output message in the end.
	EndWithMessage bool

	// Template determines console output template if not empty.
	Template *template.Template

	// TimeField specifies the time filed name of output message.
	TimeField string
}

const (
	colorReset    = "\x1b[0m"
	colorRed      = "\x1b[31m"
	colorGreen    = "\x1b[32m"
	colorYellow   = "\x1b[33m"
	colorCyan     = "\x1b[36m"
	colorDarkGray = "\x1b[90m"
)

func (w *ConsoleWriter) writeTo(out io.Writer, p []byte) (n int, err error) {
	if w.Template != nil {
		return w.tmplWrite(p)
	}

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

	var timeField = w.TimeField
	if timeField == "" {
		timeField = "time"
	}
	if v, ok := m[timeField]; ok {
		if w.ColorOutput || w.ANSIColor {
			fmt.Fprintf(b, "%s%s%s ", colorDarkGray, v, colorReset)
		} else {
			fmt.Fprintf(b, "%s ", v)
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
		default:
			c, s = colorYellow, "???"
		}
		if w.ColorOutput || w.ANSIColor {
			fmt.Fprintf(b, "%s%s%s ", c, s, colorReset)
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
			fmt.Fprintf(b, "%s>%s %s", colorCyan, colorReset, v)
		} else {
			fmt.Fprintf(b, "> %s", v)
		}
	} else {
		if w.ColorOutput || w.ANSIColor {
			fmt.Fprintf(b, "%s>%s", colorCyan, colorReset)
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
				fmt.Fprintf(b, " %s%s=%v%s", colorRed, k, v, colorReset)
			} else {
				fmt.Fprintf(b, " %s%s=%s%v%s", colorCyan, k, colorDarkGray, v, colorReset)
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
				fmt.Fprintf(b, "%s %s", colorReset, v)
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

const ConsoleIndentTemplate = `{{.DarkGray}}{{.Time}}{{.Reset}} {{.LevelColor}}{{.Level}}{{.Reset}} {{.Caller}} {{.Cyan}}>{{.Reset}} {{.Message}}
{{range $i, $x := .KeyValue}}{{if eq $x.Key "error" -}}
{{ "\t" }}{{$.Red}}{{$x.Key}}={{$x.Value}}{{$.Reset -}}
{{else -}}
{{ "\t" }}{{$.Cyan}}{{$x.Key}}={{$.Reset}}{{$.DarkGray}}{{$x.Value}}{{$.Reset -}}
{{end}}
{{end}}{{.Stack}}`

func (w *ConsoleWriter) tmplWrite(p []byte) (n int, err error) {
	type KeyValue struct {
		Key   string
		Value interface{}
	}

	output := struct {
		Reset      string
		Red        string
		Green      string
		Yellow     string
		Cyan       string
		DarkGray   string
		LevelColor string
		Time       string
		Level      string
		Caller     string
		Message    string
		Stack      string
		KeyValue   []KeyValue
	}{
		Reset:      colorReset,
		Red:        colorRed,
		Green:      colorGreen,
		Yellow:     colorYellow,
		Cyan:       colorCyan,
		DarkGray:   colorDarkGray,
		LevelColor: colorReset,
	}

	var m map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		n, err = os.Stderr.Write(p)
		return
	}

	var timeField = w.TimeField
	if timeField == "" {
		timeField = "time"
	}
	if v, ok := m[timeField]; ok {
		output.Time = v.(string)
	}

	if v, ok := m["level"]; ok {
		switch l, _ := v.(string); ParseLevel(l) {
		case DebugLevel:
			output.LevelColor, output.Level = output.Yellow, "DBG"
		case InfoLevel:
			output.LevelColor, output.Level = output.Green, "INF"
		case WarnLevel:
			output.LevelColor, output.Level = output.Red, "WRN"
		case ErrorLevel:
			output.LevelColor, output.Level = output.Red, "ERR"
		case FatalLevel:
			output.LevelColor, output.Level = output.Red, "FTL"
		default:
			output.LevelColor, output.Level = output.Yellow, "???"
		}
	}

	if v, ok := m["caller"]; ok {
		output.Caller = v.(string)
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
		output.Message = v.(string)
	}

	if v, ok := m["stack"]; ok {
		if s, ok := v.(string); ok {
			output.Stack = s
		} else {
			b, _ := json.MarshalIndent(v, "", "  ")
			output.Stack = string(b)
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
		output.KeyValue = append(output.KeyValue, KeyValue{k, fmt.Sprint(v)})
	}

	b := bbpool.Get().(*bb)
	b.Reset()
	defer bbpool.Put(b)

	w.Template.Execute(b, &output)
	if len(b.B) > 0 && b.B[len(b.B)-1] != '\n' {
		b.B = append(b.B, '\n')
	}

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
