// Package log provides a modern logging interface that can bridge between
// the legacy magicCommon log system and the standard library slog.
package log

import (
	"context"
	"log/slog"
	"os"

	"github.com/muidea/magicCommon/foundation/log"
)

// Logger is a unified logger interface that supports both legacy and modern logging
type Logger interface {
	Errorf(format string, args ...any)
	Infof(format string, args ...any)
	Debugf(format string, args ...any)
	Warnf(format string, args ...any)

	// Modern slog-style methods
	Error(msg string, args ...any)
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)

	// Context-aware methods
	ErrorContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}

// Config holds logging configuration
type Config struct {
	// UseSlog enables the standard library slog backend
	UseSlog bool
	// Level sets the minimum log level
	Level slog.Level
	// JSONFormat outputs logs in JSON format when using slog
	JSONFormat bool
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() Config {
	return Config{
		UseSlog:    true,
		Level:      slog.LevelInfo,
		JSONFormat: false,
	}
}

type loggerImpl struct {
	useSlog bool
	slog    *slog.Logger
}

// New creates a new logger with the given configuration
func New(cfg Config) Logger {
	if !cfg.UseSlog {
		return &legacyLogger{}
	}

	var handler slog.Handler
	if cfg.JSONFormat {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: cfg.Level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: cfg.Level,
		})
	}

	return &loggerImpl{
		useSlog: true,
		slog:    slog.New(handler),
	}
}

// Default returns a logger with default configuration
func Default() Logger {
	return New(DefaultConfig())
}

func (l *loggerImpl) Errorf(format string, args ...any) {
	l.slog.Error(format, args...)
}

func (l *loggerImpl) Infof(format string, args ...any) {
	l.slog.Info(format, args...)
}

func (l *loggerImpl) Debugf(format string, args ...any) {
	l.slog.Debug(format, args...)
}

func (l *loggerImpl) Warnf(format string, args ...any) {
	l.slog.Warn(format, args...)
}

func (l *loggerImpl) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

func (l *loggerImpl) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

func (l *loggerImpl) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

func (l *loggerImpl) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

func (l *loggerImpl) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.slog.ErrorContext(ctx, msg, args...)
}

func (l *loggerImpl) InfoContext(ctx context.Context, msg string, args ...any) {
	l.slog.InfoContext(ctx, msg, args...)
}

func (l *loggerImpl) DebugContext(ctx context.Context, msg string, args ...any) {
	l.slog.DebugContext(ctx, msg, args...)
}

func (l *loggerImpl) WarnContext(ctx context.Context, msg string, args ...any) {
	l.slog.WarnContext(ctx, msg, args...)
}

// legacyLogger wraps the old magicCommon log system for backward compatibility
type legacyLogger struct{}

func (l *legacyLogger) Errorf(format string, args ...any) {
	log.Errorf(format, args...)
}

func (l *legacyLogger) Infof(format string, args ...any) {
	log.Infof(format, args...)
}

func (l *legacyLogger) Debugf(format string, args ...any) {
	log.Debugf(format, args...)
}

func (l *legacyLogger) Warnf(format string, args ...any) {
	log.Warnf(format, args...)
}

func (l *legacyLogger) Error(msg string, args ...any) {
	log.Errorf(msg, args...)
}

func (l *legacyLogger) Info(msg string, args ...any) {
	log.Infof(msg, args...)
}

func (l *legacyLogger) Debug(msg string, args ...any) {
	log.Debugf(msg, args...)
}

func (l *legacyLogger) Warn(msg string, args ...any) {
	log.Warnf(msg, args...)
}

func (l *legacyLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	log.Errorf(msg, args...)
}

func (l *legacyLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	log.Infof(msg, args...)
}

func (l *legacyLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	log.Debugf(msg, args...)
}

func (l *legacyLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	log.Warnf(msg, args...)
}
