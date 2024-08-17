package log

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a structured logger for the API.
type Logger struct {
	logger *zap.Logger
}

// NewLogger instantiates a new testing, development or production Logger.
func NewLogger(env string) (Logger, error) {
	var config zap.Config

	switch env {
	case "testing":
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel) // mute
	case "development":
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = shortTimeEncoder
	case "production":
		logFilePath, err := setupLogFile()
		if err != nil {
			return Logger{}, err
		}

		config = zap.Config{
			Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
			Development: false,
			Encoding:    "json",
			OutputPaths: []string{"stdout", logFilePath},
			EncoderConfig: zapcore.EncoderConfig{
				LevelKey:       "level",
				TimeKey:        "ts",
				NameKey:        "logger",
				CallerKey:      "caller",
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     utcTimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		}
	}

	logger, err := config.Build(zap.AddCallerSkip(1)) // skip log/logger.go
	if err != nil {
		return Logger{}, err
	}

	if len(config.OutputPaths) == 2 {
		absLogFilePath, err := filepath.Abs(config.OutputPaths[1])
		if err != nil {
			log.Fatalf("Failed to resolve filepath: %v", err)
		}
		logger.Info("logger initialized", zap.String("logFilePath", absLogFilePath))
	}

	return Logger{logger: logger}, nil
}

// Info logs an informational message.
func (l *Logger) Info(msg string, props ...zap.Field) {
	l.logger.Info(msg, props...)
}

// Error logs an error message.
func (l *Logger) Error(err error, props ...zap.Field) {
	fields := append(props, zap.Error(err))
	l.logger.Error(err.Error(), fields...)
}

// Fatal logs a fatal error message.
func (l *Logger) Fatal(err error, props ...zap.Field) {
	fields := append(props, zap.Error(err))
	l.logger.Fatal(err.Error(), fields...)
}

// Str creates a zap.Field with a map having a string value.
func Str(key, value string) zap.Field {
	return zap.String(key, value)
}

// Int creates a zap.Field with a map having an int value.
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// https://github.com/uber-go/zap/issues/661#issuecomment-520686037
func utcTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format("2006-01-02T15:04:05Z0700")) // e.g. 2019-08-13T04:39:11Z
}

func shortTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05"))
}

func setupLogFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	logFilePath := filepath.Join(home, ".n8n-shortlink", "n8n-shortlink.log")

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		file, err := os.Create(logFilePath)
		if err != nil {
			return "", err
		}
		file.Close()
	}

	return logFilePath, nil
}

// ReportEnvs reports source for environment variables
func (l *Logger) ReportEnvs() {
	envs := []string{
		"N8N_SHORTLINK_ENVIRONMENT",
		"N8N_SHORTLINK_HOST",
		"N8N_SHORTLINK_PORT",
		"N8N_SHORTLINK_RATE_LIMITER_ENABLED",
		"N8N_SHORTLINK_RATE_LIMITER_RPS",
		"N8N_SHORTLINK_RATE_LIMITER_BURST",
		"N8N_SHORTLINK_RATE_LIMITER_INACTIVITY",
	}

	availableEnvs := []string{}
	missingEnvs := []string{}

	for _, env := range envs {
		if value := os.Getenv(env); value == "" {
			missingEnvs = append(missingEnvs, env)
		} else {
			availableEnvs = append(availableEnvs, env)
		}
	}

	if len(missingEnvs) > 0 {
		l.Info("missing env vars, falling back to defaults", Str("missingEnvs", strings.Join(missingEnvs, ", ")))
		if len(availableEnvs) > 0 {
			l.Info("available env vars", Str("availableEnvs", strings.Join(availableEnvs, ", ")))
		}
	} else {
		l.Info("all env vars are available")
	}
}
