package log

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestLoggerCategorizedLogLevels(t *testing.T) {
	var b bytes.Buffer
	logger := Logger{Level: DebugLevel, Writer: &IOWriter{Writer: &b}}

	// Categorized logger must be indepented of base logger
	cat1Logger := logger.Categorized("cat1")
	cat1Logger.SetLevel(TraceLevel)
	logger.Debug().Msg("logger debug here")
	cat1Logger.Debug().Msg("cat1Logger debug here")
	if !strings.Contains(b.String(), `"logger debug here"`) {
		t.Fatal("logger.Debug must be logged")
	}
	if !strings.Contains(b.String(), `"cat1Logger debug here"`) {
		t.Fatal("cat1Logger.Debug must be logged for ")
	}
	if !strings.Contains(b.String(), `"category":"cat1"`) {
		t.Fatal("cat1Logger.Debug is missing category")
	}

	// Changing loglevel on category must not change base logger
	b.Reset()
	cat1Logger.SetLevel(InfoLevel)
	logger.Debug().Msg("logger debug here")
	cat1Logger.Debug().Msg("cat1Logger debug here")
	if !strings.Contains(b.String(), `"logger debug here"`) {
		t.Fatal("logger.Debug must be logged")
	}
	if strings.Contains(b.String(), `"cat1Logger debug here"`) {
		t.Fatal("cat1Logger.Debug must be logged")
	}

	// Loglevel on other categorized logger has own loglevel
	b.Reset()
	cat2Logger := logger.Categorized("cat2")
	cat2Logger.SetLevel(TraceLevel)
	cat2Logger.Trace().Msg("cat2Logger trace here")
	if !strings.Contains(b.String(), `"cat2Logger trace here"`) {
		t.Fatal("logger.Debug for category cat2 must be logged")
	}
	if !strings.Contains(b.String(), `"category":"cat2"`) {
		t.Fatal("cat2Logger.Debug is missing category")
	}
}

func BenchmarkCategorizedLogger(b *testing.B) {
	logger := Logger{
		TimeFormat: TimeFormatUnix,
		Level:      DebugLevel,
		Writer:     IOWriter{io.Discard},
	}

	catLogger := logger.Categorized("one")
	catLogger.SetLevel(DebugLevel)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		catLogger.Debug().Str("foo", "bar").Msgf("hello %s", "world")
	}
}
