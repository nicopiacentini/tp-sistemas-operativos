package utils_general

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

var logger *slog.Logger

func ConfigurarLogger(nombre string) {
	logFile, err := os.OpenFile(nombre, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)

	handler := slog.NewTextHandler(mw, &slog.HandlerOptions{
		// Configuraciones(opcional)
	})
	logger = slog.New(handler)
}

func LoggearMensaje(level string, format string, args ...interface{}) {
	if logger == nil {
		panic("Logger no configurado. Llama a ConfigurarLogger primero.")
	}
	message := fmt.Sprintf(format, args...)
	switch level {
	case "INFO":
		logger.Info(message)
	case "WARN":
		logger.Warn(message)
	case "ERROR":
		logger.Error(message)
	case "DEBUG":
		logger.Debug(message)
	default:
		logger.Info(message)
	}
}
