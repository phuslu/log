package log

import (
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
// Note: The performance of ConsoleWriter is not good enough, because it will
// parses JSON input into structured records, then output in a specific order.
// Roughly 2x faster than logrus.TextFormatter, 0.8x fast as zap.ConsoleEncoder,
// and 5x faster than zerolog.ConsoleWriter.
type ConsoleWriter struct {
	// ColorOutput determines if used colorized output.
	ColorOutput bool

	// QuoteString determines if quoting string values.
	QuoteString bool

	// EndWithMessage determines if output message in the end.
	EndWithMessage bool

	// Formatter specifies an optional text formatter for creating a customized output,
	// If it is set, ColorOutput, QuoteString and EndWithMessage will be ignore.
	Formatter func(w io.Writer, args *FormatterArgs) (n int, err error)

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

// WriteEntry implements Writer.
func (w *ConsoleWriter) WriteEntry(e *Entry) (int, error) {
	return w.Write(e.buf)
}

func (w *ConsoleWriter) write(out io.Writer, p []byte) (n int, err error) {
	b0 := bbpool.Get().(*bb)
	defer bbpool.Put(b0)
	b0.B = append(b0.B[:0], p...)

	var args FormatterArgs
	parseFormatterArgs(b0.B, &args)

	if len(args.Time) == 0 {
		n, err = out.Write(p)
		return
	}

	// formatting console writer
	if w.Formatter != nil {
		return w.Formatter(out, &args)
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

	// colorful level string
	var color, three string
	switch args.Level {
	case "trace":
		color, three = Magenta, "TRC"
	case "debug":
		color, three = Yellow, "DBG"
	case "info":
		color, three = Green, "INF"
	case "warn":
		color, three = Red, "WRN"
	case "error":
		color, three = Red, "ERR"
	case "fatal":
		color, three = Red, "FTL"
	case "panic":
		color, three = Red, "PNC"
	default:
		color, three = Red, "???"
	}

	// pretty console writer
	if w.ColorOutput {
		// header
		fmt.Fprintf(b, "%s%s%s %s%s%s ", Gray, args.Time, Reset, color, three, Reset)
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
			if w.QuoteString && kv.ValueType == 's' {
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
		fmt.Fprintf(b, "%s %s ", args.Time, three)
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
			if w.QuoteString && kv.ValueType == 's' {
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
		if args.Stack[len(args.Stack)-1] != '\n' {
			b.B = append(b.B, '\n')
		}
	} else {
		b.B = append(b.B, '\n')
	}
	return out.Write(b.B)
}

var _ Writer = (*ConsoleWriter)(nil)
var _ io.Writer = (*ConsoleWriter)(nil)
