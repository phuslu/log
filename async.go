package log

import (
	"io"
	"sync"
)

// AsyncWriter is an Writer that writes asynchronously.
type AsyncWriter struct {
	// ChannelSize is the size of the data channel, the default size is 1.
	ChannelSize uint

	// Writer specifies the writer of output.
	Writer Writer

	once    sync.Once
	ch      chan *Entry
	chClose chan error
}

// Close implements io.Closer, and closes the underlying Writer.
func (w *AsyncWriter) Close() (err error) {
	w.ch <- nil
	err = <-w.chClose
	if closer, ok := w.Writer.(io.Closer); ok {
		if err1 := closer.Close(); err1 != nil {
			err = err1
		}
	}
	return
}

// WriteEntry implements Writer.
func (w *AsyncWriter) WriteEntry(e *Entry) (int, error) {
	w.once.Do(func() {
		// channels
		w.ch = make(chan *Entry, w.ChannelSize)
		w.chClose = make(chan error)
		go func() {
			var err error
			for entry := range w.ch {
				if entry == nil {
					break
				}
				_, err = w.Writer.WriteEntry(entry)
				epool.Put(entry)
			}
			w.chClose <- err
		}()
	})

	// cheating to logger pool
	entry := epool.Get().(*Entry)
	entry.Level = e.Level
	entry.buf, e.buf = e.buf, entry.buf

	w.ch <- entry
	return len(entry.buf), nil
}

var _ Writer = (*AsyncWriter)(nil)
