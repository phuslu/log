package log

// GrpcLogger implements methods to satisfy interface
// google.golang.org/grpc/grpclog.LoggerV2.
type GrpcLogger struct {
	logger  Logger
	context Context
}

// Grpc wraps the Logger to provide a LoggerV2 logger
func (l *Logger) Grpc(context Context) (g *GrpcLogger) {
	g = &GrpcLogger{
		logger:  *l,
		context: context,
	}
	return
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (g *GrpcLogger) Info(args ...interface{}) {
	if g.logger.silent(InfoLevel) {
		return
	}
	e := g.logger.header(InfoLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgs(args...)
}

// Infoln logs to INFO log. Arguments are handled in the manner of fmt.Println.
func (g *GrpcLogger) Infoln(args ...interface{}) {
	if g.logger.silent(InfoLevel) {
		return
	}
	e := g.logger.header(InfoLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgs(args...)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func (g *GrpcLogger) Infof(format string, args ...interface{}) {
	if g.logger.silent(InfoLevel) {
		return
	}
	e := g.logger.header(InfoLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgf(format, args...)
}

// Warning logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (g *GrpcLogger) Warning(args ...interface{}) {
	if g.logger.silent(WarnLevel) {
		return
	}
	e := g.logger.header(WarnLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgs(args...)
}

// Warningln logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func (g *GrpcLogger) Warningln(args ...interface{}) {
	if g.logger.silent(WarnLevel) {
		return
	}
	e := g.logger.header(WarnLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgs(args...)
}

// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (g *GrpcLogger) Warningf(format string, args ...interface{}) {
	if g.logger.silent(WarnLevel) {
		return
	}
	e := g.logger.header(WarnLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (g *GrpcLogger) Error(args ...interface{}) {
	if g.logger.silent(ErrorLevel) {
		return
	}
	e := g.logger.header(ErrorLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgs(args...)
}

// Errorln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func (g *GrpcLogger) Errorln(args ...interface{}) {
	if g.logger.silent(ErrorLevel) {
		return
	}
	e := g.logger.header(ErrorLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgs(args...)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func (g *GrpcLogger) Errorf(format string, args ...interface{}) {
	if g.logger.silent(ErrorLevel) {
		return
	}
	e := g.logger.header(ErrorLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgf(format, args...)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (g *GrpcLogger) Fatal(args ...interface{}) {
	if g.logger.silent(FatalLevel) {
		return
	}
	e := g.logger.header(FatalLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgs(args...)
}

// Fatalln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (g *GrpcLogger) Fatalln(args ...interface{}) {
	if g.logger.silent(FatalLevel) {
		return
	}
	e := g.logger.header(FatalLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msgs(args...)
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (g *GrpcLogger) Fatalf(format string, args ...interface{}) {
	if g.logger.silent(FatalLevel) {
		return
	}
	e := g.logger.header(FatalLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
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
