package log

import (
	"testing"
)

type grpcGatewayLogger interface {
	WithValues(keysAndValues ...interface{}) GrpcGatewayLogger
	Debug(msg string)
	Info(msg string)
	Warning(msg string)
	Error(msg string)
}

func TestGrpcGatewayLoggerNil(t *testing.T) {
	var gglog grpcGatewayLogger = DefaultLogger.GrpcGateway()

	gglog.Debug("Grpc Gateway Debug")
	gglog.Info("Grpc Gateway Info")
	gglog.Warning("Grpc Gateway Warning")
	gglog.Error("Grpc Gateway Error")

	gglog = gglog.WithValues("a_key", "a_value")
	gglog.Debug("Grpc Gateway WithValues test")
	gglog.Info("Grpc Gateway WithValues test")
	gglog.Warning("Grpc Gateway WithValues test")
	gglog.Error("Grpc Gateway WithValues test")
}

func TestGrpcGatewayLogger(t *testing.T) {
	DefaultLogger.Caller = -1
	DefaultLogger.Writer = &ConsoleWriter{ColorOutput: true, EndWithMessage: true}

	var gglog grpcGatewayLogger = DefaultLogger.GrpcGateway()

	gglog.Debug("Grpc Gateway Debug")
	gglog.Info("Grpc Gateway Info")
	gglog.Warning("Grpc Gateway Warning")
	gglog.Error("Grpc Gateway Error")

	gglog = gglog.WithValues("a_key", "a_value")
	gglog.Debug("Grpc Gateway WithValues test")
	gglog.Info("Grpc Gateway WithValues test")
	gglog.Warning("Grpc Gateway WithValues test")
	gglog.Error("Grpc Gateway WithValues test")
}

func TestGrpcGatewayLoggerLevel(t *testing.T) {
	DefaultLogger.Caller = 1
	DefaultLogger.Writer = &ConsoleWriter{ColorOutput: true, EndWithMessage: true}
	DefaultLogger.Level = noLevel

	var gglog grpcGatewayLogger = DefaultLogger.GrpcGateway()

	gglog.Debug("Grpc Gateway Debug")
	gglog.Info("Grpc Gateway Info")
	gglog.Warning("Grpc Gateway Warning")
	gglog.Error("Grpc Gateway Error")

	gglog = gglog.WithValues("a_key", "a_value")
	gglog.Debug("Grpc Gateway WithValues test")
	gglog.Info("Grpc Gateway WithValues test")
	gglog.Warning("Grpc Gateway WithValues test")
	gglog.Error("Grpc Gateway WithValues test")
}

func TestGrpcGatewayLoggerChangingValues(t *testing.T) {
	var gglog grpcGatewayLogger = DefaultLogger.GrpcGateway()

	gglog.Info("Grpc Gateway Info")

	gglogUnique := gglog
	gglogUnique.WithValues("a_key", "a_value").Info("Grpc Gateway WithValues test")
	gglogUnique.WithValues("b_key", "b_value").Info("Grpc Gateway WithValues test")

	gglogAcumulate := gglog
	gglogAcumulate = gglogAcumulate.WithValues("a_key", "a_value")
	gglogAcumulate = gglogAcumulate.WithValues("b_key", "b_value")
	gglogAcumulate.Info("Grpc Gateway WithValues test")
}
