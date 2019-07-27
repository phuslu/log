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
	"unsafe"
)

var _ io.WriteCloser = (*Writer)(nil)

var timeNow = time.Now

var hostname, _ = os.Hostname()

type Writer struct {
	size int64
	file unsafe.Pointer // *os.File

	mu sync.Mutex

	Filename   string
	MaxSize    int64
	MaxBackups int
	LocalTime  bool
	HostName   bool
}

func (w *Writer) File() *os.File {
	return (*os.File)(atomic.LoadPointer(&w.file))
}

func (w *Writer) Close() (err error) {
	w.mu.Lock()

	err = w.File().Close()
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
		atomic.StorePointer(&w.file, unsafe.Pointer(os.Stderr))
		return nil
	}

	if file := w.File(); file != nil {
		err = file.Close()
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

	var file *os.File
	file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil && file == nil {
		return
	}
	atomic.StorePointer(&w.file, unsafe.Pointer(file))

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
		matches, err = filepath.Glob(prefix + ".20[0-9][0-9]*" + ext)
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
