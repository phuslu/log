package log

import (
	"io"
	"sync"
	"time"
)

// AsyncWriter is an io.WriteCloser that writes asynchronously.
type AsyncWriter struct {
	// ChannelSize is the size of the data channel, the default size is 1.
	ChannelSize int

	// BatchSize is the batch size of underlying writer
	BatchSize int

	// SyncDuration specifies the sync duration of underlying writer when BatchSize is non-zero.
	SyncDuration time.Duration

	// Writer specifies the writer of output.
	Writer io.Writer

	once     sync.Once
	ch       chan []byte
	chDone   chan error
	sync     chan struct{}
	syncDone chan error
}

// Sync syncs all pending log I/O.
func (w *AsyncWriter) Sync() (err error) {
	if w.sync == nil {
		return
	}
	w.sync <- struct{}{}
	err = <-w.syncDone
	return
}

// Close implements io.Closer, and closes the underlying Writer.
func (w *AsyncWriter) Close() (err error) {
	w.ch <- nil // instead of close(w.ch) to avoid panic other goroutine
	err = <-w.chDone
	return
}

var b1kpool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024)
	},
}

// Write implements io.Writer.  If a write would cause the log buffer to be larger
// than Size, the buffer is written to the underlying Writer and cleared.
func (w *AsyncWriter) Write(p []byte) (int, error) {
	w.once.Do(func() {
		if w.ChannelSize == 0 {
			w.ChannelSize = 1
		}
		// channels
		w.ch = make(chan []byte, w.ChannelSize)
		w.chDone = make(chan error)
		if w.BatchSize <= 1 {
			go w.consumer0()
		} else {
			w.sync = make(chan struct{})
			w.syncDone = make(chan error)
			go w.consumerN()
		}
	})

	// copy and sends data
	w.ch <- append(b1kpool.Get().([]byte)[:0], p...)
	return len(p), nil
}

func (w *AsyncWriter) consumer0() {
	var err error
	for b := range w.ch {
		if b == nil {
			break
		}
		_, err = w.Writer.Write(b)
		b1kpool.Put(b)
	}
	w.chDone <- err
}

func (w *AsyncWriter) consumerN() {
	var err error
	buf := make([]byte, 0)
	batch := 0
	dur := w.SyncDuration
	if dur == 0 {
		dur = 5 * time.Second
	}
	ticker := time.NewTicker(dur)
	for {
		select {
		case b := <-w.ch:
			isNil := b == nil
			if len(b) != 0 {
				buf = append(buf, b...)
				batch++
				if cap(b) <= bbcap {
					b1kpool.Put(b)
				}
			}
			// full or closed
			if (batch >= w.BatchSize || isNil) && len(buf) != 0 {
				_, err = w.Writer.Write(buf)
				buf = buf[:0]
				batch = 0
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
			if len(buf) != 0 {
				_, err = w.Writer.Write(buf)
				buf = buf[:0]
				batch = 0
			} else {
				err = nil
			}
			w.syncDone <- err
		case <-ticker.C:
			if len(buf) != 0 {
				_, err = w.Writer.Write(buf)
				buf = buf[:0]
				batch = 0
			} else {
				err = nil
			}
		}
	}
}
