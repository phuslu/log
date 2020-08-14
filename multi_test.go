package log

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestNewMultiWriters(t *testing.T) {
	// TODO: Use mock writers to validate that the logs are routed correctly
	w := &MultiWriter{
		InfoWriter: &FileWriter{
			Filename: "file-info.log",
		},
		WarnWriter: &FileWriter{
			Filename: "file-warn.log",
		},
		ErrorWriter: &FileWriter{
			Filename: "file-error.log",
		},
	}
	logger := Logger{
		Level:        InfoLevel,
		Caller:       1,
		MultiWriters: w,
	}
	assertNLogs := func(want int) {
		matches, _ := filepath.Glob("file-*.*.log")
		if len(matches) != want {
			t.Fatalf("filepath glob return %+v number mismatch, got %+v want %+v", matches, len(matches), want)
		}
	}

	logger.Info().Int("id", 42).Msg("I'm loving it.")
	assertNLogs(1)

	logger.Warn().Int("id", 43).Msg("I double dare you.")
	assertNLogs(2)

	logger.Error().Str("action", "cleanup").Msg("World")
	assertNLogs(3)

	matches, _ := filepath.Glob("file-*.log")
	for i := range matches {
		err := os.Remove(matches[i])
		if err != nil {
			t.Fatalf("os remove %s error: %+v", matches[i], err)
		}
	}
}

func TestMultiWriter(t *testing.T) {
	w := &MultiWriter{
		InfoWriter: &FileWriter{
			Filename: "file-info.log",
		},
		WarnWriter: &FileWriter{
			Filename: "file-warn.log",
		},
		ErrorWriter: &FileWriter{
			Filename: "file-error.log",
		},
		StderrWriter: &ConsoleWriter{
			ColorOutput: true,
		},
		StderrLevel: ErrorLevel,
	}

	var err error
	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err = fmt.Fprintf(w, `{"ts":1234567890,"level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json mutli writer error: %+v", err)
		}
		_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json mutli writer error: %+v", err)
		}
		_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277+08:00","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json mutli writer error: %+v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Errorf("test close mutli writer error: %+v", err)
	}

	matches, _ := filepath.Glob("file-*.*.log")
	want := 3
	if len(matches) != want {
		t.Fatalf("filepath glob return %+v number mismatch, got %+v want %+v", matches, len(matches), want)
	}

	for i := range matches {
		err := os.Remove(matches[i])
		if err != nil {
			t.Fatalf("os remove %s error: %+v", matches[i], err)
		}
	}
}

type errorWriter struct {
	io.WriteCloser
}

var errorWriterOK = errors.New("errorWriter return OK")

func (ew errorWriter) Write(p []byte) (n int, err error) {
	n, err = ew.WriteCloser.Write(p)
	if err == nil {
		err = errorWriterOK
	}
	return
}

func (ew errorWriter) Close() (err error) {
	err = ew.WriteCloser.Close()
	if err == nil {
		err = errorWriterOK
	}
	return
}

func TestMultiWriterError(t *testing.T) {
	w := &MultiWriter{
		InfoWriter: errorWriter{&FileWriter{
			Filename: "file-info.log",
		}},
		WarnWriter: errorWriter{&FileWriter{
			Filename: "file-warn.log",
		}},
		ErrorWriter: errorWriter{&FileWriter{
			Filename: "file-error.log",
		}},
		StderrWriter: &ConsoleWriter{
			ColorOutput: true,
		},
		StderrLevel: ErrorLevel,
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err != errorWriterOK {
			t.Errorf("test json mutli writer error: %+v", err)
		}
	}

	if err := w.Close(); err != errorWriterOK {
		t.Errorf("test close mutli writer error: %+v", err)
	}

	matches, _ := filepath.Glob("file-*.*.log")
	want := 3
	if len(matches) != want {
		t.Fatalf("filepath glob return %+v number mismatch, got %+v want %+v", matches, len(matches), want)
	}

	matches, _ = filepath.Glob("file-*.log")
	for i := range matches {
		err := os.Remove(matches[i])
		if err != nil {
			t.Fatalf("os remove %s error: %+v", matches[i], err)
		}
	}
}

func TestMultiWriterLevel(t *testing.T) {
	w := &MultiWriter{
		StderrWriter: &ConsoleWriter{
			ColorOutput: true,
		},
		StderrLevel: InfoLevel,
		ParseLevel: func(data []byte) (level Level) {
			v := struct {
				Level string `json:"level"`
			}{}
			if err := json.Unmarshal(data, &v); err == nil {
				level = ParseLevel(v.Level)
			}
			return
		},
	}

	var err error
	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err = fmt.Fprintf(w, `{"time":1234567890,"level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json mutli writer error: %+v", err)
		}
		_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json mutli writer error: %+v", err)
		}
		_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277+08:00","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json mutli writer error: %+v", err)
		}
	}

	w.ParseLevel = func(p []byte) Level { return ParseLevel(string(p[49])) }
	w.StderrLevel = ErrorLevel
	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277+08:00","level":"%s","foo":"bar","n":42,"message":"hello user-defined parse level mutli writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json mutli writer error: %+v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Errorf("test close mutli writer error: %+v", err)
	}
}

func BenchmarkNewMultiWriter(b *testing.B) {
	w := &MultiWriter{
		InfoWriter: &FileWriter{
			Filename: "file-info.log",
		},
		WarnWriter: &FileWriter{
			Filename: "file-warn.log",
		},
		ErrorWriter: &FileWriter{
			Filename: "file-error.log",
		},
	}
	logger := Logger{
		Level:        InfoLevel,
		Caller:       1,
		MultiWriters: w,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Int("id", 42).Msg("I'm loving it.")
		logger.Warn().Int("id", 43).Msg("I double dare you.")
		logger.Error().Str("action", "cleanup").Msg("World")
	}
	b.StopTimer()
	matches, _ := filepath.Glob("file-*.log")
	for i := range matches {
		err := os.Remove(matches[i])
		if err != nil {
			b.Fatalf("os remove %s error: %+v", matches[i], err)
		}
	}
}

func BenchmarkOldMultiWriter(b *testing.B) {
	w := &MultiWriter{
		InfoWriter: &FileWriter{
			Filename: "file-info.log",
		},
		WarnWriter: &FileWriter{
			Filename: "file-warn.log",
		},
		ErrorWriter: &FileWriter{
			Filename: "file-error.log",
		},
	}
	logger := Logger{
		Level:  InfoLevel,
		Caller: 1,
		Writer: w,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Int("id", 42).Strs("names", []string{"apple", "orange"}).Msg("I'm loving it.")
		logger.Warn().Int("id", 43).Msg("I double dare you.")
		logger.Error().Str("action", "cleanup").Msg("World")
	}
	b.StopTimer()
	matches, _ := filepath.Glob("file-*.log")
	for i := range matches {
		err := os.Remove(matches[i])
		if err != nil {
			b.Fatalf("os remove %s error: %+v", matches[i], err)
		}
	}
}
