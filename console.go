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
	Template *template.Template

	// Out is the output destination. using os.Stderr if empty.
	Out io.Writer
}

const (
	colorReset         = "\x1b[0m"
	colorBlack         = "\x1b[30m"
	colorRed           = "\x1b[31m"
	colorGreen         = "\x1b[32m"
	colorYellow        = "\x1b[33m"
	colorBlue          = "\x1b[34m"
	colorMagenta       = "\x1b[35m"
	colorCyan          = "\x1b[36m"
	colorWhite         = "\x1b[37m"
	colorGray          = "\x1b[90m"
	colorBrightRed     = "\x1b[91m"
	colorBrightGreen   = "\x1b[92m"
	colorBrightYellow  = "\x1b[93m"
	colorBrightBlue    = "\x1b[94m"
	colorBrightMagenta = "\x1b[95m"
	colorBrightCyan    = "\x1b[96m"
	colorBrightWhite   = "\x1b[97m"
)

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

	var timeField = w.TimeField
	if timeField == "" {
		timeField = "time"
	}
	if v, ok := m[timeField]; ok {
		if w.ColorOutput || w.ANSIColor {
			fmt.Fprintf(b, "%s%s%s ", colorGray, v, colorReset)
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
				fmt.Fprintf(b, " %s%s=%s%v%s", colorCyan, k, colorGray, v, colorReset)
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

const ConsoleIndentTemplate = `{{.Gray}}{{.Time}}{{.Reset}} {{.LevelColor}}{{.Level3}}{{.Reset}} {{.Caller}} {{.Cyan}}>{{.Reset}} {{.Message}}
{{range $i, $x := .KeyValue -}}
{{if eq $x.Key "error" -}}
{{ "\t" }}{{$.Red}}{{$x.Key}}={{$x.Value}}{{$.Reset -}}
{{else -}}
{{ "\t" }}{{$.Cyan}}{{$x.Key}}={{$.Reset}}{{$.Gray}}{{$x.Value}}{{$.Reset -}}
{{end}}
{{end}}{{.Stack}}`

func (w *ConsoleWriter) writet(out io.Writer, p []byte) (n int, err error) {
	type KeyValue struct {
		Key   string
		Value interface{}
	}

	o := struct {
		Reset         string
		Black         string
		Red           string
		Green         string
		Yellow        string
		Blue          string
		Magenta       string
		Cyan          string
		White         string
		Gray          string
		BrightRed     string
		BrightGreen   string
		BrightYellow  string
		BrightBlue    string
		BrightMagenta string
		BrightCyan    string
		BrightWhite   string
		LevelColor    string
		Level         string
		Level3        string
		Time          string
		Caller        string
		Message       string
		Stack         string
		KeyValue      []KeyValue
	}{
		Reset:         colorReset,
		Black:         colorBlack,
		Red:           colorRed,
		Green:         colorGreen,
		Yellow:        colorYellow,
		Blue:          colorBlue,
		Magenta:       colorMagenta,
		Cyan:          colorCyan,
		White:         colorWhite,
		Gray:          colorGray,
		BrightRed:     colorBrightRed,
		BrightGreen:   colorBrightGreen,
		BrightYellow:  colorBrightYellow,
		BrightBlue:    colorBrightBlue,
		BrightMagenta: colorBrightMagenta,
		BrightCyan:    colorBrightCyan,
		BrightWhite:   colorBrightWhite,
		LevelColor:    colorReset,
	}

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
		switch l, _ := v.(string); ParseLevel(l) {
		case DebugLevel:
			o.Level, o.LevelColor, o.Level3 = "debug", o.Yellow, "DBG"
		case InfoLevel:
			o.Level, o.LevelColor, o.Level3 = "info", o.Green, "INF"
		case WarnLevel:
			o.Level, o.LevelColor, o.Level3 = "warn", o.Red, "WRN"
		case ErrorLevel:
			o.Level, o.LevelColor, o.Level3 = "error", o.Red, "ERR"
		case FatalLevel:
			o.Level, o.LevelColor, o.Level3 = "fatal", o.Red, "FTL"
		default:
			o.Level, o.LevelColor, o.Level3 = "????", o.Yellow, "???"
		}
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
