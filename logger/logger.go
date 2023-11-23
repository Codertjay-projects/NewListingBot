package logger

import (
	"NewListingBot/config"
	"context"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"

	"log"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger

var loggerFieldsAll loggerfields = "logger.fields"

type loggerfields string

// InitLogger init app logger
func InitLogger() {
	var err error
	logger, err = zap.NewDevelopment(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Println("error loading config")
	}
	// Initialize Sentry with your DSN
	err = sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		EnableTracing:    cfg.SentryEnableTracing,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		panic(err)
	}
}

// With adding fields to app logger
func With(ctx context.Context, fields ...zap.Field) context.Context {
	data := ctx.Value(loggerFieldsAll)
	var storedFields []zap.Field
	if data != nil {
		storedFields = data.([]zap.Field)
	}
	storedFields = append(storedFields, fields...)
	return context.WithValue(ctx, loggerFieldsAll, storedFields)
}

// Error error level on app logger
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	defer ShutdownSentry()

	data := ctx.Value(loggerFieldsAll)
	var storedFields []zap.Field
	if data != nil {
		storedFields = data.([]zap.Field)
	}
	storedFields = append(storedFields, fields...)

	logger.Error(msg, storedFields...)
	// Capture the error and send it to Sentry
	sentry.CaptureException(errors.New(fmt.Sprintf("The Error: %s ", msg)))

}

// Info info level to app logger
func Info(ctx context.Context, msg string, fields ...zap.Field) {

	defer ShutdownSentry()

	data := ctx.Value(loggerFieldsAll)

	var storedFields []zap.Field
	if data != nil {
		storedFields = data.([]zap.Field)
	}
	storedFields = append(storedFields, fields...)

	logger.Info(msg, storedFields...)

	sentry.CaptureMessage(fmt.Sprintf("The Info %s ", msg))
}

// ShutdownSentry should be called when your application exits to flush Sentry events
func ShutdownSentry() {
	sentry.Flush(2 * time.Second) // Adjust the timeout as needed
}
