package log

import (
	"errors"
	"io"
	"runtime"
	"sync"
	"unsafe"
)

// AsyncWriter is a Writer that writes asynchronously.
type AsyncWriter struct {
	// Writer specifies the writer of output.
	Writer Writer

	// ChannelSize is the size of the data channel, the default size is 1.
	ChannelSize uint

	// DiscardOnFull determines whether to discard new entry when the channel is full.
	DiscardOnFull bool

	// DisableWritev disables the writev syscall if the Writer is a FileWriter.
	DisableWritev bool

	once    sync.Once
	ch      chan *Entry
	chClose chan error
	file    *FileWriter

	errOnce  sync.Once
	firstErr error
}

func (w *AsyncWriter) init() {
	w.ch = make(chan *Entry, w.ChannelSize)
	w.chClose = make(chan error)
	w.file, _ = w.Writer.(*FileWriter)
	if w.file != nil && runtime.GOOS == "linux" && unsafe.Sizeof(uintptr(0)) == 8 && !w.DisableWritev {
		go w.writever()
	} else {
		go w.writer()
	}
}

// Close implements io.Closer, and closes the underlying Writer.
//
// Close must be called after all producers stopped calling Write or
// WriteEntry.  It waits until every accepted entry has been processed and
// returns the first background write error, if any.  A later successful write
// or a close error of the underlying Writer must not hide that first error.
func (w *AsyncWriter) Close() (err error) {
	w.once.Do(w.init)
	close(w.ch)
	err = <-w.chClose
	if closer, ok := w.Writer.(io.Closer); ok {
		if err1 := closer.Close(); err1 != nil && err == nil {
			err = err1
		}
	}
	return
}

// latchErr records the first background write error.  Later successes and
// later errors never replace it.
func (w *AsyncWriter) latchErr(err error) {
	if err == nil {
		return
	}
	w.errOnce.Do(func() {
		w.firstErr = err
	})
}

// recycle returns an entry to the pool, unless its buffer grew too large to
// keep around.
func (w *AsyncWriter) recycle(e *Entry) {
	if cap(e.buf) <= bbcap {
		epool.Put(e)
	}
}

var ErrAsyncWriterFull = errors.New("async writer is full")

var eepool = sync.Pool{
	New: func() any {
		return &Entry{
			Level: InfoLevel,
		}
	},
}

// Write implements io.Writer.
//
// The payload is copied, so the caller may reuse p as soon as Write returns,
// as required by io.Writer; holding on to p would race with the background
// writer and silently corrupt the log.
func (w *AsyncWriter) Write(p []byte) (n int, err error) {
	e := eepool.Get().(*Entry)
	if cap(e.buf) < len(p) {
		e.buf = make([]byte, len(p))
	} else {
		e.buf = e.buf[:len(p)]
	}
	copy(e.buf, p)
	n, err = w.WriteEntry(e)
	if cap(e.buf) <= bbcap {
		e.buf = e.buf[:0]
		eepool.Put(e)
	}
	return
}

// WriteEntry implements Writer.
func (w *AsyncWriter) WriteEntry(e *Entry) (int, error) {
	w.once.Do(w.init)

	// cheating to logger pool
	entry := epool.Get().(*Entry)
	entry.Level = e.Level
	entry.buf, e.buf = e.buf, entry.buf

	// snapshot length before sending, entry is owned by the writer goroutine afterwards
	n := len(entry.buf)

	if w.DiscardOnFull {
		select {
		case w.ch <- entry:
			return n, nil
		default:
			if cap(entry.buf) <= bbcap {
				epool.Put(entry)
			}
			return 0, ErrAsyncWriterFull
		}
	} else {
		w.ch <- entry
		return n, nil
	}
}

func (w *AsyncWriter) writer() {
	for entry := range w.ch {
		if entry == nil {
			break
		}
		if w.firstErr == nil {
			if _, err := w.Writer.WriteEntry(entry); err != nil {
				w.latchErr(err)
			}
		}
		w.recycle(entry)
	}
	w.chClose <- w.firstErr
}

var _ Writer = (*AsyncWriter)(nil)
var _ io.Writer = (*AsyncWriter)(nil)
