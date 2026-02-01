package monitoring

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// LogLevel represents logging level
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns string representation of log level
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses log level from string
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// ValidationLogger provides structured logging for validation system
type ValidationLogger struct {
	mu       sync.RWMutex
	logger   *log.Logger
	minLevel LogLevel
	output   io.Writer
	fields   map[string]interface{}
	metrics  *MetricsCollector
}

// NewValidationLogger creates a new validation logger
func NewValidationLogger(level string, output io.Writer) *ValidationLogger {
	if output == nil {
		output = os.Stdout
	}

	return &ValidationLogger{
		logger:   log.New(output, "", 0), // No prefix, we'll format ourselves
		minLevel: ParseLogLevel(level),
		output:   output,
		fields:   make(map[string]interface{}),
	}
}

// WithMetrics attaches metrics collector to logger
func (l *ValidationLogger) WithMetrics(metrics *MetricsCollector) *ValidationLogger {
	l.metrics = metrics
	return l
}

// WithField adds a field to all subsequent log entries
func (l *ValidationLogger) WithField(key string, value interface{}) *ValidationLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fields[key] = value
	return l
}

// WithFields adds multiple fields to all subsequent log entries
func (l *ValidationLogger) WithFields(fields map[string]interface{}) *ValidationLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, v := range fields {
		l.fields[k] = v
	}
	return l
}

// Debug logs a debug message
func (l *ValidationLogger) Debug(msg string, fields ...map[string]interface{}) {
	l.log(LevelDebug, msg, nil, fields...)
}

// Info logs an info message
func (l *ValidationLogger) Info(msg string, fields ...map[string]interface{}) {
	l.log(LevelInfo, msg, nil, fields...)
}

// Warn logs a warning message
func (l *ValidationLogger) Warn(msg string, fields ...map[string]interface{}) {
	l.log(LevelWarn, msg, nil, fields...)
}

// Error logs an error message
func (l *ValidationLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	l.log(LevelError, msg, err, fields...)
}

// Fatal logs a fatal message and exits
func (l *ValidationLogger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	l.log(LevelFatal, msg, err, fields...)
	os.Exit(1)
}

// LogValidation logs a validation operation
func (l *ValidationLogger) LogValidation(
	operation string,
	modelName string,
	scenario string,
	duration time.Duration,
	err error,
	fields ...map[string]interface{},
) {
	entryFields := map[string]interface{}{
		"operation": operation,
		"model":     modelName,
		"scenario":  scenario,
		"duration":  duration.String(),
	}

	if err != nil {
		entryFields["error"] = err.Error()
		l.Error(fmt.Sprintf("Validation failed: %s", operation), err, mergeFields(entryFields, fields...))
	} else {
		l.Info(fmt.Sprintf("Validation succeeded: %s", operation), mergeFields(entryFields, fields...))
	}

	// Record metrics if available
	if l.metrics != nil {
		l.metrics.RecordValidation(duration, err)
	}
}

// LogCacheAccess logs cache access
func (l *ValidationLogger) LogCacheAccess(
	operation string,
	cacheType string,
	key string,
	hit bool,
	duration time.Duration,
	fields ...map[string]interface{},
) {
	entryFields := map[string]interface{}{
		"cache_operation": operation,
		"cache_type":      cacheType,
		"cache_key":       key,
		"cache_hit":       hit,
		"cache_duration":  duration.String(),
	}

	level := LevelDebug
	msg := fmt.Sprintf("Cache %s: %s", operation, cacheType)

	if hit {
		msg += " (hit)"
	} else {
		msg += " (miss)"
	}

	l.log(level, msg, nil, mergeFields(entryFields, fields...))

	// Record metrics if available
	if l.metrics != nil {
		if hit {
			l.metrics.RecordCacheHit()
		} else {
			l.metrics.RecordCacheMiss()
		}
	}
}

