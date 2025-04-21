//go:build go1.21

//nolint:staticcheck
package log

import (
	"fmt"
	stdLog "log"
	"log/slog"
	"os"
	"testing"
)

func TestStdLogWriter(t *testing.T) {
	w := &stdLogWriter{
		Logger: Logger{
			Level:  InfoLevel,
			Writer: IOWriter{os.Stderr},
		},
	}

	fmt.Fprint(w, "hello from stdLog debug Print")
	fmt.Fprintln(w, "hello from stdLog debug Println")
	fmt.Fprintf(w, "hello from stdLog debug %s", "Printf")
}

func TestStdLogLogger(t *testing.T) {
	logger := Logger{
		Level:   DebugLevel,
		Caller:  -1,
		Context: NewContext(nil).Str("tag", "std_log").Value(),
		Writer:  &ConsoleWriter{ColorOutput: true, EndWithMessage: true},
	}

	stdLog := logger.Std("", stdLog.LstdFlags)
	stdLog.Print("hello from stdLog Print")
	stdLog.Println("hello from stdLog Println")
	stdLog.Printf("hello from stdLog %s", "Printf")

	stdLog = logger.Std("", 0)
	stdLog.Print("hello from stdLog Print")
	stdLog.Println("hello from stdLog Println")
	stdLog.Printf("hello from stdLog %s", "Printf")
}

func TestStdSlogNormal(t *testing.T) {
	var logger *slog.Logger = (&Logger{
		Level:      InfoLevel,
		TimeField:  "date",
		TimeFormat: "2006-01-02",
		Caller:     1,
	}).Slog()

	logger.Info("hello from slog Info")
	logger.Warn("hello from slog Warn")
	logger.Error("hello from slog Error")
}

func TestStdSlogAttrs(t *testing.T) {
	var logger *slog.Logger = (&Logger{
		Level:      InfoLevel,
		TimeField:  "date",
		TimeFormat: "2006-01-02",
		Caller:     -1,
	}).Slog()

	sublogger := logger.With("logger_name", "sub_logger").WithGroup("g").With("everything", 42)
	sublogger.Info("hello from sub attr slog")
	logger.Info("hello from origin slog")
}

func TestStdSlogGroup(t *testing.T) {
	var logger *slog.Logger = (&Logger{
		Level:  InfoLevel,
		Caller: 1,
	}).Slog()

	logger1 := logger.WithGroup("g").With("1", "2").With("3", "4")
	logger1.Info("hello from group slog 1")
	logger1.Info("hello from group slog 2")

	logger2 := logger1.WithGroup("g1").With("a", "b").With("c", "d").
		WithGroup("g2").With("foo", "bar").With("bar", "foo").
		WithGroup("g3").With("x", 1).With("y", 2).With("z", 3)
	logger2.Info("hello from group slog 3")
	logger2.Info("hello from group slog 4")

	logger1.Info("hello from group slog 1")
	logger1.Info("hello from group slog 2")
}
