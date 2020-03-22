package log

import (
	"fmt"
)

// GrpcLogger implements methods to satisfy interface
// google.golang.org/grpc/grpclog.LoggerV2.
type GrpcLogger struct {
	Logger Logger
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (l GrpcLogger) Info(args ...interface{}) {
	l.Logger.Info().Msg(fmt.Sprint(args...))
}

// Infoln logs to INFO log. Arguments are handled in the manner of fmt.Println.
func (l GrpcLogger) Infoln(args ...interface{}) {
	l.Logger.Info().Msg(fmt.Sprintln(args...))
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func (l GrpcLogger) Infof(format string, args ...interface{}) {
	l.Logger.Info().Msgf(format, args...)
}

// Warning logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (l GrpcLogger) Warning(args ...interface{}) {
	l.Logger.Warn().Msg(fmt.Sprint(args...))
}

// Warningln logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func (l GrpcLogger) Warningln(args ...interface{}) {
	l.Logger.Warn().Msg(fmt.Sprintln(args...))
}

// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (l GrpcLogger) Warningf(format string, args ...interface{}) {
	l.Logger.Warn().Msgf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (l GrpcLogger) Error(args ...interface{}) {
	l.Logger.Error().Msg(fmt.Sprint(args...))
}

// Errorln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func (l GrpcLogger) Errorln(args ...interface{}) {
	l.Logger.Error().Msg(fmt.Sprintln(args...))
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func (l GrpcLogger) Errorf(format string, args ...interface{}) {
	l.Logger.Error().Msgf(format, args...)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l GrpcLogger) Fatal(args ...interface{}) {
	l.Logger.Fatal().Msg(fmt.Sprint(args...))
}

// Fatalln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l GrpcLogger) Fatalln(args ...interface{}) {
	l.Logger.Fatal().Msg(fmt.Sprintln(args...))
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l GrpcLogger) Fatalf(format string, args ...interface{}) {
	l.Logger.Fatal().Msgf(format, args...)
}

// V reports whether verbosity level l is at least the requested verbose level.
func (l GrpcLogger) V(level int) bool {
	return level >= int(l.Logger.Level)
}
