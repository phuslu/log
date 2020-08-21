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

func TestMultiWriter(t *testing.T) {
	w := &MultiWriter{
		InfoWriter:   &FileWriter{Filename: "file-info.log"},
		WarnWriter:   &FileWriter{Filename: "file-warn.log"},
		ErrorWriter:  &FileWriter{Filename: "file-error.log"},
		StderrWriter: &ConsoleWriter{ColorOutput: true},
		StderrLevel:  ErrorLevel,
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := fmt.Fprintf(w, `{"ts":1234567890,"level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
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
	err = errorWriterOK
	return
}

func (ew errorWriter) Close() (err error) {
	ew.WriteCloser.Close()
	err = errorWriterOK
	return
}

func TestMultiWriterError(t *testing.T) {
	file, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
	if err != nil {
		t.Errorf("open null file error: %+v", err)
	}

	w := &MultiWriter{
		InfoWriter:   file,
		WarnWriter:   file,
		ErrorWriter:  file,
		StderrWriter: &ConsoleWriter{ColorOutput: true},
		StderrLevel:  TraceLevel,
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json error writer error: %+v", err)
		}
	}

	if err := w.Close(); err == nil {
		t.Errorf("test close error writer error: %+v", err)
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

	w.ParseLevel = func(p []byte) Level { return ParseLevelByte(p[49]) }
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
