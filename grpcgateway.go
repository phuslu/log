package log

// GrpcGatewayLogger implements methods to satisfy interface
// github.com/grpc-ecosystem/go-grpc-middleware/blob/v2/interceptors/logging/logging.go
type GrpcGatewayLogger struct {
	logger  Logger
	context Context
}

// GrpcGateway wraps the Logger to provide a GrpcGateway logger
func (l *Logger) GrpcGateway() GrpcGatewayLogger {
	return GrpcGatewayLogger{
		logger: *l,
	}
}

// WithValues adds some key-value pairs of context to a logger.
// See Info for documentation on how key/value pairs work.
func (g GrpcGatewayLogger) WithValues(keysAndValues ...interface{}) GrpcGatewayLogger {
	g.context = append(g.context, NewContext(nil).KeysAndValues(keysAndValues...).Value()...)
	return g
}

// Debug logs a debug with the message and key/value pairs as context.
func (g GrpcGatewayLogger) Debug(msg string) {
	if g.logger.silent(DebugLevel) {
		return
	}
	e := g.logger.header(DebugLevel)
	if caller, full := g.logger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var rpc [1]uintptr
		e.caller(callers(caller, rpc[:]), rpc[:], full)
	}
	e.Context(g.context).Msg(msg)
}

// Info logs an info with the message and key/value pairs as context.
func (g GrpcGatewayLogger) Info(msg string) {
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
	e.Context(g.context).Msg(msg)
}

// Warning logs a warning with the message and key/value pairs as context.
func (g GrpcGatewayLogger) Warning(msg string) {
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
	e.Context(g.context).Msg(msg)
}

// Error logs an error with the message and key/value pairs as context.
func (g GrpcGatewayLogger) Error(msg string) {
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
	e.Context(g.context).Msg(msg)
}
