package log

import (
	"testing"
)

func TestGrpcLogger(t *testing.T) {
	DefaultLogger.Caller = -1
	DefaultLogger.Writer = &ConsoleWriter{ColorOutput: true, EndWithMessage: true}

	var grpclog grpcLoggerV2 = DefaultLogger.Grpc(NewContext(nil).Str("tag", "hi sugar").Value())

	notTest = false

	grpclog.Info("hello", "grpclog Info message")
	grpclog.Infoln("hello", "grpclog Infoln message")
	grpclog.Infof("hello %s", "grpclog Infof message")
	grpclog.Warning("hello", "grpclog Warning message")
	grpclog.Warningln("hello", "grpclog Warningln message")
	grpclog.Warningf("hello %s", "grpclog Warningf message")
	grpclog.Error("hello", "grpclog Error message")
	grpclog.Errorln("hello", "grpclog Errorln message")
	grpclog.Errorf("hello %s", "grpclog Errorf message")
	grpclog.Fatal("hello", "grpclog Fatal message")
	grpclog.Fatalln("hello", "grpclog Fatalln message")
	grpclog.Fatalf("hello %s", "grpclog Fatalf message")

	if grpclog.V(0) {
		grpclog.Fatalf("hello %s", "grpclog debug level json")
	}
}

func TestGrpcLoggerLevel(t *testing.T) {
	DefaultLogger.Caller = 1
	DefaultLogger.Writer = &ConsoleWriter{ColorOutput: true, EndWithMessage: true}
	DefaultLogger.Level = noLevel

	var grpclog grpcLoggerV2 = DefaultLogger.Grpc(NewContext(nil).Str("tag", "hi sugar").Value())

	notTest = false

	grpclog.Info("hello", "grpclog Info message")
	grpclog.Infoln("hello", "grpclog Infoln message")
	grpclog.Infof("hello %s", "grpclog Infof message")
	grpclog.Warning("hello", "grpclog Warning message")
	grpclog.Warningln("hello", "grpclog Warningln message")
	grpclog.Warningf("hello %s", "grpclog Warningf message")
	grpclog.Error("hello", "grpclog Error message")
	grpclog.Errorln("hello", "grpclog Errorln message")
	grpclog.Errorf("hello %s", "grpclog Errorf message")
	grpclog.Fatal("hello", "grpclog Fatal message")
	grpclog.Fatalln("hello", "grpclog Fatalln message")
	grpclog.Fatalf("hello %s", "grpclog Fatalf message")
}
