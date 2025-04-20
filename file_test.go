package log

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileWriter(t *testing.T) {
	const filename string = "file-output.log"
	const text string = "hello file writer!\n"

	w := &FileWriter{
		Filename: filename,
	}
	_, err := wlprintf(w, InfoLevel, text)
	if err != nil {
		t.Fatalf("file writer error: %+v", err)
	}

	// _ = w.Rotate()
	w.Close()

	matches, err := filepath.Glob("file-output.*.log")
	if err != nil {
		t.Fatalf("filepath glob error: %+v", err)
	}
	if len(matches) == 0 {
		t.Fatal("filepath glob return empty")
	}

	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read file error: %+v", err)
	}

	if string(data) != text {
		t.Fatalf("read file content mismath: data=[%s], text=[%s]", data, text)
	}

	err = os.Remove(matches[0])
	if err != nil {
		t.Fatalf("os remove %s error: %+v", matches[0], err)
	}

	os.Remove(filename)
}

func TestFileWriterStderr(t *testing.T) {
	const text1 string = "hello file writer!\n"

	w := &FileWriter{}

	_, err := wlprintf(w, InfoLevel, text1)
	if err != nil {
		t.Fatalf("file writer error: %+v", err)
	}
}

func TestFileWriterCreate(t *testing.T) {
	const text1 string = "hello file writer!\n"

	w := &FileWriter{
		Filename: "/nonexists/output.log",
	}

	_, err := wlprintf(w, InfoLevel, text1)
	if err == nil {
		t.Fatalf("file writer should not write")
	}

	t.Logf("file writer return error: %+v", err)
}

func TestFileWriterEnsureFolder(t *testing.T) {
	var remove = func(dirname string) {
		matches, _ := filepath.Glob(dirname + "/*")
		for i := range matches {
			os.Remove(matches[i])
		}
		os.Remove(dirname)
	}

	filename := "logs/file-hostname.log"
	text1 := "1. hello file writer!\n"
	text2 := "2. hello file writer!\n"
	w := &FileWriter{
		Filename:     filename,
		EnsureFolder: true,
	}

	remove(filepath.Dir(filename))

	_, err := fmt.Fprint(w, text1)
	if err != nil {
		t.Logf("file writer return error: %+v", err)
	}

	remove(filepath.Dir(filename))

	_, err = fmt.Fprint(w, text1)
	if err != nil {
		t.Logf("file writer return error: %+v", err)
	}

	err = w.Rotate()
	if err != nil {
		t.Logf("file writer rotate error: %+v", err)
	}

	_, err = fmt.Fprint(w, text2)
	if err != nil {
		t.Logf("file writer return error: %+v", err)
	}

	err = w.Close()
	if err != nil {
		t.Logf("file writer return error: %+v", err)
	}

	remove(filepath.Dir(filename))
}

