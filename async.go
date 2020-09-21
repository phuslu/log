package log

import (
	"io"
	"sync"
)

// AsyncWriter is an io.WriteCloser that writes asynchronously.
type AsyncWriter struct {
	// ChannelSize is the size of the data channel, the default size is 1.
	ChannelSize uint

	// Writer specifies the writer of output.
	Writer io.Writer

	once    sync.Once
	ch      chan *Event
	chClose chan error
}

// Close implements io.Closer, and closes the underlying Writer.
func (w *AsyncWriter) Close() (err error) {
	w.ch <- nil
	err = <-w.chClose
	return
}

// Write implements io.Writer.
func (w *AsyncWriter) Write(p []byte) (int, error) {
	e := epool.Get().(*Event)
	e.buf = append(e.buf[:0], p...)
	return w.WriteEvent(e)
}

// WriteEvent implements eventWriter.
func (w *AsyncWriter) WriteEvent(e *Event) (int, error) {
	w.once.Do(func() {
		// channels
		w.ch = make(chan *Event, w.ChannelSize)
		w.chClose = make(chan error)
		go func() {
			var err error
			for e := range w.ch {
				if e == nil {
					break
				}
				_, err = w.Writer.Write(e.buf)
				e.Discard()
			}
			w.chClose <- err
		}()
	})

	w.ch <- e
	return len(e.buf), nil
}
