package log

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type FastFileWriter struct {
	size int64
	file atomic.Value // *os.File
	mu   sync.Mutex

	Filename string

	MaxSize int64

	MaxBackups int

	FileMode os.FileMode

	LocalTime bool

	HostName bool

	ProcessID bool
}

func (w *FastFileWriter) Write(p []byte) (n int, err error) {
	file := w.file.Load()

	if file != nil {
		n, err = file.(*os.File).Write(p)
		// see https://github.com/golang/go/issues/7970
		if e, ok := err.(*os.PathError); ok && e.Err == os.ErrClosed {
			n, err = w.file.Load().(*os.File).Write(p)
		}
	} else {
		w.mu.Lock()
		// double check
		file = w.file.Load()
		if file == nil {
			err = w.create()
		}
		w.mu.Unlock()
		if err != nil {
			return
		}
		n, err = w.file.Load().(*os.File).Write(p)
	}

	if w.MaxSize > 0 && atomic.AddInt64(&w.size, int64(n)) > w.MaxSize {
		w.mu.Lock()
		if atomic.LoadInt64(&w.size) > w.MaxSize {
			w.rotate()
		}
		w.mu.Unlock()
	}

	return
}

func (w *FastFileWriter) Close() (err error) {
	w.mu.Lock()

	file := w.file.Load()
	if file != nil {
		err = file.(*os.File).Close()
		atomic.StoreInt64(&w.size, 0)
	}

	w.mu.Unlock()
	return
}

func (w *FastFileWriter) Rotate() (err error) {
	w.mu.Lock()
	err = w.rotate()
	w.mu.Unlock()
	return
}

func (w *FastFileWriter) rotate() error {
	oldfile := w.file.Load()

	file, err := os.OpenFile(w.fileinfo(timeNow()))
	if err != nil {
		return err
	}
	w.file.Store(file)
	atomic.StoreInt64(&w.size, 0)

	go func(oldfile interface{}, newname, filename string, backups int) {
		if oldfile != nil {
			oldfile.(*os.File).Close()
		}

		os.Remove(filename)
		os.Symlink(filepath.Base(newname), filename)

		uid, _ := strconv.Atoi(os.Getenv("SUDO_UID"))
		gid, _ := strconv.Atoi(os.Getenv("SUDO_GID"))
		if uid != 0 && gid != 0 && os.Geteuid() == 0 {
			os.Lchown(filename, uid, gid)
			os.Chown(newname, uid, gid)
		}

		ext := filepath.Ext(filename)
		pattern := filename[0:len(filename)-len(ext)] + ".20*" + ext
		if names, _ := filepath.Glob(pattern); len(names) > 0 {
			sort.Strings(names)
			for i := 0; i < len(names)-backups-1; i++ {
				os.Remove(names[i])
			}
		}
	}(oldfile, file.Name(), w.Filename, w.MaxBackups)

	return nil
}

func (w *FastFileWriter) create() error {
	file, err := os.OpenFile(w.fileinfo(timeNow()))
	if err != nil {
		return err
	}
	w.file.Store(file)
	atomic.StoreInt64(&w.size, 0)

	os.Remove(w.Filename)
	os.Symlink(filepath.Base(file.Name()), w.Filename)

	return nil
}

func (w *FastFileWriter) fileinfo(now time.Time) (filename string, flag int, perm os.FileMode) {
	if !w.LocalTime {
		now = now.UTC()
	}

	ext := filepath.Ext(w.Filename)
	prefix := w.Filename[0 : len(w.Filename)-len(ext)]
	filename = prefix + now.Format(".2006-01-02T15-04-05")
	if w.HostName {
		if w.ProcessID {
			filename += "." + hostname + "-" + pid + ext
		} else {
			filename += "." + hostname + ext
		}
	} else {
		if w.ProcessID {
			filename += "." + pid + ext
		} else {
			filename += ext
		}
	}

	flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	perm = w.FileMode
	if perm == 0 {
		perm = 0644
	}

	return
}
