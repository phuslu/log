package log

import (
	stdLog "log"
	"runtime"
	"unsafe"
)

// Std wraps the Logger to provide *stdLog.Logger
func (l *Logger) Std(level Level, context Context, prefix string, flag int) *stdLog.Logger {
	w := &levelWriter{
		logger:  *l,
		level:   level,
		context: context,
	}
	return stdLog.New(w, prefix, flag)
}

type levelWriter struct {
	logger  Logger
	level   Level
	context Context
}

func (w *levelWriter) Write(p []byte) (int, error) {
	e := w.logger.header(w.level)
	if e == nil {
		return 0, nil
	}
	if w.logger.Caller > 0 {
		e.caller(runtime.Caller(w.logger.Caller + 2))
	}
	e.Context(w.context).Msg(*(*string)(unsafe.Pointer(&p)))
	return len(p), nil
}
