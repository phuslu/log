package log

import (
	"io"
	"os"
	"sync"
	"time"
)

type BufferWriter struct {
	Size     int
	Duration time.Duration
	Writer   io.Writer

	once sync.Once
	mu   sync.Mutex
	buf  []byte
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

func (w *BufferWriter) Close() (err error) {
	w.mu.Lock()
	_, err = w.Writer.Write(w.buf)
	w.buf = w.buf[:0]
	if closer, ok := w.Writer.(io.Closer); ok {
		err = closer.Close()
	}
	w.mu.Unlock()
	return
}

func (w *BufferWriter) Write(p []byte) (n int, err error) {
	w.once.Do(func() {
		if w.Size == 0 {
			return
		}
		if w.buf == nil {
			w.buf = make([]byte, 0, w.Size+os.Getpagesize())
		}
		if w.Duration > 0 {
			if w.Duration < 100*time.Millisecond {
				w.Duration = 100 * time.Millisecond
			}
			go func(w *BufferWriter) {
				for {
					time.Sleep(w.Duration)
					if len(w.buf) != 0 {
						w.Flush()
					}
				}
			}(w)
		}
	})

	w.mu.Lock()
	if w.Size > 0 {
		w.buf = append(w.buf, p...)
		n = len(p)
		if len(w.buf) > w.Size {
			_, err = w.Writer.Write(w.buf)
			w.buf = w.buf[:0]
		}
	} else {
		n, err = w.Writer.Write(p)
	}
	w.mu.Unlock()

	return
}
