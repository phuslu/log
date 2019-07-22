package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type GlogLogger struct {
	Level  Level
	Writer io.Writer
}

type GlogEvent struct {
	buf   []byte
	level Level
	write func(p []byte) (n int, err error)
}

func (l GlogLogger) Info(args ...interface{}) {
	l.withLevel(InfoLevel).Print(args...)
}

func (l GlogLogger) Infoln(args ...interface{}) {
	l.withLevel(InfoLevel).Println(args...)
}

func (l GlogLogger) Infof(format string, args ...interface{}) {
	l.withLevel(InfoLevel).Printf(format, args...)
}

func (l GlogLogger) InfoDepth(depth int, args ...interface{}) {
	l.withLevel(InfoLevel).PrintDepth(depth, args...)
}

func (l GlogLogger) Warning(args ...interface{}) {
	l.withLevel(WarnLevel).Print(args...)
}

func (l GlogLogger) Warningln(args ...interface{}) {
	l.withLevel(WarnLevel).Println(args...)
}

func (l GlogLogger) Warningf(format string, args ...interface{}) {
	l.withLevel(WarnLevel).Printf(format, args...)
}

func (l GlogLogger) WarningDepth(depth int, args ...interface{}) {
	l.withLevel(WarnLevel).PrintDepth(depth, args...)
}

func (l GlogLogger) Error(args ...interface{}) {
	l.withLevel(ErrorLevel).Print(args...)
}

func (l GlogLogger) Errorln(args ...interface{}) {
	l.withLevel(ErrorLevel).Println(args...)
}

func (l GlogLogger) Errorf(format string, args ...interface{}) {
	l.withLevel(ErrorLevel).Printf(format, args...)
}

func (l GlogLogger) ErrorDepth(depth int, args ...interface{}) {
	l.withLevel(ErrorLevel).PrintDepth(depth, args...)
}

func (l GlogLogger) Fatal(args ...interface{}) {
	l.withLevel(FatalLevel).Print(args...)
}

func (l GlogLogger) Fatalln(args ...interface{}) {
	l.withLevel(FatalLevel).Println(args...)
}

func (l GlogLogger) Fatalf(format string, args ...interface{}) {
	l.withLevel(FatalLevel).Printf(format, args...)
}

func (l GlogLogger) FatalDepth(depth int, args ...interface{}) {
	l.withLevel(FatalLevel).PrintDepth(depth, args...)
}

func (l GlogLogger) V(level int) bool {
	return level >= int(l.Level)
}

var gepool = sync.Pool{
	New: func() interface{} {
		return new(GlogEvent)
	},
}

var pid = int64(os.Getpid())

func (l GlogLogger) withLevel(level Level) (e *GlogEvent) {
	if level < l.Level {
		return
	}
	// [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg
	e = gepool.Get().(*GlogEvent)
	e.buf = e.buf[:0]
	e.level = level
	e.write = l.Writer.Write
	// level
	switch level {
	case DebugLevel:
		e.buf = append(e.buf, 'D')
	case InfoLevel:
		e.buf = append(e.buf, 'I')
	case WarnLevel:
		e.buf = append(e.buf, 'W')
	case ErrorLevel:
		e.buf = append(e.buf, 'E')
	case FatalLevel:
		e.buf = append(e.buf, 'F')
	default:
		e.buf = append(e.buf, '?')
	}
	// time
	now := timeNow()
	e.time(now)
	e.buf = append(e.buf, ' ')
	// threadid
	e.buf = strconv.AppendInt(e.buf, pid, 10)
	e.buf = append(e.buf, ' ')

	return
}

func (e *GlogEvent) Printf(format string, args ...interface{}) {
	if e == nil {
		return
	}
	_, file, line, _ := runtime.Caller(2)
	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}
	e.buf = append(e.buf, file...)
	e.buf = append(e.buf, ':')
	e.buf = strconv.AppendInt(e.buf, int64(line), 10)
	e.buf = append(e.buf, ']', ' ')
	e.buf = append(e.buf, fmt.Sprintf(format, args...)...)
	e.buf = append(e.buf, '\n')
	e.write(e.buf)
	if e.level == FatalLevel {
		e.write(stacks(false))
		e.write(stacks(true))
		os.Exit(255)
	}
	gepool.Put(e)
}

