package log

import (
	"testing"

	"google.golang.org/grpc/grpclog"
)

func TestGrpcLogger(t *testing.T) {
	log := Logger{
		Level:      ParseLevel("debug"),
		EscapeHTML: false,
		Writer: &Writer{
			LocalTime: true,
		},
	}

	grpclog.SetLoggerV2(GrpcLogger{log})

	grpclog.Println("hello", "grpclog from json")
}

func TestGrpcLogger2(t *testing.T) {
	log := GlogLogger{
		Level:  ParseLevel("debug"),
		Writer: &Writer{},
	}

	grpclog.SetLoggerV2(log)

	grpclog.Print("hello grpclog from glog")
}
