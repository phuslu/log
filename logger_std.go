package log

import (
	stdLog "log"
)

type stdLogWriter struct {
	Logger
}

func (w *stdLogWriter) Write(p []byte) (int, error) {
	if w.Logger.silent(w.Logger.Level) {
		return 0, nil
	}
	e := w.Logger.header(w.Level)
	if caller := w.Logger.Caller; caller != 0 {
		if caller < 0 {
			caller = -caller
		}
		var pc uintptr
		e.caller(caller1(caller+2, &pc, 1, 1), pc, w.Logger.Caller < 0)
	}
	e.Msg(b2s(p))
	return len(p), nil
}

// Std wraps the Logger to provide *stdLog.Logger
func (l *Logger) Std(prefix string, flag int) *stdLog.Logger {
	return stdLog.New(&stdLogWriter{*l}, prefix, flag)
}
