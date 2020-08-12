package log

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FileWriter is an io.WriteCloser that writes to the specified filename.
//
// Backups use the log file name given to FileWriter, in the form
// `name.timestamp.ext` where name is the filename without the extension,
// timestamp is the time at which the log was rotated formatted with the
// time.Time format of `2006-01-02T15-04-05` and the extension is the
// original extension.  For example, if your FileWriter.Filename is
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
type FileWriter struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.
	Filename string

	// MaxSize is the maximum size in bytes of the log file before it gets rotated.
	MaxSize int64

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files
	MaxBackups int

	// make aligncheck happy
	mu   sync.Mutex
	size int64
	file *os.File

	// FileMode represents the file's mode and permission bits.  The default
	// mode is 0644
	FileMode os.FileMode

	// LocalTime determines if the time used for formatting the timestamps in
	// log files is the computer's local time.  The default is to use UTC time.
	LocalTime bool

	// HostName determines if the hostname used for formatting in log files.
	HostName bool

	// ProcessID determines if the pid used for formatting in log files.
	ProcessID bool
}

// Write implements io.Writer.  If a write would cause the log file to be larger
// than MaxSize, the file is closed, renamed to include a timestamp of the
// current time, and a new log file is created using the original log file name.
// If the length of the write is greater than MaxSize, an error is returned.
func (w *FileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()

	if w.file == nil {
		if w.Filename == "" {
			n, err = os.Stderr.Write(p)
			w.mu.Unlock()
			return
		}
		err = w.create()
		if err != nil {
			w.mu.Unlock()
			return
		}
	}

	n, err = w.file.Write(p)
	if err != nil {
		w.mu.Unlock()
		return
	}

	w.size += int64(n)
	if w.MaxSize > 0 && w.size > w.MaxSize && w.Filename != "" {
		err = w.rotate()
	}

	w.mu.Unlock()
	return
}

// Close implements io.Closer, and closes the current logfile.
func (w *FileWriter) Close() (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		err = w.file.Close()
		w.file = nil
		w.size = 0
	}
	return
}

// Rotate causes Logger to close the existing log file and immediately create a
// new one.  This is a helper function for applications that want to initiate
// rotations outside of the normal rotation rules, such as in response to
// SIGHUP.  After rotating, this initiates compression and removal of old log
// files according to the configuration.
func (w *FileWriter) Rotate() (err error) {
	w.mu.Lock()
	err = w.rotate()
	w.mu.Unlock()
	return
}

var hostname = func() string {
	s, _ := os.Hostname()
	if strings.HasPrefix(s, "localhost") {
		sec, nsec := walltime()
		ts := sec*1000000 + int64(nsec)/1000
		s = "localhost-" + strconv.FormatInt(ts, 10)
	}
	return s
}()

var pid = strconv.Itoa(os.Getpid())

func (w *FileWriter) rotate() (err error) {
	now := timeNow()
	if !w.LocalTime {
		now = now.UTC()
	}

	filename := w.getFileName(now)

	perm := w.FileMode
	if perm == 0 {
		perm = 0644
	}

	file := w.file
	w.file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)
	w.size = 0

	go func(oldfile *os.File, newname, filename string, backups int) {
		if oldfile != nil {
			oldfile.Close()
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
	}(file, filename, w.Filename, w.MaxBackups)

	return
}

// getFileName returns a new file name based on the original name and the given time.
func (w *FileWriter) getFileName(now time.Time) string {
	ext := filepath.Ext(w.Filename)
	prefix := w.Filename[0 : len(w.Filename)-len(ext)]
	filename := prefix + now.Format(".2006-01-02T15-04-05")
	var hostnamePart, pidPart string
	if w.HostName {
		hostnamePart = "." + hostname
	}
	if w.ProcessID {
		pidPart = "-" + pid
	}
	filename += hostnamePart + pidPart + ext
	return filename
}

func (w *FileWriter) create() (err error) {
	now := timeNow()
	if !w.LocalTime {
		now = now.UTC()
	}

	filename := w.getFileName(now)

	perm := w.FileMode
	if perm == 0 {
		perm = 0644
	}

	w.file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	w.size = 0

	os.Remove(w.Filename)
	os.Symlink(filepath.Base(filename), w.Filename)

	return
}

// Writer is an alias for FileWriter
//
// Deprecated: Use FileWriter instead.
type Writer = FileWriter
