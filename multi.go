package log

import (
	"bytes"
	"io"
)

// LeveledWriter is an interface that wraps WriteAtLevel.
type LeveledWriter interface {
	// WriteAtLevel decides which writers to use by checking the specified Level.
	WriteAtLevel(Level, []byte) (int, error)
}

// MultiWriter is an io.WriteCloser that log to different writers by different levels
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

	// ParseLevel specifies an optional callback for parse log level in output
	//    log.DebugLogger.Writer = &log.MultiWriter {
	//        InfoWriter:   &log.FileWriter{Filename: "main-info.log"},
	//        ErrorWriter:  &log.FileWriter{Filename: "main-error.log"},
	//        StderrWriter: &log.ConsoleWriter{ColorOutput: true},
	//        StderrLevel:  log.ErrorLevel,
	//        ParseLevel:   func([]byte) log.Level { return log.ParseLevel(string(p[49])) },
	//  }
	ParseLevel func([]byte) Level
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

// Write implements io.Writer.
func (w *MultiWriter) Write(p []byte) (n int, err error) {
	var level = noLevel
	if w.ParseLevel != nil {
		level = w.ParseLevel(p)
	} else {
		var l byte
		// guess level by fixed offset
		lp := len(p)
		if lp > 49 {
			_ = p[49]
			switch {
			case p[32] == 'Z' && p[42] == ':' && p[43] == '"':
				l = p[44]
			case p[32] == '+' && p[47] == ':' && p[48] == '"':
				l = p[49]
			}
		}
		// guess level by "level":" beginning
		if l == 0 {
			if i := bytes.Index(p, levelBegin); i > 0 && i+len(levelBegin)+1 < lp {
				l = p[i+len(levelBegin)]
			}
		}
		// convert byte to Level
		switch l {
		case 't':
			level = TraceLevel
		case 'd':
			level = DebugLevel
		case 'i':
			level = InfoLevel
		case 'w':
			level = WarnLevel
		case 'e':
			level = ErrorLevel
		case 'f':
			level = FatalLevel
		case 'p':
			level = PanicLevel
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
