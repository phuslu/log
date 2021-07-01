package log

import "runtime"

// GrcpGatewayLogger implements methods to satisfy interface
// github.com/grpc-ecosystem/go-grpc-middleware/blob/v2/interceptors/logging/logging.go
type GrcpGatewayLogger struct {
	logger  Logger
	context Context
}

// GrcpGateway wraps the Logger to provide a GrcpGateway logger
func (l *Logger) GrcpGateway() GrcpGatewayLogger {
	return GrcpGatewayLogger{
		logger: *l,
	}
}

// WithValues adds some key-value pairs of context to a logger.
// See Info for documentation on how key/value pairs work.
func (g GrcpGatewayLogger) WithValues(keysAndValues ...interface{}) GrcpGatewayLogger {
	g.context = append(g.context, NewContext(nil).KeysAndValues(keysAndValues...).Value()...)
	return g
}

// Debug logs a debug with the message and key/value pairs as context.
func (g GrcpGatewayLogger) Debug(msg string) {
	if g.logger.silent(DebugLevel) {
		return
	}
	e := g.logger.header(DebugLevel)
	if g.logger.Caller > 0 {
		_, file, line, _ := runtime.Caller(g.logger.Caller)
		e.caller(file, line, g.logger.FullpathCaller)
	}
	e.Context(g.context).Msg(msg)
}

// Info logs an info with the message and key/value pairs as context.
func (g GrcpGatewayLogger) Info(msg string) {
	if g.logger.silent(InfoLevel) {
		return
	}
	e := g.logger.header(InfoLevel)
	if g.logger.Caller > 0 {
		_, file, line, _ := runtime.Caller(g.logger.Caller)
		e.caller(file, line, g.logger.FullpathCaller)
	}
	e.Context(g.context).Msg(msg)
}

// Warning logs a warning with the message and key/value pairs as context.
func (g GrcpGatewayLogger) Warning(msg string) {
	if g.logger.silent(WarnLevel) {
		return
	}
	e := g.logger.header(WarnLevel)
	if g.logger.Caller > 0 {
		_, file, line, _ := runtime.Caller(g.logger.Caller)
		e.caller(file, line, g.logger.FullpathCaller)
	}
	e.Context(g.context).Msg(msg)
}

// Error logs an error with the message and key/value pairs as context.
func (g GrcpGatewayLogger) Error(msg string) {
	if g.logger.silent(ErrorLevel) {
		return
	}
	e := g.logger.header(ErrorLevel)
	if g.logger.Caller > 0 {
		_, file, line, _ := runtime.Caller(g.logger.Caller)
		e.caller(file, line, g.logger.FullpathCaller)
	}
	e.Context(g.context).Msg(msg)
}
