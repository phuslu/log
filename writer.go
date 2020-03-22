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

// Writer is an io.WriteCloser that writes to the specified filename.
//
// Writer opens or creates the logfile on first Write.  If the file exists and
// is less than MaxSize megabytes, Writer will open and append to that file.
// If the file exists and its size is >= MaxSize megabytes, the file is renamed
// by putting the current time in a timestamp in the name immediately before the
// file's extension (or the end of the filename if there's no extension). A new
// log file is then created using original filename.
//
// Whenever a write would cause the current log file exceed MaxSize megabytes,
// the current file is closed, renamed, and a new log file created with the
// original name. Thus, the filename you give Writer is always the "current" log
// file.
//
// Backups use the log file name given to Writer, in the form
// `name.timestamp.ext` where name is the filename without the extension,
// timestamp is the time at which the log was rotated formatted with the
// time.Time format of `2006-01-02T15-04-05` and the extension is the
// original extension.  For example, if your Writer.Filename is
// `/var/log/foo/server.log`, a backup created at 6:30pm on Nov 11 2016 would
// use the filename `/var/log/foo/server.2016-11-04T18-30-00.log`
//
// Cleaning Up Old Log Files
//
// Whenever a new logfile gets created, old log files may be deleted.  The most
// recent files according to the encoded timestamp will be retained, up to a
// number equal to MaxBackups (or all of them if MaxBackups is 0).  Any files
// with an encoded timestamp older than MaxAge days are deleted, regardless of
// MaxBackups.  Note that the time encoded in the timestamp is the rotation
// time, which may differ from the last time that file was written to.
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

// Write implements io.Writer.  If a write would cause the log file to be larger
// than MaxSize, the file is closed, renamed to include a timestamp of the
// current time, and a new log file is created using the original log file name.
// If the length of the write is greater than MaxSize, an error is returned.
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

// Close implements io.Closer, and closes the current logfile.
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

// Rotate causes Logger to close the existing log file and immediately create a
// new one.  This is a helper function for applications that want to initiate
// rotations outside of the normal rotation rules, such as in response to
// SIGHUP.  After rotating, this initiates compression and removal of old log
// files according to the configuration.
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
	if w.Filename == "" {
		w.file = os.Stderr
		return
	}

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
