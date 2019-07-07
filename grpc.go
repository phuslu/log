package log

import (
	"fmt"
)

type GrpcLogger struct {
	JSONLogger JSONLogger
}

func (l GrpcLogger) Info(args ...interface{}) {
	l.JSONLogger.Info().Msg(fmt.Sprint(args...))
}

func (l GrpcLogger) Infoln(args ...interface{}) {
	l.JSONLogger.Info().Msg(fmt.Sprintln(args...))
}

func (l GrpcLogger) Infof(format string, args ...interface{}) {
	l.JSONLogger.Info().Msgf(format, args...)
}

func (l GrpcLogger) Warning(args ...interface{}) {
	l.JSONLogger.Warn().Msg(fmt.Sprint(args...))
}

func (l GrpcLogger) Warningln(args ...interface{}) {
	l.JSONLogger.Warn().Msg(fmt.Sprintln(args...))
}

func (l GrpcLogger) Warningf(format string, args ...interface{}) {
	l.JSONLogger.Info().Msgf(format, args...)
}

func (l GrpcLogger) Error(args ...interface{}) {
	l.JSONLogger.Error().Msg(fmt.Sprint(args...))
}
func (l GrpcLogger) Errorln(args ...interface{}) {
	l.JSONLogger.Error().Msg(fmt.Sprintln(args...))
}

func (l GrpcLogger) Errorf(format string, args ...interface{}) {
	l.JSONLogger.Info().Msgf(format, args...)
}

func (l GrpcLogger) Fatal(args ...interface{}) {
	l.JSONLogger.Fatal().Msg(fmt.Sprint(args...))
}
func (l GrpcLogger) Fatalln(args ...interface{}) {
	l.JSONLogger.Fatal().Msg(fmt.Sprintln(args...))
}

func (l GrpcLogger) Fatalf(format string, args ...interface{}) {
	l.JSONLogger.Info().Msgf(format, args...)
}

func (l GrpcLogger) V(level int) bool {
	return level >= int(l.JSONLogger.Level)
}
