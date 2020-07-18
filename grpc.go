package log

import (
	"runtime"
)

// GrpcLogger implements methods to satisfy interface
// google.golang.org/grpc/grpclog.LoggerV2.
type GrpcLogger struct {
	logger  Logger
	context Context
}

// Grpc wraps the Logger to provide a more ergonomic, but slightly slower,
// API. Grpcing a Logger is quite inexpensive, so it's reasonable for a
// single application to use both Loggers and GrpcLoggers, converting
// between them on the boundaries of performance-sensitive code.
func (l *Logger) Grpc(context Context) (g *GrpcLogger) {
	g = &GrpcLogger{
		logger:  *l,
		context: context,
	}
	return
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (g *GrpcLogger) Info(args ...interface{}) {
	e := g.logger.header(InfoLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	print(e.Context(g.context), args)
}

// Infoln logs to INFO log. Arguments are handled in the manner of fmt.Println.
func (g *GrpcLogger) Infoln(args ...interface{}) {
	e := g.logger.header(InfoLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	print(e.Context(g.context), args)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func (g *GrpcLogger) Infof(format string, args ...interface{}) {
	e := g.logger.header(InfoLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	e.Context(g.context).Msgf(format, args...)
}

// Warning logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (g *GrpcLogger) Warning(args ...interface{}) {
	e := g.logger.header(WarnLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	print(e.Context(g.context), args)
}

// Warningln logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func (g *GrpcLogger) Warningln(args ...interface{}) {
	e := g.logger.header(WarnLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	print(e.Context(g.context), args)
}

// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (g *GrpcLogger) Warningf(format string, args ...interface{}) {
	e := g.logger.header(WarnLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	e.Context(g.context).Msgf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (g *GrpcLogger) Error(args ...interface{}) {
	e := g.logger.header(ErrorLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	print(e.Context(g.context), args)
}

// Errorln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func (g *GrpcLogger) Errorln(args ...interface{}) {
	e := g.logger.header(ErrorLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	print(e.Context(g.context), args)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func (g *GrpcLogger) Errorf(format string, args ...interface{}) {
	e := g.logger.header(ErrorLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	e.Context(g.context).Msgf(format, args...)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (g *GrpcLogger) Fatal(args ...interface{}) {
	e := g.logger.header(FatalLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	print(e.Context(g.context), args)
}

// Fatalln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (g *GrpcLogger) Fatalln(args ...interface{}) {
	e := g.logger.header(FatalLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	print(e.Context(g.context), args)
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (g *GrpcLogger) Fatalf(format string, args ...interface{}) {
	e := g.logger.header(FatalLevel)
	if e == nil {
		return
	}
	if g.logger.Caller > 0 {
		e.caller(runtime.Caller(g.logger.Caller))
	}
	e.Context(g.context).Msgf(format, args...)
}

// V reports whether verbosity level l is at least the requested verbose leveg.
func (g *GrpcLogger) V(level int) bool {
	return level >= int(g.logger.Level)
}

type grpcLoggerV2 interface {
	Info(args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})
	Warning(args ...interface{})
	Warningln(args ...interface{})
	Warningf(format string, args ...interface{})
	Error(args ...interface{})
	Errorln(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalln(args ...interface{})
	Fatalf(format string, args ...interface{})
	V(l int) bool
}

var _ grpcLoggerV2 = (*GrpcLogger)(nil)