// LogLayerPerformance logs layer performance
func (l *ValidationLogger) LogLayerPerformance(
	layer string,
	operation string,
	duration time.Duration,
	success bool,
	fields ...map[string]interface{},
) {
	entryFields := map[string]interface{}{
		"layer":     layer,
		"operation": operation,
		"duration":  duration.String(),
		"success":   success,
	}

	level := LevelDebug
	msg := fmt.Sprintf("Layer %s: %s", layer, operation)

	if !success {
		level = LevelWarn
		msg += " (failed)"
	}

	l.log(level, msg, nil, mergeFields(entryFields, fields...))

	// Record metrics if available
	if l.metrics != nil {
		l.metrics.RecordLayerTime(layer, duration)
	}
}

// GetMetrics returns current metrics if available
func (l *ValidationLogger) GetMetrics() (*Metrics, error) {
	if l.metrics == nil {
		return nil, fmt.Errorf("metrics collector not attached")
	}
	metrics := l.metrics.GetMetrics()
	return &metrics, nil
}

// ResetMetrics resets metrics if available
func (l *ValidationLogger) ResetMetrics() error {
	if l.metrics == nil {
		return fmt.Errorf("metrics collector not attached")
	}
	l.metrics.Reset()
	return nil
}

// SetLevel changes the minimum log level
func (l *ValidationLogger) SetLevel(level string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = ParseLogLevel(level)
}

// GetLevel returns current log level
func (l *ValidationLogger) GetLevel() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.minLevel.String()
}

// Private methods

func (l *ValidationLogger) log(level LogLevel, msg string, err error, fields ...map[string]interface{}) {
	if level < l.minLevel {
		return
	}

	l.mu.RLock()
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    mergeFields(l.fields, fields...),
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.mu.RUnlock()

	// Format and output
	formatted := l.formatEntry(entry)
	l.logger.Println(formatted)
}

func (l *ValidationLogger) formatEntry(entry LogEntry) string {
	// Check if output is a terminal
	if file, ok := l.output.(*os.File); ok {
		stat, _ := file.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			// Terminal output - use human readable format
			return l.formatHumanReadable(entry)
		}
	}

	// Non-terminal output - use JSON
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple format
		return fmt.Sprintf("%s [%s] %s", entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
	}
	return string(data)
}

func (l *ValidationLogger) formatHumanReadable(entry LogEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
	level := entry.Level.String()

	var color string
	switch entry.Level {
	case LevelDebug:
		color = "\033[90m" // Gray
	case LevelInfo:
		color = "\033[36m" // Cyan
	case LevelWarn:
		color = "\033[33m" // Yellow
	case LevelError, LevelFatal:
		color = "\033[31m" // Red
	default:
		color = "\033[0m" // Reset
	}

	reset := "\033[0m"

	// Build message
	msg := fmt.Sprintf("%s %s[%s]%s %s", timestamp, color, level, reset, entry.Message)

	// Add fields
	if len(entry.Fields) > 0 {
		fieldStrs := make([]string, 0, len(entry.Fields))
		for k, v := range entry.Fields {
			fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", k, v))
		}
		msg += " " + strings.Join(fieldStrs, " ")
	}

	// Add error if present
	if entry.Error != "" {
		msg += fmt.Sprintf(" error=%s", entry.Error)
	}

	return msg
}

func mergeFields(base map[string]interface{}, additional ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base fields
	for k, v := range base {
		result[k] = v
	}

	// Merge additional fields
	for _, fields := range additional {
		for k, v := range fields {
			result[k] = v
		}
	}

	return result
}

// DefaultLogger returns a default validation logger
func DefaultLogger() *ValidationLogger {
	return NewValidationLogger("info", os.Stdout)
}

// FileLogger creates a logger that writes to a file
func FileLogger(level string, filepath string) (*ValidationLogger, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return NewValidationLogger(level, file), nil
}

// MultiLogger creates a logger that writes to multiple outputs
func MultiLogger(level string, outputs ...io.Writer) *ValidationLogger {
	multiWriter := io.MultiWriter(outputs...)
	return NewValidationLogger(level, multiWriter)
}
