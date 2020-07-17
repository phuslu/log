package log

// GrpcLogger implements methods to satisfy interface
// google.golang.org/grpc/grpclog.LoggerV2.
type GrpcLogger struct {
	logger  Logger
	level   Level
	context Context
}

// Grpc wraps the Logger to provide a more ergonomic, but slightly slower,
// API. Grpcing a Logger is quite inexpensive, so it's reasonable for a
// single application to use both Loggers and GrpcLoggers, converting
// between them on the boundaries of performance-sensitive code.
func (l *Logger) Grpc(context Context) (logger *GrpcLogger) {
	logger = &GrpcLogger{
		logger:  *l,
		context: context,
	}
	if logger.logger.Caller > 0 {
		logger.logger.Caller += 1
	}
	return
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (l *GrpcLogger) Info(args ...interface{}) {
	e := l.logger.WithLevel(InfoLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	print(e, args)
}

// Infoln logs to INFO log. Arguments are handled in the manner of fmt.Println.
func (l *GrpcLogger) Infoln(args ...interface{}) {
	e := l.logger.WithLevel(InfoLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	print(e, args)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func (l *GrpcLogger) Infof(format string, args ...interface{}) {
	e := l.logger.WithLevel(InfoLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	e.Msgf(format, args...)
}

// Warning logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (l *GrpcLogger) Warning(args ...interface{}) {
	e := l.logger.WithLevel(WarnLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	print(e, args)
}

// Warningln logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func (l *GrpcLogger) Warningln(args ...interface{}) {
	e := l.logger.WithLevel(WarnLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	print(e, args)
}

// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (l *GrpcLogger) Warningf(format string, args ...interface{}) {
	e := l.logger.WithLevel(WarnLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	e.Msgf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (l *GrpcLogger) Error(args ...interface{}) {
	e := l.logger.WithLevel(ErrorLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	print(e, args)
}

// Errorln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func (l *GrpcLogger) Errorln(args ...interface{}) {
	e := l.logger.WithLevel(ErrorLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	print(e, args)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func (l *GrpcLogger) Errorf(format string, args ...interface{}) {
	e := l.logger.WithLevel(ErrorLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	e.Msgf(format, args...)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *GrpcLogger) Fatal(args ...interface{}) {
	e := l.logger.WithLevel(FatalLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	print(e, args)
}

// Fatalln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *GrpcLogger) Fatalln(args ...interface{}) {
	e := l.logger.WithLevel(FatalLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	print(e, args)
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *GrpcLogger) Fatalf(format string, args ...interface{}) {
	e := l.logger.WithLevel(FatalLevel)
	if len(l.context) != 0 {
		e.Context(l.context)
	}
	e.Msgf(format, args...)
}

// V reports whether verbosity level l is at least the requested verbose level.
func (l *GrpcLogger) V(level int) bool {
	return level >= int(l.logger.Level)
}
