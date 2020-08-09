package log

import (
	"bytes"
	"io"
)

// MultiWriter is an io.WriteCloser that writes to multi writers
type MultiWriter struct {
	// InfoWriter specifies the level large than info logs writes to
	InfoWriter io.Writer

	// WarnWriter specifies the level large than warn logs writes to
	WarnWriter io.Writer

	// WarnWriter specifies the level large than error logs writes to
	ErrorWriter io.Writer

	// StderrWriter specifies the stderr writer
	StderrWriter io.Writer

	// StderrLevel specifies the minimal level logs it will be writes to stderr
	StderrLevel Level

	// ParseLevel defines the callback finds out level from output, optional
	ParseLevel func([]byte) Level
}

// Close implements io.Closer, and closes the underlaying Writers.
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
		Flush(writer)
		if closer, ok := writer.(io.Closer); ok {
			if err1 := closer.Close(); err1 != nil {
				err = err1
			}
		}
	}
	return
}

var levelBegin = []byte(`"level":"`)

// Write implements io.Writer.
func (w *MultiWriter) Write(p []byte) (n int, err error) {
	var level = noLevel
	if w.ParseLevel != nil {
		level = w.ParseLevel(p)
	} else {
		if i := bytes.Index(p, levelBegin); i > 0 && i+len(levelBegin)+1 < len(p) {
			switch p[i+len(levelBegin)] {
			case 't', 'T':
				level = TraceLevel
			case 'd', 'D':
				level = DebugLevel
			case 'i', 'I':
				level = InfoLevel
			case 'w', 'W':
				level = WarnLevel
			case 'e', 'E':
				level = ErrorLevel
			case 'f', 'F':
				level = FatalLevel
			case 'p', 'P':
				level = PanicLevel
			}
		}
	}

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
