package log

import (
	"io"
	"sync"
	"time"
)

// AsyncWriter is an io.WriteCloser that writes asynchronously.
type AsyncWriter struct {
	// BufferSize is the size in bytes of the buffer, using 32Kb by default.
	BufferSize int

	// ChannelSize is the size of the data channel, using 100 by default.
	ChannelSize int

	// SyncDuration is the period of the writer sync duration, using 5s by default.
	SyncDuration time.Duration

	// Writer specifies the writer of output.
	Writer io.Writer

	once     sync.Once
	ch       chan []byte
	chDone   chan error
	sync     chan struct{}
	syncDone chan error
}

// Sync sends all pending log I/O.
func (w *AsyncWriter) Sync() (err error) {
	w.sync <- struct{}{}
	err = <-w.syncDone
	return
}

// Close implements io.Closer, and closes the underlying Writer.
func (w *AsyncWriter) Close() (err error) {
	close(w.ch)
	err = <-w.chDone
	return
}

var a2kpool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 2048)
	},
}

// Write implements io.Writer.  If a write would cause the log buffer to be larger
// than Size, the buffer is written to the underlying Writer and cleared.
func (w *AsyncWriter) Write(p []byte) (n int, err error) {
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
		// data routine
		go func(w *AsyncWriter) {
			var err error
			buf := make([]byte, 0, w.BufferSize+4096)
			ticker := time.NewTicker(w.SyncDuration)
			for {
				select {
				case b, ok := <-w.ch:
					if len(b) != 0 {
						buf = append(buf, b...)
						a2kpool.Put(b)
					}
					// full or closed
					if len(buf) >= w.BufferSize || !ok {
						_, err = w.Writer.Write(buf)
						buf = buf[:0]
					}
					if !ok {
						// channel closed, so close writer and quit.
						if closer, ok := w.Writer.(io.Closer); ok {
							err1 := closer.Close()
							if err1 != nil && err == nil {
								err = err1
							}
						}
						ticker.Stop()
						w.chDone <- err
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
		}(w)
	})

	// copy and sends data
	w.ch <- append(a2kpool.Get().([]byte)[:0], p...)

	return
}
