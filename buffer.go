package log

import (
	"io"
	"os"
	"sync"
	"time"
)

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
			w.buf = make([]byte, 0, w.BufferSize+os.Getpagesize())
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
