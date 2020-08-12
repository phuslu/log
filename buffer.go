package log

import (
	"io"
	"os"
	"sync"
	"time"
)

// BufferWriter is an io.WriteCloser that writes with fixed size buffer.
type BufferWriter struct {
	// BufferSize is the size in bytes of the buffer before it gets flushed.
	BufferSize int

	// FlushDuration is the period of the writer flush duration
	FlushDuration time.Duration

	// Writer specifies the writer of output.
	Writer io.Writer

	once sync.Once
	mu   sync.Mutex
	buf  []byte
}

// Flush flushes all pending log I/O.
func (w *BufferWriter) Flush() (err error) {
	w.mu.Lock()
	if len(w.buf) != 0 {
		_, err = w.Writer.Write(w.buf)
		w.buf = w.buf[:0]
	}
	w.mu.Unlock()
	return
}

// Close implements io.Closer, and closes the underlaying Writer.
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

// Write implements io.Writer.  If a write would cause the log buffer to be larger
// than Size, the buffer is written to the underlaying Writer and cleared.
func (w *BufferWriter) Write(p []byte) (n int, err error) {
	w.once.Do(func() {
		if w.BufferSize == 0 {
			return
		}
		if page := os.Getpagesize(); w.BufferSize%page != 0 {
			w.BufferSize = (w.BufferSize + page) / page * page
		}
		if w.buf == nil {
			w.buf = make([]byte, 0, w.BufferSize)
		}
		if w.FlushDuration > 0 {
			if w.FlushDuration < 100*time.Millisecond {
				w.FlushDuration = 100 * time.Millisecond
			}
			go func(w *BufferWriter) {
				for {
					time.Sleep(w.FlushDuration)
					w.Flush()
				}
			}(w)
		}
	})

	w.mu.Lock()
	if w.BufferSize > 0 {
		w.buf = append(w.buf, p...)
		n = len(p)
		if len(w.buf) > w.BufferSize {
			_, err = w.Writer.Write(w.buf)
			w.buf = w.buf[:0]
		}
	} else {
		n, err = w.Writer.Write(p)
	}
	w.mu.Unlock()

	return
}

// The Flusher interface is implemented by BufferWriters that allow
// an Logger to flush buffered data to the output.
type Flusher interface {
	// Flush sends any buffered data to the output.
	Flush() error
}

// Flush writes any buffered data to the underlying io.Writer.
func Flush(writer io.Writer) (err error) {
	if flusher, ok := writer.(Flusher); ok {
		err = flusher.Flush()
	}
	return
}
