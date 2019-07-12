package log

import (
	"testing"

	"google.golang.org/grpc/grpclog"
)

func TestGrpcLogger(t *testing.T) {
	log := Logger{
		Level: ParseLevel("debug"),
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

	grpclog.Print("hello grpclog Print")
	grpclog.Println("hello grpclog Println")
	grpclog.Printf("hello grpclog %s", "Printf")
	grpclog.Info("hello grpclog Info")
	grpclog.Infoln("hello grpclog Infoln")
	grpclog.Infof("hello grpclog %s", "Infof")
	grpclog.Warning("hello grpclog Warning")
	grpclog.Warningln("hello grpclog Warningln")
	grpclog.Warningf("hello grpclog %s", "Warningf")
	grpclog.Error("hello grpclog Error")
	grpclog.Errorln("hello grpclog Errorln")
	grpclog.Errorf("hello grpclog %s", "Errorf")
	// grpclog.Fatal("hello grpclog Fatal")
	// grpclog.Fatalln("hello grpclog Fatalln")
	// grpclog.Fatalf("hello grpclog %s", "Fatalf")

	if grpclog.V(0) {
		grpclog.Printf("hello grpclog V")
	}
}
