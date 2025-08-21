package logger

import (
	"context"
	"io"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger with additional functionality
type Logger struct {
	*logrus.Logger
}

// New creates a new logger instance
func New(level string, output io.Writer) *Logger {
	logger := logrus.New()

	// Set output
	if output != nil {
		logger.SetOutput(output)
	} else {
		logger.SetOutput(os.Stdout)
	}

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	// Set formatter
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	return &Logger{Logger: logger}
}

// WithTracing adds tracing context to log entries
func (l *Logger) WithTracing(ctx context.Context) *logrus.Entry {
	entry := l.WithContext(ctx)
	
	if span := opentracing.SpanFromContext(ctx); span != nil {
		if spanContext, ok := span.Context().(interface {
			TraceID() string
			SpanID() string
		}); ok {
			entry = entry.WithFields(logrus.Fields{
				"trace_id": spanContext.TraceID(),
				"span_id":  spanContext.SpanID(),
			})
		}
	}
	
	return entry
}

// WithService adds service name to log entries
func (l *Logger) WithService(serviceName string) *logrus.Entry {
	return l.WithField("service", serviceName)
}

// WithRequestID adds request ID to log entries
func (l *Logger) WithRequestID(requestID string) *logrus.Entry {
	return l.WithField("request_id", requestID)
}

// WithError adds error to log entries
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}
