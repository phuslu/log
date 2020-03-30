package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileWriter(t *testing.T) {
	filename := "file-output.log"
	text := "hello file writer!\n"

	w := &FileWriter{
		Filename: filename,
	}
	_, err := fmt.Fprintf(w, text)
	if err != nil {
		t.Fatalf("file writer error: %+v", err)
	}

	// w.Rotate()
	w.Close()

	matches, err := filepath.Glob("file-output.*.log")
	if err != nil {
		t.Fatalf("filepath glob error: %+v", err)
	}
	if len(matches) == 0 {
		t.Fatal("filepath glob return empty")
	}

	data, err := ioutil.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("ioutil read file error: %+v", err)
	}

	if string(data) != text {
		t.Fatalf("ioutil read file content mismath: data=[%s], text=[%s]", data, text)
	}

	err = os.Remove(matches[0])
	if err != nil {
		t.Fatalf("os remove %s error: %+v", matches[0], err)
	}

	os.Remove(filename)
}

func TestFileWriterRotate(t *testing.T) {
	filename := "file-rotate.log"
	text1 := "hello file writer!\n"
	text2 := "hello rotated file writer!\n"

	w := &FileWriter{
		Filename:   filename,
		MaxBackups: 2,
	}

	// text 1
	_, err := fmt.Fprintf(w, text1)
	if err != nil {
		t.Fatalf("file writer error: %+v", err)
	}

	time.Sleep(time.Second)
	w.Rotate()

	// text 2
	_, err = fmt.Fprintf(w, text2)
	if err != nil {
		t.Fatalf("file writer error: %+v", err)
	}

	w.Close()

	matches, err := filepath.Glob("file-rotate.*.log")
	if err != nil {
		t.Fatalf("filepath glob error: %+v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("filepath glob return %+v number mismath", matches)
	}

	data, err := ioutil.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("ioutil read file error: %+v", err)
	}

	if string(data) != text1 {
		t.Fatalf("ioutil read file content mismath: data=[%s], text1=[%s]", data, text1)
	}

	data, err = ioutil.ReadFile(matches[1])
	if err != nil {
		t.Fatalf("ioutil read file error: %+v", err)
	}

	if string(data) != text2 {
		t.Fatalf("ioutil read file content mismath: data=[%s], text2=[%s]", data, text2)
	}

	for i := range matches {
		err = os.Remove(matches[i])
		if err != nil {
			t.Fatalf("os remove %s error: %+v", matches[i], err)
		}
	}

	os.Remove(filename)
}

func TestFileWriterBackups(t *testing.T) {
	filename := "file-backup.log"

	w := &FileWriter{
		Filename:   filename,
		MaxBackups: 1,
	}

	time.Sleep(time.Second)
	w.Rotate()

	time.Sleep(time.Second)
	w.Rotate()
	w.Close()

	matches, err := filepath.Glob("file-backup.*.log")
	if err != nil {
		t.Fatalf("filepath glob error: %+v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("filepath glob return %+v number mismath", matches)
	}

	matches, _ = filepath.Glob("file-backup.*.log")
	for i := range matches {
		err = os.Remove(matches[i])
		if err != nil {
			t.Fatalf("os remove %s error: %+v", matches[i], err)
		}
	}

	os.Remove(filename)
}
