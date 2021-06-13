package log

import (
	"testing"
)

type grcpGatewayLogger interface {
	WithValues(keysAndValues ...interface{}) GrcpGatewayLogger
	Debug(msg string)
	Info(msg string)
	Warning(msg string)
	Error(msg string)
}

func TestGrcpGatewayLoggerNil(t *testing.T) {
	var gglog grcpGatewayLogger = DefaultLogger.GrcpGateway()

	gglog.Debug("Grcp Gateway Debug")
	gglog.Info("Grcp Gateway Info")
	gglog.Warning("Grcp Gateway Warning")
	gglog.Error("Grcp Gateway Error")

	gglog = gglog.WithValues("a_key", "a_value")
	gglog.Debug("Grcp Gateway WithValues test")
	gglog.Info("Grcp Gateway WithValues test")
	gglog.Warning("Grcp Gateway WithValues test")
	gglog.Error("Grcp Gateway WithValues test")
}

func TestGrcpGatewayLogger(t *testing.T) {
	DefaultLogger.Caller = 1
	DefaultLogger.Writer = &ConsoleWriter{ColorOutput: true, EndWithMessage: true}

	var gglog grcpGatewayLogger = DefaultLogger.GrcpGateway()

	gglog.Debug("Grcp Gateway Debug")
	gglog.Info("Grcp Gateway Info")
	gglog.Warning("Grcp Gateway Warning")
	gglog.Error("Grcp Gateway Error")

	gglog = gglog.WithValues("a_key", "a_value")
	gglog.Debug("Grcp Gateway WithValues test")
	gglog.Info("Grcp Gateway WithValues test")
	gglog.Warning("Grcp Gateway WithValues test")
	gglog.Error("Grcp Gateway WithValues test")
}

func TestGrcpGatewayLoggerLevel(t *testing.T) {
	DefaultLogger.Caller = 1
	DefaultLogger.Writer = &ConsoleWriter{ColorOutput: true, EndWithMessage: true}
	DefaultLogger.Level = noLevel

	var gglog grcpGatewayLogger = DefaultLogger.GrcpGateway()

	gglog.Debug("Grcp Gateway Debug")
	gglog.Info("Grcp Gateway Info")
	gglog.Warning("Grcp Gateway Warning")
	gglog.Error("Grcp Gateway Error")

	gglog = gglog.WithValues("a_key", "a_value")
	gglog.Debug("Grcp Gateway WithValues test")
	gglog.Info("Grcp Gateway WithValues test")
	gglog.Warning("Grcp Gateway WithValues test")
	gglog.Error("Grcp Gateway WithValues test")
}

func TestGrcpGatewayLoggerChangingValues(t *testing.T) {
	var gglog grcpGatewayLogger = DefaultLogger.GrcpGateway()

	gglog.Info("Grcp Gateway Info")

	gglogUnique := gglog
	gglogUnique.WithValues("a_key", "a_value").Info("Grcp Gateway WithValues test")
	gglogUnique.WithValues("b_key", "b_value").Info("Grcp Gateway WithValues test")

	gglogAcumulate := gglog
	gglogAcumulate = gglogAcumulate.WithValues("a_key", "a_value")
	gglogAcumulate = gglogAcumulate.WithValues("b_key", "b_value")
	gglogAcumulate.Info("Grcp Gateway WithValues test")
}
