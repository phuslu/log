package log

import (
	"fmt"
	"unsafe"
)

type RandomLogger struct {
	Logger
	N uint32
}

var _ = unsafe.Sizeof(0)

//go:noescape
//go:linkname fastrandn runtime.fastrandn
func fastrandn(x uint32) uint32

func (l RandomLogger) WithLevel(level Level) (e *Event) {
	if fastrandn(l.N) != 0 {
		return nil
	}
	return l.Logger.WithLevel(level)
}

func (l RandomLogger) Debug() *Event {
	return l.WithLevel(DebugLevel)
}

func (l RandomLogger) Info() *Event {
	return l.WithLevel(InfoLevel)
}

func (l RandomLogger) Warn() *Event {
	return l.WithLevel(WarnLevel)
}

func (l RandomLogger) Error() *Event {
	return l.WithLevel(ErrorLevel)
}

func (l RandomLogger) Fatal() *Event {
	return l.WithLevel(FatalLevel)
}

func (l RandomLogger) Print(v ...interface{}) {
	l.WithLevel(l.Level).Msg(fmt.Sprint(v...))
}

func (l RandomLogger) Printf(format string, v ...interface{}) {
	l.WithLevel(l.Level).Msgf(format, v...)
}
