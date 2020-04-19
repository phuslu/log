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

	var grpclog grpcLoggerV2 = &GrpcLogger{logger}

	osExit = func(int) {}

	grpclog.Info("hello", "grpclog Info message")
	grpclog.Infoln("hello", "grpclog Infoln message")
	grpclog.Infof("hello", "grpclog Infof message")
	grpclog.Warning("hello", "grpclog Warning message")
	grpclog.Warningln("hello", "grpclog Warningln message")
	grpclog.Warningf("hello", "grpclog Warningf message")
	grpclog.Error("hello", "grpclog Error message")
	grpclog.Errorln("hello", "grpclog Errorln message")
	grpclog.Errorf("hello", "grpclog Errorf message")
	grpclog.Fatal("hello", "grpclog Fatal message")
	grpclog.Fatalln("hello", "grpclog Fatalln message")
	grpclog.Fatalf("hello", "grpclog Fatalf message")

	if grpclog.V(0) {
		grpclog.Fatalf("hello", "grpclog debug level json")
	}
}

func TestGrpcLoggerLevel(t *testing.T) {
	var grpclog grpcLoggerV2 = &GrpcLogger{
		Logger: Logger{
			Level:  ParseLevel("warn"),
			Caller: 2,
		},
	}
	grpclog.Info("hello", "grpclog Info message")
}
