package log

import (
	"testing"
)

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

func TestGrpcLogger(t *testing.T) {
	logger := Logger{
		Level:  ParseLevel("debug"),
		Caller: 2,
	}

	var grpclog grpcLoggerV2 = GrpcLogger{logger}

	grpclog.Info("hello", "grpclog from json")
}