func TestFileWriterHostname(t *testing.T) {
	const filename string = "file-hostname.log"
	const text1 string = "1. hello file writer!\n"
	const text2 string = "2. hello file writer!\n"

	for _, hostname := range []bool{false, true} {
		for _, pid := range []bool{false, true} {
			w := &FileWriter{
				Filename:  filename,
				HostName:  hostname,
				ProcessID: pid,
			}

			_, err := wlprintf(w, InfoLevel, text1)
			if err != nil {
				t.Logf("file writer return error: %+v", err)
			}

			time.Sleep(time.Second)
			os.Setenv("USER", "root")
			_ = w.Rotate()
			w.Close()

			_, err = wlprintf(w, InfoLevel, text2)
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
	const filename string = "file-rotate.log"
	const header string = "# I AM A FILEWRITER HEADER\n"
	const text1 string = "hello file writer!\n"
	const text2 string = "hello rotated file writer!\n"

	// trigger chown
	os.Setenv("USER", "root")

	w := &FileWriter{
		Filename:   filename,
		MaxBackups: 2,
		Header: func(_ os.FileInfo) []byte {
			return []byte(header)
		},
	}

	// text 1
	_, err := wlprintf(w, InfoLevel, text1)
	if err != nil {
		t.Fatalf("file writer error: %+v", err)
	}

	time.Sleep(time.Second)
	_ = w.Rotate()

	// text 2
	_, err = wlprintf(w, InfoLevel, text2)
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

	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read file error: %+v", err)
	}

	if string(data) != header+text1 {
		t.Fatalf("read file content mismath: data=[%s], text1=[%s]", data, text1)
	}

	data, err = os.ReadFile(matches[1])
	if err != nil {
		t.Fatalf("read file error: %+v", err)
	}

	if string(data) != header+text2 {
		t.Fatalf("read file content mismath: data=[%s], text2=[%s]", data, text2)
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
	const filename string = "file-rotate-by-size.log"
	const text string = "hello file writer!\n"

	w := &FileWriter{
		Filename:   filename,
		MaxSize:    int64(len(text)) + 2,
		MaxBackups: 2,
	}

	// text 1
	_, err := wlprintf(w, InfoLevel, text)
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
	_, err = wlprintf(w, InfoLevel, text)
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
		_, err = wlprintf(w, InfoLevel, text)
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
	const filename string = "file-backup.log"

	w := &FileWriter{
		Filename:   filename,
		MaxBackups: 1,
	}

	time.Sleep(time.Second)
	_ = w.Rotate()

	time.Sleep(time.Second)
	_ = w.Rotate()
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

func TestFileWriterFileargs(t *testing.T) {
	filename := "file-output.log"
	d := time.Date(2020, 8, 12, 16, 7, 0, 0, time.UTC)

	t.Run("neither hostname nor pid appears", func(t *testing.T) {
		w := &FileWriter{Filename: filename}
		expected := "file-output.2020-08-12T16-07-00.log"
		if name, _, _ := w.fileargs(d); name != expected {
			t.Fatalf("expected: %q, actual: %q", expected, name)
		}
	})
	t.Run("hostname or pid appears", func(t *testing.T) {
		origHost := hostname
		hostname = "shire"
		defer func() { hostname = origHost }()
		origPid := pid
		pid = 198400
		defer func() { pid = origPid }()

		w := &FileWriter{Filename: filename, HostName: true}

		cases := []struct {
			hostName  bool
			processID bool
			expected  string
		}{
			{hostName: true, expected: "file-output.2020-08-12T16-07-00.shire.log"},
			{processID: true, expected: "file-output.2020-08-12T16-07-00.198400.log"},
			{hostName: true, processID: true, expected: "file-output.2020-08-12T16-07-00.shire-198400.log"},
		}
		for _, c := range cases {
			w.HostName = c.hostName
			w.ProcessID = c.processID
			if name, _, _ := w.fileargs(d); name != c.expected {
				t.Fatalf("expected: %q, actual: %q", c.expected, name)
			}
		}
	})
}

func TestFileWriter_MaxSizeRotation_SingleVsMultiInstance(t *testing.T) {
	tempDir := t.TempDir()
	baseFilename := filepath.Join(tempDir, "rotation-test.log")

	type testCase struct {
		name           string
		timeFormat     string
		hostName       bool
		processID      bool
		expectFiles    int
		allowOversize  bool
		singleInstance bool
	}

	testCases := []testCase{
		{"_DayPrecision_MultiInstance", "2006-01-02", false, false, 1, true, false},
		{"_DayPrecision_SingleInstance", "2006-01-02", false, false, 1, true, true},
		{"_MsPrecision_MultiInstance", "2006-01-02T15-04-05.000", false, false, 4, false, false},
		{"_MsPrecision_SingleInstance", "2006-01-02T15-04-05", false, false, 1, true, true},
		{"HostName_DayPrecision_MultiInstance", "2006-01-02", true, false, 1, true, false},
		{"HostName_DayPrecision_SingleInstance", "2006-01-02", true, false, 1, true, true},
		{"HostName_MsPrecision_MultiInstance", "2006-01-02T15-04-05.000", true, false, 4, false, false},
		{"HostName_MsPrecision_SingleInstance", "2006-01-02T15-04-05", true, false, 1, true, true},
		{"ProcessID_DayPrecision_MultiInstance", "2006-01-02", false, true, 1, true, false},
		{"ProcessID_DayPrecision_SingleInstance", "2006-01-02", false, true, 1, true, true},
		{"ProcessID_MsPrecision_MultiInstance", "2006-01-02T15-04-05.000", false, true, 4, false, false},
		{"ProcessID_MsPrecision_SingleInstance", "2006-01-02T15-04-05", false, true, 1, true, true},
		{"HostNameProcessID_DayPrecision_MultiInstance", "2006-01-02", true, true, 1, true, false},
		{"HostNameProcessID_DayPrecision_SingleInstance", "2006-01-02", true, true, 1, true, true},
		{"HostNameProcessID_MsPrecision_MultiInstance", "2006-01-02T15-04-05.000", true, true, 4, false, false},
		{"HostNameProcessID_MsPrecision_SingleInstance", "2006-01-02T15-04-05", true, true, 1, true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Remove(baseFilename)
			matches, _ := filepath.Glob(filepath.Join(tempDir, "rotation-test.*"))
			for _, m := range matches {
				os.Remove(m)
			}

			maxSize := int64(1024)
			maxBackups := 3
			dataChunkSize := 500
			iterations := 10
			dataChunk := make([]byte, dataChunkSize)
			for i := range dataChunk {
				dataChunk[i] = byte('A' + (i % 26))
			}
			dataChunk[dataChunkSize-1] = '\n'

			if tc.singleInstance {
				w := &FileWriter{
					Filename:   baseFilename,
					MaxSize:    maxSize,
					MaxBackups: maxBackups,
					TimeFormat: tc.timeFormat,
					HostName:   tc.hostName,
					ProcessID:  tc.processID,
				}
				for i := 0; i < iterations; i++ {
					_, err := w.Write(dataChunk)
					if err != nil {
						t.Fatalf("iteration %d: file write error: %+v", i, err)
					}
				}
				err := w.Close()
				if err != nil {
					t.Fatalf("error closing writer: %+v", err)
				}
			} else {
				for i := 0; i < iterations; i++ {
					w := &FileWriter{
						Filename:   baseFilename,
						MaxSize:    maxSize,
						MaxBackups: maxBackups,
						TimeFormat: tc.timeFormat,
						HostName:   tc.hostName,
						ProcessID:  tc.processID,
					}
					_, err := w.Write(dataChunk)
					if err != nil {
						t.Fatalf("iteration %d: file write error: %+v", i, err)
					}
					err = w.Close()
					if err != nil {
						t.Fatalf("iteration %d: error closing writer: %+v", i, err)
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
			time.Sleep(100 * time.Millisecond)

			pattern := filepath.Join(tempDir, "rotation-test*")
			matches, err := filepath.Glob(pattern)
			if err != nil {
				t.Fatalf("filepath glob error: %+v", err)
			}

			// Discard possible symlink (main log file symlink)
			filtered := matches[:0]
			for _, m := range matches {
				info, err := os.Lstat(m)
				if err == nil && (info.Mode()&os.ModeSymlink == 0) {
					filtered = append(filtered, m)
				}
			}
			matches = filtered

			if len(matches) != tc.expectFiles {
				t.Errorf("Expected %d log files, found %d. Files: %v", tc.expectFiles, len(matches), matches)
			}

			largestFileSize := int64(0)
			var largestFileName string
			for _, m := range matches {
				info, err := os.Stat(m)
				if err == nil && info.Size() > largestFileSize {
					largestFileSize = info.Size()
					largestFileName = filepath.Base(m)
				}
			}
			if !tc.allowOversize && largestFileSize > maxSize {
				t.Errorf("File %q size (%d bytes) exceeds MaxSize (%d bytes)", largestFileName, largestFileSize, maxSize)
			}
			if tc.allowOversize && largestFileSize <= maxSize {
				t.Errorf("Expected at least one oversized file, but largest file is %d bytes", largestFileSize)
			}
		})
	}
}
