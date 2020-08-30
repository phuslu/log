package log

import (
	"io"
	"sync"
	"time"
)

// AsyncWriter is an io.WriteCloser that writes asynchronously.
type AsyncWriter struct {
	// BufferSize is the size in bytes of the buffer, the default size is 32KB.
	BufferSize int

	// ChannelSize is the size of the data channel, the default size is 100.
	ChannelSize int

	// SyncDuration is the duration of the writer syncs, the default duration is 5s.
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
		// data consumer
		go w.consumer()
	})

	// copy and sends data
	w.ch <- append(b1kpool.Get().([]byte)[:0], p...)
	return len(p), nil
}

func (w *AsyncWriter) consumer() {
	var err error
	buf := make([]byte, 0, w.BufferSize+4096)
	ticker := time.NewTicker(w.SyncDuration)
	for {
		select {
		case b := <-w.ch:
			isNil := b == nil
			if len(b) != 0 {
				buf = append(buf, b...)
				if cap(b) <= bbcap {
					b1kpool.Put(b)
				}
			}
			// full or closed
			if len(buf) >= w.BufferSize || (isNil && len(buf) != 0) {
				_, err = w.Writer.Write(buf)
				buf = buf[:0]
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
			} else {
				err = nil
			}
			w.syncDone <- err
		case <-ticker.C:
			if len(buf) != 0 {
				_, err = w.Writer.Write(buf)
				buf = buf[:0]
			} else {
				err = nil
			}
		}
	}
}
