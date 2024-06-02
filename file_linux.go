//go:build linux
// +build linux

package log

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

func (w *FileWriter) WriteV(iovs []syscall.Iovec) (n uintptr, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if w.Filename == "" {
			n, _, err = syscall.Syscall(syscall.SYS_WRITEV, uintptr(2), uintptr(unsafe.Pointer(&iovs[0])), uintptr(len(iovs)))
			return
		}
		if w.EnsureFolder {
			err = os.MkdirAll(filepath.Dir(w.Filename), 0755)
			if err != nil {
				return
			}
		}
		err = w.create()
		if err != nil {
			return
		}
	}

	var errno syscall.Errno
	n, _, errno = syscall.Syscall(syscall.SYS_WRITEV, uintptr(w.file.Fd()), uintptr(unsafe.Pointer(&iovs[0])), uintptr(len(iovs)))
	if errno != 0 {
		err = errno
		return
	}

	w.size += int64(n)
	if w.MaxSize > 0 && w.size > w.MaxSize && w.Filename != "" {
		err = w.rotate()
	}

	return
}
