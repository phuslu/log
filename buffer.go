package log

import (
	"io"
	"os"
	"sync"
	"time"
)

// BufferWriter is an io.WriteCloser that writes with fixed size buffer.
type BufferWriter struct {
	// MaxSize is the size in bytes of the buffer before it gets flushed.
	MaxSize int

	// FlushDuration is the period of the writer flush duration
	FlushDuration time.Duration

	// Out specifies the writer of output.
	Out io.Writer

	once sync.Once
	mu   sync.Mutex
	buf  []byte
}

// Flush flushes all pending log I/O.
func (w *BufferWriter) Flush() (err error) {
	w.mu.Lock()
	if len(w.buf) != 0 {
		_, err = w.Out.Write(w.buf)
		w.buf = w.buf[:0]
	}
	w.mu.Unlock()
	return
}

// Close implements io.Closer, and closes the underlaying Writer.
func (w *BufferWriter) Close() (err error) {
	w.mu.Lock()
	_, err = w.Out.Write(w.buf)
	w.buf = w.buf[:0]
	if closer, ok := w.Out.(io.Closer); ok {
		err = closer.Close()
	}
	w.mu.Unlock()
	return
}

// Write implements io.Writer.  If a write would cause the log buffer to be larger
// than Size, the buffer is written to the underlaying Writer and cleared.
func (w *BufferWriter) Write(p []byte) (n int, err error) {
	w.once.Do(func() {
		if w.MaxSize == 0 {
			return
		}
		if page := os.Getpagesize(); w.MaxSize%page != 0 {
			w.MaxSize = (w.MaxSize + page) / page * page
		}
		if w.buf == nil {
			w.buf = make([]byte, 0, w.MaxSize)
		}
		if w.FlushDuration > 0 {
			if w.FlushDuration < 100*time.Millisecond {
				w.FlushDuration = 100 * time.Millisecond
			}
			go func(w *BufferWriter) {
				for {
					time.Sleep(w.FlushDuration)
					if len(w.buf) != 0 {
						w.Flush()
					}
				}
			}(w)
		}
	})

	w.mu.Lock()
	if w.MaxSize > 0 {
		w.buf = append(w.buf, p...)
		n = len(p)
		if len(w.buf) > w.MaxSize {
			_, err = w.Out.Write(w.buf)
			w.buf = w.buf[:0]
		}
	} else {
		n, err = w.Out.Write(p)
	}
	w.mu.Unlock()

	return
}
