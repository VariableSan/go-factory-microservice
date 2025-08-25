package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
	serviceName string
}

type Config struct {
	ServiceName string
	Level       string
	Format      string // "json" or "text"
}

// NewLogger creates a new structured logger with service name
func NewLogger(config Config) *Logger {
	level := parseLogLevel(config.Level)
	
	var handler slog.Handler
	
	opts := &slog.HandlerOptions{
		Level: level,
	}
	
	if strings.ToLower(config.Format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	
	logger := slog.New(handler).With(
		"service", config.ServiceName,
	)
	
	return &Logger{
		Logger:      logger,
		serviceName: config.ServiceName,
	}
}

// WithContext adds contextual fields to the logger
func (l *Logger) WithContext(attrs ...slog.Attr) *Logger {
	return &Logger{
		Logger:      l.Logger.With(attrsToAny(attrs)...),
		serviceName: l.serviceName,
	}
}

// WithComponent adds component field to the logger
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger:      l.Logger.With("component", component),
		serviceName: l.serviceName,
	}
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func attrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs)*2)
	for i, attr := range attrs {
		result[i*2] = attr.Key
		result[i*2+1] = attr.Value
	}
	return result
}
