package log

import (
	"io"
)

// MultiWriterEntry is an array Writer that log to different writers
type MultiWriterEntry []Writer

// Close implements io.Closer, and closes the underlying MultiWriterEntry.
func (w *MultiWriterEntry) Close() (err error) {
	for _, writer := range *w {
		if closer, ok := writer.(io.Closer); ok {
			if err1 := closer.Close(); err1 != nil {
				err = err1
			}
		}
	}
	return
}

// WriteEntry implements entryWriter.
func (w *MultiWriterEntry) WriteEntry(e *Entry) (n int, err error) {
	var err1 error
	for _, writer := range *w {
		n, err1 = writer.WriteEntry(e)
		if err1 != nil && err == nil {
			err = err1
		}
	}
	return
}

var _ Writer = (*MultiWriterEntry)(nil)
