package log

import (
	"io"
	"sync"
	"time"
)

// AsyncWriter is an io.WriteCloser that writes asynchronously.
type AsyncWriter struct {
	writer       io.Writer
	channelSize  int
	batchSize    int
	syncDuration time.Duration

	once    sync.Once
	ch      chan []byte
	chClose chan error
	chSync  chan error
}

func NewAsyncWriter(writer io.Writer, channelSize int, batchSize int, syncDuation time.Duration) (w *AsyncWriter) {
	w = &AsyncWriter{
		writer:       writer,
		channelSize:  channelSize,
		batchSize:    batchSize,
		syncDuration: syncDuation,
		ch:           make(chan []byte, channelSize),
		chClose:      make(chan error),
		chSync:       make(chan error),
	}
	if syncDuation == 0 && batchSize > 0 {
		w.syncDuration = 5 * time.Second
	}
	if batchSize <= 1 {
		go w.consumer0()
	} else {
		go w.consumerN()
	}
	return
}

// Sync syncs all pending log I/O.
func (w *AsyncWriter) Sync() (err error) {
	w.ch <- nil // nil is a sync trigger
	err = <-w.chSync
	return
}

// Close implements io.Closer, and closes the underlying Writer.
func (w *AsyncWriter) Close() (err error) {
	close(w.ch)
	err = <-w.chClose
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
	w.ch <- append(b1kpool.Get().([]byte)[:0], p...)
	return len(p), nil
}

func (w *AsyncWriter) consumer0() {
	var err error
	for b := range w.ch {
		if b == nil {
			w.chSync <- err
			continue
		}
		_, err = w.writer.Write(b)
		b1kpool.Put(b)
	}
	w.chClose <- err
}

func (w *AsyncWriter) consumerN() {
	var err error
	buf := make([]byte, 0)
	batch := 0
	ticker := time.NewTicker(w.syncDuration)
	for {
		select {
		case b, ok := <-w.ch:
			// batch
			if len(b) != 0 {
				buf = append(buf, b...)
				batch++
				if cap(b) <= bbcap {
					b1kpool.Put(b)
				}
			}
			// write
			if (batch >= w.batchSize || b == nil) && len(buf) != 0 {
				_, err = w.writer.Write(buf)
				buf = buf[:0]
				batch = 0
			}
			// close
			if !ok {
				if closer, ok := w.writer.(io.Closer); ok {
					err1 := closer.Close()
					if err1 != nil && err == nil {
						err = err1
					}
				}
				w.chClose <- err
				ticker.Stop()
				return
			}
			// sync
			if b == nil {
				w.chSync <- err
			}
		case <-ticker.C:
			if len(buf) != 0 {
				_, err = w.writer.Write(buf)
				buf = buf[:0]
				batch = 0
			} else {
				err = nil
			}
		}
	}
}
