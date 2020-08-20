package log

import (
	"io"
)

// LeveledWriter is an interface that wraps WriteAtLevel.
type LeveledWriter interface {
	// WriteAtLevel decides which writers to use by checking the specified Level.
	WriteAtLevel(Level, []byte) (int, error)
}

// MultiWriter is an io.WriteCloser that log to different writers by different levels
type MultiWriter struct {
	// InfoWriter specifies all the level logs writes to
	InfoWriter io.Writer

	// WarnWriter specifies the level large than warn logs writes to
	WarnWriter io.Writer

	// WarnWriter specifies the level large than error logs writes to
	ErrorWriter io.Writer

	// StderrWriter specifies the stderr writer
	StderrWriter io.Writer

	// StderrLevel specifies the minimal level logs it will be writes to stderr
	StderrLevel Level
}

// Close implements io.Closer, and closes the underlying LeveledWriter.
func (w *MultiWriter) Close() (err error) {
	for _, writer := range []io.Writer{
		w.InfoWriter,
		w.WarnWriter,
		w.ErrorWriter,
		w.StderrWriter,
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

// WriteAtLevel implements LeveledWriter.
func (w *MultiWriter) WriteAtLevel(level Level, p []byte) (n int, err error) {
	var err1 error
	switch level {
	case noLevel, PanicLevel, FatalLevel, ErrorLevel:
		if w.ErrorWriter != nil {
			n, err1 = w.ErrorWriter.Write(p)
			if err1 != nil && err == nil {
				err = err1
			}
		}
		fallthrough
	case WarnLevel:
		if w.WarnWriter != nil {
			n, err1 = w.WarnWriter.Write(p)
			if err1 != nil && err == nil {
				err = err1
			}
		}
		fallthrough
	default:
		if w.InfoWriter != nil {
			n, err1 = w.InfoWriter.Write(p)
			if err1 != nil && err == nil {
				err = err1
			}
		}
	}

	if w.StderrWriter != nil && level >= w.StderrLevel {
		w.StderrWriter.Write(p)
	}

	return
}

// wrapLeveledWriter wraps a LeveledWriter to implement io.Writer.
type wrapLeveledWriter struct {
	Level         Level
	LeveledWriter LeveledWriter
}

// Write implements io.Writer.
func (w wrapLeveledWriter) Write(p []byte) (int, error) {
	return w.LeveledWriter.WriteAtLevel(w.Level, p)
}

var _ io.Writer = wrapLeveledWriter{}
