// +build linux darwin freebsd openbsd netbsd dragonfly

package log

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// writev implements io.Writer.
func (w *AsyncWriter) writev(p []byte) (n int, err error) {
	w.once.Do(func() {
		if _, ok := w.Writer.(*FileWriter); !ok {
			panic("async writev requires a *log.FileWriter")
		}
		if w.BufferSize == 0 {
			w.BufferSize = 32 * 1024
		}
		if w.ChannelSize == 0 {
			w.ChannelSize = 100
		}
		if w.SyncDuration == 0 {
			w.SyncDuration = 5 * time.Second
		}
		// channels
		w.ch = make(chan []byte, w.ChannelSize)
		w.chDone = make(chan error)
		w.sync = make(chan struct{})
		w.syncDone = make(chan error)
		// data routine
		go func(w *AsyncWriter, fw *FileWriter) {
			var err error
			vec := make([]syscall.Iovec, 0, w.BufferSize/1024+1)
			buf := make([][]byte, 0, w.BufferSize/1024+1)
			bufsz := 0
			ticker := time.NewTicker(w.SyncDuration)
			for {
				select {
				case b := <-w.ch:
					isNil := b == nil
					if len(b) != 0 {
						vec = append(vec, syscall.Iovec{&b[0], uint64(len(b))})
						buf = append(buf, b)
						bufsz += len(b)
					}
					// full or closed
					if bufsz >= w.BufferSize || (isNil && bufsz != 0) {
						_, err = writevFileWriter(fw, vec)
						_ = buf[len(buf)-1]
						for i := 0; i < len(buf); i++ {
							b1kpool.Put(buf[i])
						}
						vec = vec[:0]
						buf = buf[:0]
						bufsz = 0
					}
					if isNil {
						// channel closed, so close writer and quit.
						if closer, ok := w.Writer.(io.Closer); ok {
							err1 := closer.Close()
							if err1 != nil && err == nil {
								err = err1
							}
						}
						w.chDone <- err
						ticker.Stop()
						return
					}
				case <-w.sync:
					if bufsz != 0 {
						_, err = writevFileWriter(fw, vec)
						_ = buf[len(buf)-1]
						for i := 0; i < len(buf); i++ {
							b1kpool.Put(buf[i])
						}
						vec = vec[:0]
						buf = buf[:0]
						bufsz = 0
					} else {
						err = nil
					}
					w.syncDone <- err
				case <-ticker.C:
					if bufsz != 0 {
						_, err = writevFileWriter(fw, vec)
						_ = buf[len(buf)-1]
						for i := 0; i < len(buf); i++ {
							b1kpool.Put(buf[i])
						}
						vec = vec[:0]
						buf = buf[:0]
						bufsz = 0
					} else {
						err = nil
					}
				}
			}
		}(w, w.Writer.(*FileWriter))
	})

	// copy and sends data
	w.ch <- append(b1kpool.Get().([]byte)[:0], p...)
	return len(p), nil
}

func writevFileWriter(w *FileWriter, iovec []syscall.Iovec) (n int, err error) {
	w.mu.Lock()

	if w.file == nil {
		if w.Filename == "" {
			n, err = writevFile(os.Stderr, iovec)
			w.mu.Unlock()
			return
		}
		err = w.create()
		if err != nil {
			w.mu.Unlock()
			return
		}
	}

	n, err = writevFile(w.file, iovec)
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

func writevFile(file *os.File, iovec []syscall.Iovec) (n int, err error) {
	n1, _, errno := syscall.Syscall(syscall.SYS_WRITEV, file.Fd(), uintptr(unsafe.Pointer(&iovec[0])), uintptr(len(iovec)))
	if errno != 0 {
		err = fmt.Errorf("writev failed with error: %d", errno)
		return
	}
	n = int(n1)
	return
}
