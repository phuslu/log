package log

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strconv"
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

	// Formatter specifies an optional text formatter for creating a customized output,
	// If it is set, ColorOutput, QuoteString and EndWithMessage will be ignore.
	Formatter func(args *FormatterArgs) string

	// Writer is the output destination. using os.Stderr if empty.
	Writer io.Writer
}

// Close implements io.Closer, will closes the underlying Writer if not empty.
func (w *ConsoleWriter) Close() (err error) {
	if w.Writer != nil {
		if closer, ok := w.Writer.(io.Closer); ok {
			err = closer.Close()
		}
	}
	return
}

func (w *ConsoleWriter) write(out io.Writer, p []byte) (n int, err error) {
	var args FormatterArgs

	err = parseFormatterArgs(p, &args)
	if err != nil {
		n, err = out.Write(p)
		return
	}

	b := bbpool.Get().(*bb)
	b.B = b.B[:0]
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

	// formatting console writer
	if w.Formatter != nil {
		b.B = append(b.B, w.Formatter(&args)...)
		b.B = append(b.B, '\n')
		return out.Write(b.B)
	}

	// pretty console writer
	if w.ColorOutput {
		// colorful level string
		var levelColor = Gray
		switch args.Level {
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
		fmt.Fprintf(b, "%s%s%s %s%s%s ", Gray, args.Time, Reset, levelColor, args.Level.Three(), Reset)
		if args.Caller != "" {
			fmt.Fprintf(b, "%s %s %s>%s", args.Goid, args.Caller, Cyan, Reset)
		} else {
			fmt.Fprintf(b, "%s>%s", Cyan, Reset)
		}
		if !w.EndWithMessage {
			fmt.Fprintf(b, " %s", args.Message)
		}
		// key and values
		for _, kv := range args.KeyValues {
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
			fmt.Fprintf(b, "%s %s", Reset, args.Message)
		}
	} else {
		// header
		fmt.Fprintf(b, "%s %s ", args.Time, args.Level.Three())
		if args.Caller != "" {
			fmt.Fprintf(b, "%s %s >", args.Goid, args.Caller)
		} else {
			fmt.Fprint(b, ">")
		}
		if !w.EndWithMessage {
			fmt.Fprintf(b, " %s", args.Message)
		}
		// key and values
		for _, kv := range args.KeyValues {
			if w.QuoteString {
				fmt.Fprintf(b, " %s=%s", kv.Key, strconv.Quote(kv.Value))
			} else {
				fmt.Fprintf(b, " %s=%s", kv.Key, kv.Value)
			}
		}
		// message
		if w.EndWithMessage {
			fmt.Fprintf(b, " %s", args.Message)
		}
	}

	// stack
	if args.Stack != "" {
		b.B = append(b.B, '\n')
		b.B = append(b.B, args.Stack...)
	}

	b.B = append(b.B, '\n')

	return out.Write(b.B)
}

// FormatterArgs is a parsed sturct from json input
type FormatterArgs struct {
	Time      string // "2019-07-10T05:35:54.277Z"
	Level     Level  // InfoLevel
	Caller    string // "prog.go:42"
	Goid      string // "123"
	Message   string // "a structure message"
	Stack     string // "<stack string>"
	KeyValues []struct {
		Key   string // "foo"
		Value string // "bar"
	}
}

var errInvalidJson = errors.New("invalid json object")

func parseFormatterArgs(json []byte, e *FormatterArgs) error {
	items, err := jsonParse(json)
	if err != nil {
		return err
	}
	if len(items) <= 1 {
		return errInvalidJson
	}

	e.Time = b2s(items[1].Value)
	for i := 2; i < len(items); i += 2 {
		k, v := items[i].Value, items[i+1].Value
		switch b2s(k) {
		case "level":
			e.Level = ParseLevel(b2s(v))
		case "goid":
			e.Goid = b2s(v)
		case "caller":
			e.Caller = b2s(v)
		case "message", "msg":
			if len(v) != 0 && v[len(v)-1] == '\n' {
				e.Message = b2s(v[:len(v)-1])
			} else {
				e.Message = b2s(v)
			}
		case "stack":
			e.Stack = b2s(v)
		default:
			e.KeyValues = append(e.KeyValues, struct {
				Key   string
				Value string
			}{b2s(k), b2s(v)})
		}
	}

	return nil
}

var _ Writer = (*ConsoleWriter)(nil)
