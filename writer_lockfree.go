// +build windows linux

package log

import (
	"sync/atomic"
)

// rely to cl/174957 & cl/41674
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
