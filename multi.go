package log

import (
	"bytes"
	"io"
)

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

	// ParseLevel specifies an optional callback for parse log level from JSON input
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
		if closer, ok := writer.(io.Closer); ok {
			if err1 := closer.Close(); err1 != nil {
				err = err1
			}
		}
	}
	return
}

// Write implements io.Writer.
func (w *MultiWriter) Write(p []byte) (n int, err error) {
	var level Level
	if w.ParseLevel != nil {
		level = w.ParseLevel(p)
	} else {
		level = guessLevel(p)
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

var levelBegin = []byte(`"level":"`)

func guessLevel(p []byte) Level {
	var c byte

	// guess level by fixed offset
	lp := len(p)
	if lp > 49 {
		_ = p[49]
		switch {
		case p[32] == 'Z' && p[42] == ':' && p[43] == '"':
			c = p[44]
		case p[32] == '+' && p[47] == ':' && p[48] == '"':
			c = p[49]
		}
	}

	// guess level by "level":" beginning
	if c == 0 {
		if i := bytes.Index(p, levelBegin); i > 0 && i+len(levelBegin)+1 < lp {
			c = p[i+len(levelBegin)]
		}
	}

	// convert byte to Level
	return ParseLevelByte(c)
}
