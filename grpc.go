package log

import (
	"fmt"
)

type GrpcLogger struct {
	Logger Logger
}

func (l GrpcLogger) Info(args ...interface{}) {
	l.Logger.Info().Msg(fmt.Sprint(args...))
}

func (l GrpcLogger) Infoln(args ...interface{}) {
	l.Logger.Info().Msg(fmt.Sprintln(args...))
}

func (l GrpcLogger) Infof(format string, args ...interface{}) {
	l.Logger.Info().Msgf(format, args...)
}

func (l GrpcLogger) Warning(args ...interface{}) {
	l.Logger.Warn().Msg(fmt.Sprint(args...))
}

func (l GrpcLogger) Warningln(args ...interface{}) {
	l.Logger.Warn().Msg(fmt.Sprintln(args...))
}

func (l GrpcLogger) Warningf(format string, args ...interface{}) {
	l.Logger.Warn().Msgf(format, args...)
}

func (l GrpcLogger) Error(args ...interface{}) {
	l.Logger.Error().Msg(fmt.Sprint(args...))
}
func (l GrpcLogger) Errorln(args ...interface{}) {
	l.Logger.Error().Msg(fmt.Sprintln(args...))
}

func (l GrpcLogger) Errorf(format string, args ...interface{}) {
	l.Logger.Error().Msgf(format, args...)
}

func (l GrpcLogger) Fatal(args ...interface{}) {
	l.Logger.Fatal().Msg(fmt.Sprint(args...))
}
func (l GrpcLogger) Fatalln(args ...interface{}) {
	l.Logger.Fatal().Msg(fmt.Sprintln(args...))
}

func (l GrpcLogger) Fatalf(format string, args ...interface{}) {
	l.Logger.Fatal().Msgf(format, args...)
}

func (l GrpcLogger) V(level int) bool {
	return level >= int(l.Logger.Level)
}
