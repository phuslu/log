package log

import (
	"errors"
	"io"
	"runtime"
	"sync"
	"unsafe"
)

// AsyncWriter is an Writer that writes asynchronously.
type AsyncWriter struct {
	// Writer specifies the writer of output.
	Writer Writer

	// ChannelSize is the size of the data channel, the default size is 1.
	ChannelSize uint

	// DiscardOnFull determines whether to discard new entry when the channel is full.
	DiscardOnFull bool

	// WritevDisabled disables the writev syscall if the Writer is a FileWriter.
	WritevDisabled bool

	once    sync.Once
	ch      chan *Entry
	chClose chan error
	file    *FileWriter
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

var ErrAsyncWriterFull = errors.New("async writer is full")

// WriteEntry implements Writer.
func (w *AsyncWriter) WriteEntry(e *Entry) (int, error) {
	w.once.Do(func() {
		// channels
		w.ch = make(chan *Entry, w.ChannelSize)
		w.chClose = make(chan error)
		w.file, _ = w.Writer.(*FileWriter)
		if w.file != nil && runtime.GOOS == "linux" && unsafe.Sizeof(uintptr(0)) == 8 && !w.WritevDisabled {
			go w.writever()
		} else {
			go w.writer()
		}
	})

	// cheating to logger pool
	entry := epool.Get().(*Entry)
	entry.Level = e.Level
	entry.buf, e.buf = e.buf, entry.buf

	if w.DiscardOnFull {
		select {
		case w.ch <- entry:
			return len(entry.buf), nil
		default:
			return 0, ErrAsyncWriterFull
		}
	} else {
		w.ch <- entry
		return len(entry.buf), nil
	}
}

func (w *AsyncWriter) writer() {
	var err error
	for entry := range w.ch {
		if entry == nil {
			break
		}
		_, err = w.Writer.WriteEntry(entry)
		epool.Put(entry)
	}
	w.chClose <- err
}

var _ Writer = (*AsyncWriter)(nil)
