package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var timeNow = time.Now

var _ io.WriteCloser = (*Writer)(nil)

var hostname, _ = os.Hostname()

type Writer struct {
	size int64

	mu   sync.Mutex
	file *os.File

	Filename   string
	MaxSize    int64
	MaxBackups int
	LocalTime  bool
	HostName   bool
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.file == nil {
		w.mu.Lock()
		if w.file == nil {
			err = w.rotate(false)
		}
		w.mu.Unlock()
		if err != nil {
			return
		}
	}

	n, err = w.file.Write(p)

	if w.MaxSize > 0 && atomic.AddInt64(&w.size, int64(n)) > w.MaxSize {
		w.mu.Lock()
		if atomic.LoadInt64(&w.size) > w.MaxSize {
			w.rotate(true)
		}
		w.mu.Unlock()
	}

	return
}

func (w *Writer) Close() (err error) {
	w.mu.Lock()

	err = w.file.Close()
	w.file = nil
	atomic.StoreInt64(&w.size, 0)

	w.mu.Unlock()
	return
}

func (w *Writer) Rotate() (err error) {
	w.mu.Lock()
	err = w.rotate(true)
	w.mu.Unlock()
	return
}

func (w *Writer) rotate(newFile bool) (err error) {
	if w.Filename == "" {
		w.file = os.Stderr
		return nil
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
	atomic.StoreInt64(&w.size, 0)

	if !newFile {
		if link, err := os.Readlink(w.Filename); err == nil {
			if fi, err := os.Stat(link); err == nil {
				filename = link
				atomic.StoreInt64(&w.size, fi.Size())
			}
		}
	}

	w.file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil && w.file == nil {
		return
	}

	go func(filename string) {
		os.Remove(w.Filename)
		err := os.Symlink(filename, w.Filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "symlink %+v to %+v error: %+v", filename, w.Filename, err)
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

var _ io.WriteCloser = (*BufferWriter)(nil)

type BufferWriter struct {
	BufferSize    int
	FlushDuration time.Duration
	*Writer

	mu  sync.Mutex
	buf []byte
}

func (w *BufferWriter) Flush() (err error) {
	w.mu.Lock()
	if len(w.buf) != 0 {
		_, err = w.Writer.Write(w.buf)
		w.buf = w.buf[:0]
	}
	w.mu.Unlock()
	return
}

func (w *BufferWriter) Close() error {
	w.mu.Lock()
	w.Writer.Write(w.buf)
	w.buf = w.buf[:0]
	w.mu.Unlock()
	return w.Writer.Close()
}

func (w *BufferWriter) Write(p []byte) (n int, err error) {
	if w.buf == nil {
		w.Writer.mu.Lock()
		if w.BufferSize == 0 {
			w.BufferSize = 32 * 1024
		}
		if w.FlushDuration == 0 {
			w.FlushDuration = 5 * time.Second
		}
		if w.buf == nil {
			w.buf = make([]byte, 0, w.BufferSize+1024)
		}
		go func() {
			for {
				time.Sleep(w.FlushDuration)
				if len(w.buf) != 0 {
					w.Flush()
				}
			}
		}()
		w.Writer.mu.Unlock()
	}

	w.mu.Lock()
	w.buf = append(w.buf, p...)
	n = len(p)
	if len(w.buf) > w.BufferSize {
		_, err = w.Writer.Write(w.buf)
		w.buf = w.buf[:0]
	}
	w.mu.Unlock()

	return
}
