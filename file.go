package log

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

// FileWriter is an io.WriteCloser that writes to the specified filename.
//
// FileWriter opens or creates the logfile on first Write.  If the file exists and
// is less than MaxSize megabytes, FileWriter will open and append to that file.
// If the file exists and its size is >= MaxSize megabytes, the file is renamed
// by putting the current time in a timestamp in the name immediately before the
// file's extension (or the end of the filename if there's no extension). A new
// log file is then created using original filename.
//
// Whenever a write would cause the current log file exceed MaxSize megabytes,
// the current file is closed, renamed, and a new log file created with the
// original name. Thus, the filename you give FileWriter is always the "current" log
// file.
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

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated.
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
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool

	// HostName determines if the hostname used for formatting in backup files.
	HostName bool
}

// Write implements io.FileWriter.  If a write would cause the log file to be larger
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

func (w *FileWriter) rotate() (err error) {
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

		// handle relative path, for example: `logs/main.log`, main.log --> main.2006-01-02T15-04-05.log
		_, newfile := filepath.Split(newname)
		if len(newfile) != 0 {
			newname = newfile
		}

		os.Remove(filename)
		os.Symlink(newname, filename)

		uid, _ := strconv.Atoi(os.Getenv("SUDO_UID"))
		gid, _ := strconv.Atoi(os.Getenv("SUDO_GID"))
		if uid != 0 && gid != 0 && os.Geteuid() == 0 {
			os.Lchown(filename, uid, gid)
			os.Chown(newname, uid, gid)
		}

		if names, _ := filepath.Glob(prefix + ".20*" + ext); len(names) > 0 {
			sort.Strings(names)
			for i := 0; i < len(names)-backups-1; i++ {
				os.Remove(names[i])
			}
		}
	}(file, filename, w.Filename, w.MaxBackups)

	return
}

func (w *FileWriter) create() (err error) {
	now := timeNow()
	if !w.LocalTime {
		now = now.UTC()
	}

	ext := filepath.Ext(w.Filename)
	filename := w.Filename[0:len(w.Filename)-len(ext)] + now.Format(".2006-01-02T15-04-05")
	if w.HostName {
		filename += "." + hostname + ext
	} else {
		filename += ext
	}

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
	os.Symlink(filename, w.Filename)

	return
}

// Writer is an alias for FileWriter
type Writer = FileWriter
