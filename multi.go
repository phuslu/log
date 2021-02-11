package log

import (
	"io"
)

// MultiWriter is an Writer that log to different writers by different levels
type MultiWriter struct {
	// InfoWriter specifies all the level logs writes to
	InfoWriter Writer

	// WarnWriter specifies the level greater than or equal to WarnLevel writes to
	WarnWriter Writer

	// WarnWriter specifies the level greater than or equal to ErrorLevel writes to
	ErrorWriter Writer

	// ConsoleWriter specifies the console writer
	ConsoleWriter Writer

	// ConsoleLevel specifies the level greater than or equal to it also writes to
	ConsoleLevel Level
}

// Close implements io.Closer, and closes the underlying LeveledWriter.
func (w *MultiWriter) Close() (err error) {
	for _, writer := range []Writer{
		w.InfoWriter,
		w.WarnWriter,
		w.ErrorWriter,
		w.ConsoleWriter,
	} {
		if writer == nil {
			continue
		}
		if closer, ok := writer.(io.Closer); ok {
			if err1 := closer.Close(); err1 != nil {
				err = err1
			}
		}
	}
	return
}

// WriteEntry implements entryWriter.
func (w *MultiWriter) WriteEntry(e *Entry) (n int, err error) {
	var err1 error
	switch e.Level {
	case noLevel, PanicLevel, FatalLevel, ErrorLevel:
		if w.ErrorWriter != nil {
			n, err1 = w.ErrorWriter.WriteEntry(e)
			if err1 != nil && err == nil {
				err = err1
			}
		}
		fallthrough
	case WarnLevel:
		if w.WarnWriter != nil {
			n, err1 = w.WarnWriter.WriteEntry(e)
			if err1 != nil && err == nil {
				err = err1
			}
		}
		fallthrough
	default:
		if w.InfoWriter != nil {
			n, err1 = w.InfoWriter.WriteEntry(e)
			if err1 != nil && err == nil {
				err = err1
			}
		}
	}

	if w.ConsoleWriter != nil && e.Level >= w.ConsoleLevel {
		_, _ = w.ConsoleWriter.WriteEntry(e)
	}

	return
}

var _ Writer = (*MultiWriter)(nil)
