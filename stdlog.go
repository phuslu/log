package log

import (
	stdLog "log"
	"runtime"
)

// Std wraps the Logger to provide *stdLog.Logger
func (l *Logger) Std(level Level, context Context, prefix string, flag int) *stdLog.Logger {
	w := &stdLogWriter{
		logger:  *l,
		level:   level,
		context: context,
	}
	return stdLog.New(w, prefix, flag)
}

type stdLogWriter struct {
	logger  Logger
	level   Level
	context Context
}

func (w *stdLogWriter) Write(p []byte) (int, error) {
	if w.logger.silent(w.level) {
		return 0, nil
	}
	e := w.logger.header(w.level)
	if caller, full := w.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		_, file, line, _ := runtime.Caller(caller + 2)
		e.caller(file, line, full)
	}
	e.Context(w.context).Msg(b2s(p))
	return len(p), nil
}
