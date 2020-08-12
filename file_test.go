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

func TestFileWriterStderr(t *testing.T) {
	text1 := "hello file writer!\n"

	w := &FileWriter{}

	_, err := fmt.Fprintf(w, text1)
	if err != nil {
		t.Fatalf("file writer error: %+v", err)
	}
}

func TestFileWriterCreate(t *testing.T) {
	text1 := "hello file writer!\n"

	w := &FileWriter{
		Filename: "/nonexists/output.log",
	}

	_, err := fmt.Fprintf(w, text1)
	if err == nil {
		t.Fatalf("file writer should not write")
	}

	t.Logf("file writer return error: %+v", err)
}

func TestFileWriterHostname(t *testing.T) {
	filename := "file-hostname.log"
	text1 := "1. hello file writer!\n"
	text2 := "2. hello file writer!\n"

	for _, hostname := range []bool{false, true} {
		for _, pid := range []bool{false, true} {
			w := &FileWriter{
				Filename:  filename,
				HostName:  hostname,
				ProcessID: pid,
			}

			_, err := fmt.Fprintf(w, text1)
			if err != nil {
				t.Logf("file writer return error: %+v", err)
			}

			time.Sleep(time.Second)
			os.Setenv("USER", "root")
			w.Rotate()
			w.Close()

			_, err = fmt.Fprintf(w, text2)
			if err != nil {
				t.Logf("file writer return error: %+v", err)
			}

			w.Close()

			matches, _ := filepath.Glob("file-hostname.*.log")
			for i := range matches {
				os.Remove(matches[i])
			}

			os.Remove(filename)
		}
	}
}

func TestFileWriterRotate(t *testing.T) {
	filename := "file-rotate.log"
	text1 := "hello file writer!\n"
	text2 := "hello rotated file writer!\n"

	// trigger chown
	os.Setenv("USER", "root")

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

func TestFileWriterRotateBySize(t *testing.T) {
	filename := "file-rotate-by-size.log"
	text := "hello file writer!\n"

	w := &FileWriter{
		Filename:   filename,
		MaxSize:    int64(len(text)) + 2,
		MaxBackups: 2,
	}

	// text 1
	_, err := fmt.Fprintf(w, text)
	if err != nil {
		t.Fatalf("file writer error: %+v", err)
	}

	matches, err := filepath.Glob("file-rotate-by-size.*.log")
	if err != nil {
		t.Fatalf("filepath glob error: %+v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("filepath glob return %+v number mismath", matches)
	}

	time.Sleep(time.Second)

	// text 2
	_, err = fmt.Fprintf(w, text)
	if err != nil {
		t.Fatalf("file writer error: %+v", err)
	}

	matches, err = filepath.Glob("file-rotate-by-size.*.log")
	if err != nil {
		t.Fatalf("filepath glob error: %+v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("filepath glob return %+v number mismath", matches)
	}

	// mock
	os.Setenv("SUDO_UID", "1000")
	os.Setenv("SUDO_GID", "1000")

	// text 3 ~ 6
	for i := 3; i <= 6; i++ {
		_, err = fmt.Fprintf(w, text)
		time.Sleep(time.Second)
		if err != nil {
			t.Fatalf("file writer error: %+v", err)
		}
	}

	matches, err = filepath.Glob("file-rotate-by-size.*.log")
	if err != nil {
		t.Fatalf("filepath glob error: %+v", err)
	}
	if len(matches) > w.MaxBackups+1 {
		t.Fatalf("filepath glob return %+v number mismath", matches)
	}

	w.Close()

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

func TestGetFileName(t *testing.T) {
	filename := "file-output.log"
	d := time.Date(2020, 8, 12, 16, 7, 0, 0, time.UTC)

	t.Run("neither hostname nor pid appears", func(t *testing.T) {
		w := &FileWriter{Filename: filename}
		expected := "file-output.2020-08-12T16-07-00.log"
		if name := w.getFileName(d); name != expected {
			t.Fatalf("expected: %q, actual: %q", expected, name)
		}
	})
	t.Run("hostname or pid appears", func(t *testing.T) {
		origHost := hostname
		hostname = "shire"
		defer func() { hostname = origHost }()
		origPid := pid
		pid = "198400"
		defer func() { pid = origPid }()

		w := &FileWriter{Filename: filename, HostName: true}

		cases := []struct{
			hostName bool
			processID bool
			expected string
		}{
			{hostName: true, expected: "file-output.2020-08-12T16-07-00.shire.log"},
			{processID: true, expected: "file-output.2020-08-12T16-07-00-198400.log"},
			{hostName: true, processID: true, expected: "file-output.2020-08-12T16-07-00.shire-198400.log"},
		}
		for _, c := range cases {
			w.HostName = c.hostName
			w.ProcessID = c.processID
			if name := w.getFileName(d); name != c.expected {
				t.Fatalf("expected: %q, actual: %q", c.expected, name)
			}
		}
	})
}
