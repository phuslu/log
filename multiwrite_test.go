package log

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestTypeMultiWriterEntry(t *testing.T) {
	w := &MultiWriterEntry{
		&FileWriter{Filename: "file-1.log"},
		&FileWriter{Filename: "file-2.log"},
		&ConsoleWriter{ColorOutput: true},
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := wlprintf(w, ParseLevel(level), `{"ts":1234567890,"level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json mutli writer error: %+v", err)
		}
		_, err = wlprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json mutli writer error: %+v", err)
		}
		_, err = wlprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277+08:00","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
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

type errorWriterEntry struct {
	io.WriteCloser
}

var errorWriterEntryOK = errors.New("errorWriterEntry return OK")

func (ew errorWriterEntry) Write(p []byte) (n int, err error) {
	n, _ = ew.WriteCloser.Write(p)
	err = errorWriterEntryOK
	return
}

func (ew errorWriterEntry) Close() (err error) {
	ew.WriteCloser.Close()
	err = errorWriterEntryOK
	return
}

func TestMultiWriterEntryError(t *testing.T) {
	file, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
	if err != nil {
		t.Errorf("open null file error: %+v", err)
	}

	w := &MultiWriterEntry{
		IOWriter{errorWriterEntry{file}},
		IOWriter{errorWriterEntry{file}},
		&ConsoleWriter{ColorOutput: true},
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := wlprintf(w, ParseLevel(level), `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json mutli writer"}`+"\n", level)
		if err == nil {
			t.Errorf("test json error writer error: %+v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Errorf("test close error writer error: %+v", err)
	}
}
