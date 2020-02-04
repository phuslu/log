package log

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
)

var _ io.WriteCloser = (*Writer)(nil)

var timeNow = time.Now

var hostname, _ = os.Hostname()

type Writer struct {
	size int64
	file *os.File
	mu   sync.Mutex

	Filename   string
	MaxSize    int64
	MaxBackups int
	LocalTime  bool
	HostName   bool
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()

	if w.file == nil {
		err = w.create()
		if err != nil {
			w.mu.Unlock()
			return
		}
	}

	n, err = w.file.Write(p)
	if err != nil {
		return
	}

	w.size += int64(n)
	if w.MaxSize > 0 && w.size > w.MaxSize {
		err = w.rotate()
	}

	w.mu.Unlock()
	return
}

func (w *Writer) Close() (err error) {
	w.mu.Lock()
	if w.file != nil {
		err = w.file.Close()
		w.file = nil
		w.size = 0
	}
	w.mu.Unlock()
	return
}

func (w *Writer) Rotate() (err error) {
	w.mu.Lock()
	err = w.rotate()
	w.mu.Unlock()

	return
}

func (w *Writer) rotate() (err error) {
	if w.Filename == "" {
		w.file = os.Stderr
		return
	}

	if w.file != nil {
		err = w.file.Close()
		if err != nil {
			return
		}
	}

	now := timeNow()
	if !w.LocalTime {
		now = now.UTC()
	}

	ext := filepath.Ext(w.Filename)
	prefix := w.Filename[0 : len(w.Filename)-len(ext)]
	filename := prefix + now.Format(".2006-01-02T15-04-05")
	if w.HostName {
		filename += "." + hostname + ext
	} else {
		filename += ext
	}

	w.file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	w.size = 0

	go func(filename string) {
		os.Remove(w.Filename)
		err := os.Symlink(filename, w.Filename)
		if err != nil {
			return
		}

		switch runtime.GOOS {
		case "linux":
			uid, _ := strconv.Atoi(os.Getenv("SUDO_UID"))
			gid, _ := strconv.Atoi(os.Getenv("SUDO_GID"))
			if uid != 0 && gid != 0 && os.Geteuid() == 0 {
				os.Lchown(w.Filename, uid, gid)
				os.Chown(filename, uid, gid)
			}
		}

		var matches []string
		matches, err = filepath.Glob(prefix + ".20*" + ext)
		if err != nil {
			return
		}

		sort.Strings(matches)
		for i := 0; i < len(matches)-w.MaxBackups-1; i++ {
			os.Remove(matches[i])
		}
	}(filename)

	return
}

func (w *Writer) create() (err error) {
	var filename string

	if link, err := os.Readlink(w.Filename); err == nil {
		if fi, err := os.Stat(link); err == nil {
			filename = link
			w.size = fi.Size()
		}
	}

	if filename == "" {
		now := timeNow()
		if !w.LocalTime {
			now = now.UTC()
		}
		ext := filepath.Ext(w.Filename)
		filename = w.Filename[0 : len(w.Filename)-len(ext)]
		filename += now.Format(".2006-01-02T15-04-05")
		if w.HostName {
			filename += "." + hostname + ext
		} else {
			filename += ext
		}
		w.size = 0
	}

	w.file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	os.Remove(w.Filename)
	err = os.Symlink(filename, w.Filename)

	return
}