func (e *GlogEvent) Print(args ...interface{}) {
	if e == nil {
		return
	}
	_, file, line, _ := runtime.Caller(2)
	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}
	e.buf = append(e.buf, file...)
	e.buf = append(e.buf, ':')
	e.buf = strconv.AppendInt(e.buf, int64(line), 10)
	e.buf = append(e.buf, ']', ' ')
	e.buf = append(e.buf, fmt.Sprint(args...)...)
	e.buf = append(e.buf, '\n')
	e.write(e.buf)
	if e.level == FatalLevel {
		e.write(stacks(false))
		e.write(stacks(true))
		os.Exit(255)
	}
	gepool.Put(e)
}

func (e *GlogEvent) Println(args ...interface{}) {
	if e == nil {
		return
	}
	_, file, line, _ := runtime.Caller(2)
	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}
	e.buf = append(e.buf, file...)
	e.buf = append(e.buf, ':')
	e.buf = strconv.AppendInt(e.buf, int64(line), 10)
	e.buf = append(e.buf, ']', ' ')
	e.buf = append(e.buf, fmt.Sprintln(args...)...)
	e.buf = append(e.buf, '\n')
	e.write(e.buf)
	if e.level == FatalLevel {
		e.write(stacks(false))
		e.write(stacks(true))
		os.Exit(255)
	}
	gepool.Put(e)
}

func (e *GlogEvent) PrintDepth(depth int, args ...interface{}) {
	if e == nil {
		return
	}
	_, file, line, _ := runtime.Caller(2 + depth)
	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}
	e.buf = append(e.buf, file...)
	e.buf = append(e.buf, ':')
	e.buf = strconv.AppendInt(e.buf, int64(line), 10)
	e.buf = append(e.buf, ']', ' ')
	e.buf = append(e.buf, fmt.Sprint(args...)...)
	e.buf = append(e.buf, '\n')
	e.write(e.buf)
	if e.level == FatalLevel {
		e.write(stacks(false))
		e.write(stacks(true))
		os.Exit(255)
	}
	gepool.Put(e)
}

func (e *GlogEvent) time(now time.Time) {
	var n = len(e.buf)
	e.buf = append(e.buf, "0102 15:04:05.999999"...)
	var a, b int
	// month
	a = int(now.Month())
	b = a / 10
	e.buf[n+1] = byte('0' + a - 10*b)
	e.buf[n] = byte('0' + b)
	// day
	a = now.Day()
	b = a / 10
	e.buf[n+3] = byte('0' + a - 10*b)
	e.buf[n+2] = byte('0' + b)
	// hour
	a = now.Hour()
	b = a / 10
	e.buf[n+6] = byte('0' + a - 10*b)
	e.buf[n+5] = byte('0' + b)
	// minute
	a = now.Minute()
	b = a / 10
	e.buf[n+9] = byte('0' + a - 10*b)
	e.buf[n+8] = byte('0' + b)
	// second
	a = now.Second()
	b = a / 10
	e.buf[n+12] = byte('0' + a - 10*b)
	e.buf[n+11] = byte('0' + b)
	// milli second
	a = now.Nanosecond() / 1000
	b = a / 10
	e.buf[n+19] = byte('0' + a - 10*b)
	a = b
	b = a / 10
	e.buf[n+18] = byte('0' + a - 10*b)
	a = b
	b = a / 10
	e.buf[n+17] = byte('0' + a - 10*b)
	a = b
	b = a / 10
	e.buf[n+16] = byte('0' + a - 10*b)
	a = b
	b = a / 10
	e.buf[n+15] = byte('0' + a - 10*b)
	e.buf[n+14] = byte('0' + b)
}
