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

func (l RandomLogger) withLevel(level Level) (e *Event) {
	if fastrandn(l.N) != 0 {
		return nil
	}
	return l.Logger.withLevel(level)
}

func (l RandomLogger) Debug() *Event {
	return l.withLevel(DebugLevel)
}

func (l RandomLogger) Info() *Event {
	return l.withLevel(InfoLevel)
}

func (l RandomLogger) Warn() *Event {
	return l.withLevel(WarnLevel)
}

func (l RandomLogger) Error() *Event {
	return l.withLevel(ErrorLevel)
}

func (l RandomLogger) Fatal() *Event {
	return l.withLevel(FatalLevel)
}

func (l RandomLogger) Print(v ...interface{}) {
	l.withLevel(l.Level).Msg(fmt.Sprint(v...))
}

func (l RandomLogger) Printf(format string, v ...interface{}) {
	l.withLevel(l.Level).Msgf(format, v...)
}
