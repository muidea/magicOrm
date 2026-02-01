package monitoring

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// MetricsExporter exports validation metrics via HTTP
type MetricsExporter struct {
	mu           sync.RWMutex
	metrics      *MetricsCollector
	logger       *ValidationLogger
	server       *http.Server
	exportConfig ExportConfig
	customLabels map[string]string
}

// ExportConfig holds exporter configuration
type ExportConfig struct {
	Enabled          bool          `json:"enabled"`
	Port             int           `json:"port"`
	Path             string        `json:"path"`
	HealthCheckPath  string        `json:"health_check_path"`
	MetricsPath      string        `json:"metrics_path"`
	EnablePrometheus bool          `json:"enable_prometheus"`
	EnableJSON       bool          `json:"enable_json"`
	RefreshInterval  time.Duration `json:"refresh_interval"`
	Timeout          time.Duration `json:"timeout"`
}

// DefaultExportConfig returns default export configuration
func DefaultExportConfig() ExportConfig {
	return ExportConfig{
		Enabled:          true,
		Port:             9090,
		Path:             "/metrics",
		HealthCheckPath:  "/health",
		MetricsPath:      "/metrics/json",
		EnablePrometheus: true,
		EnableJSON:       true,
		RefreshInterval:  30 * time.Second,
		Timeout:          10 * time.Second,
	}
}

// NewMetricsExporter creates a new metrics exporter
func NewMetricsExporter(metrics *MetricsCollector, logger *ValidationLogger, config ExportConfig) *MetricsExporter {
	if logger == nil {
		logger = DefaultLogger()
	}

	return &MetricsExporter{
		metrics:      metrics,
		logger:       logger,
		exportConfig: config,
		customLabels: make(map[string]string),
	}
}

// Start starts the metrics exporter server
func (e *MetricsExporter) Start() error {
	if !e.exportConfig.Enabled {
		e.logger.Info("Metrics exporter is disabled")
		return nil
	}

	mux := http.NewServeMux()

	// Register handlers
	if e.exportConfig.EnablePrometheus {
		mux.HandleFunc(e.exportConfig.Path, e.prometheusHandler)
	}

	if e.exportConfig.EnableJSON {
		mux.HandleFunc(e.exportConfig.MetricsPath, e.jsonHandler)
	}

	mux.HandleFunc(e.exportConfig.HealthCheckPath, e.healthHandler)
	mux.HandleFunc("/", e.defaultHandler)

	e.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", e.exportConfig.Port),
		Handler:      mux,
		ReadTimeout:  e.exportConfig.Timeout,
		WriteTimeout: e.exportConfig.Timeout,
	}

	e.logger.Info("Starting metrics exporter", map[string]interface{}{
		"port": e.exportConfig.Port,
		"path": e.exportConfig.Path,
	})

	go func() {
		if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.logger.Error("Metrics exporter failed", err, map[string]interface{}{
				"port": e.exportConfig.Port,
			})
		}
	}()

	return nil
}

// Stop stops the metrics exporter server
func (e *MetricsExporter) Stop() error {
	if e.server == nil {
		return nil
	}

	e.logger.Info("Stopping metrics exporter")

	ctx, cancel := contextWithTimeout(e.exportConfig.Timeout)
	defer cancel()

	return e.server.Shutdown(ctx)
}

// WithLabel adds a custom label to all exported metrics
func (e *MetricsExporter) WithLabel(key, value string) *MetricsExporter {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.customLabels[key] = value
	return e
}

// WithLabels adds multiple custom labels to all exported metrics
func (e *MetricsExporter) WithLabels(labels map[string]string) *MetricsExporter {
	e.mu.Lock()
	defer e.mu.Unlock()
	for k, v := range labels {
		e.customLabels[k] = v
	}
	return e
}

// Handler methods

func (e *MetricsExporter) prometheusHandler(w http.ResponseWriter, r *http.Request) {
	metrics := e.metrics.GetMetrics()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Build labels string
	labels := e.buildLabelsString()

	// Export validation metrics
	fmt.Fprintf(w, "# HELP validation_total_validations Total number of validations performed\n")
	fmt.Fprintf(w, "# TYPE validation_total_validations counter\n")
	fmt.Fprintf(w, "validation_total_validations%s %d\n", labels, metrics.TotalValidations)

	fmt.Fprintf(w, "# HELP validation_rate_validations_per_second Validation rate in validations per second\n")
	fmt.Fprintf(w, "# TYPE validation_rate_validations_per_second gauge\n")
	fmt.Fprintf(w, "validation_rate_validations_per_second%s %.2f\n", labels, metrics.ValidationRate)

	fmt.Fprintf(w, "# HELP validation_average_duration_seconds Average validation duration in seconds\n")
	fmt.Fprintf(w, "# TYPE validation_average_duration_seconds gauge\n")
	fmt.Fprintf(w, "validation_average_duration_seconds%s %.6f\n", labels, metrics.AverageValidationTime.Seconds())

	fmt.Fprintf(w, "# HELP validation_total_duration_seconds Total validation duration in seconds\n")
	fmt.Fprintf(w, "# TYPE validation_total_duration_seconds counter\n")
	fmt.Fprintf(w, "validation_total_duration_seconds%s %.6f\n", labels, metrics.TotalValidationTime.Seconds())

	// Cache metrics
	fmt.Fprintf(w, "# HELP validation_cache_hits_total Total number of cache hits\n")
	fmt.Fprintf(w, "# TYPE validation_cache_hits_total counter\n")
	fmt.Fprintf(w, "validation_cache_hits_total%s %d\n", labels, metrics.CacheHits)

	fmt.Fprintf(w, "# HELP validation_cache_misses_total Total number of cache misses\n")
	fmt.Fprintf(w, "# TYPE validation_cache_misses_total counter\n")
	fmt.Fprintf(w, "validation_cache_misses_total%s %d\n", labels, metrics.CacheMisses)

	fmt.Fprintf(w, "# HELP validation_cache_hit_rate Cache hit rate (0-1)\n")
	fmt.Fprintf(w, "# TYPE validation_cache_hit_rate gauge\n")
	fmt.Fprintf(w, "validation_cache_hit_rate%s %.4f\n", labels, metrics.CacheHitRate)

	// Error metrics
	fmt.Fprintf(w, "# HELP validation_errors_total Total number of validation errors\n")
	fmt.Fprintf(w, "# TYPE validation_errors_total counter\n")
	fmt.Fprintf(w, "validation_errors_total%s %d\n", labels, metrics.TotalErrors)

	fmt.Fprintf(w, "# HELP validation_error_rate Validation error rate (0-1)\n")
	fmt.Fprintf(w, "# TYPE validation_error_rate gauge\n")
	fmt.Fprintf(w, "validation_error_rate%s %.4f\n", labels, metrics.ErrorRate)

	// Error by type
	for errorType, count := range metrics.ErrorsByType {
		typeLabels := fmt.Sprintf("%s,error_type=\"%s\"", labels, errorType)
		fmt.Fprintf(w, "# HELP validation_errors_by_type_total Total errors by type\n")
		fmt.Fprintf(w, "# TYPE validation_errors_by_type_total counter\n")
		fmt.Fprintf(w, "validation_errors_by_type_total%s %d\n", typeLabels, count)
	}

	// Layer performance metrics
	for layer, avgTime := range metrics.LayerAverages {
		layerLabels := fmt.Sprintf("%s,layer=\"%s\"", labels, layer)
		fmt.Fprintf(w, "# HELP validation_layer_average_duration_seconds Average duration per validation layer\n")
		fmt.Fprintf(w, "# TYPE validation_layer_average_duration_seconds gauge\n")
		fmt.Fprintf(w, "validation_layer_average_duration_seconds%s %.6f\n", layerLabels, avgTime.Seconds())
	}

	for layer, count := range metrics.LayerCounts {
		layerLabels := fmt.Sprintf("%s,layer=\"%s\"", labels, layer)
		fmt.Fprintf(w, "# HELP validation_layer_operations_total Total operations per validation layer\n")
		fmt.Fprintf(w, "# TYPE validation_layer_operations_total counter\n")
		fmt.Fprintf(w, "validation_layer_operations_total%s %d\n", layerLabels, count)
	}

	// Resource metrics
	fmt.Fprintf(w, "# HELP validation_concurrent_validations_current Current number of concurrent validations\n")
	fmt.Fprintf(w, "# TYPE validation_concurrent_validations_current gauge\n")
	fmt.Fprintf(w, "validation_concurrent_validations_current%s %d\n", labels, metrics.CurrentConcurrentValidations)

	fmt.Fprintf(w, "# HELP validation_concurrent_validations_peak Peak number of concurrent validations\n")
	fmt.Fprintf(w, "# TYPE validation_concurrent_validations_peak gauge\n")
	fmt.Fprintf(w, "validation_concurrent_validations_peak%s %d\n", labels, metrics.PeakConcurrentValidations)

	fmt.Fprintf(w, "# HELP validation_memory_usage_bytes Memory usage in bytes\n")
	fmt.Fprintf(w, "# TYPE validation_memory_usage_bytes gauge\n")
	fmt.Fprintf(w, "validation_memory_usage_bytes%s %d\n", labels, metrics.MemoryUsage)

	// System metrics
	fmt.Fprintf(w, "# HELP validation_uptime_seconds System uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE validation_uptime_seconds gauge\n")
	fmt.Fprintf(w, "validation_uptime_seconds%s %.0f\n", labels, metrics.Uptime.Seconds())

	// Custom metrics from logger if available
	if e.logger.metrics != nil {
		// Additional metrics can be added here
	}

	e.logger.Debug("Prometheus metrics exported", map[string]interface{}{
		"client": r.RemoteAddr,
		"path":   r.URL.Path,
	})
}

func (e *MetricsExporter) jsonHandler(w http.ResponseWriter, r *http.Request) {
	metrics := e.metrics.GetMetrics()

	// Add custom labels to response
	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"metrics":   metrics,
		"labels":    e.customLabels,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		e.logger.Error("Failed to encode JSON response", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	e.logger.Debug("JSON metrics exported", map[string]interface{}{
		"client": r.RemoteAddr,
		"path":   r.URL.Path,
	})
}

func (e *MetricsExporter) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    e.metrics.GetMetrics().Uptime.String(),
		"version":   "1.0.0",
	}

	if err := json.NewEncoder(w).Encode(healthStatus); err != nil {
		e.logger.Error("Failed to encode health response", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (e *MetricsExporter) defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>MagicORM Validation Metrics</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .endpoint { margin: 20px 0; padding: 10px; background: #f5f5f5; border-radius: 5px; }
        .endpoint h3 { margin-top: 0; }
        .endpoint code { background: #e0e0e0; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>MagicORM Validation Metrics</h1>
    <p>Validation system monitoring endpoints:</p>
    
    <div class="endpoint">
        <h3>Prometheus Metrics</h3>
        <p>Export metrics in Prometheus format:</p>
        <code>GET %s</code>
    </div>
    
    <div class="endpoint">
        <h3>JSON Metrics</h3>
        <p>Export metrics in JSON format:</p>
        <code>GET %s</code>
    </div>
    
    <div class="endpoint">
        <h3>Health Check</h3>
        <p>Check system health:</p>
        <code>GET %s</code>
    </div>
    
    <div class="endpoint">
        <h3>System Information</h3>
        <ul>
            <li>Uptime: %s</li>
            <li>Total Validations: %d</li>
            <li>Cache Hit Rate: %.2f%%</li>
            <li>Error Rate: %.2f%%</li>
        </ul>
    </div>
</body>
</html>
`,
		e.exportConfig.Path,
		e.exportConfig.MetricsPath,
		e.exportConfig.HealthCheckPath,
		e.metrics.GetMetrics().Uptime.String(),
		e.metrics.GetMetrics().TotalValidations,
		e.metrics.GetMetrics().CacheHitRate*100,
		e.metrics.GetMetrics().ErrorRate*100,
	)
}

// Helper methods

func (e *MetricsExporter) buildLabelsString() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.customLabels) == 0 {
		return ""
	}

	labels := "{"
	first := true
	for k, v := range e.customLabels {
		if !first {
			labels += ","
		}
		labels += fmt.Sprintf("%s=\"%s\"", k, v)
		first = false
	}
	labels += "}"

	return labels
}

func contextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}

// Convenience functions

// StartDefaultExporter starts a default metrics exporter
func StartDefaultExporter(metrics *MetricsCollector, logger *ValidationLogger) (*MetricsExporter, error) {
	config := DefaultExportConfig()
	exporter := NewMetricsExporter(metrics, logger, config)
	return exporter, exporter.Start()
}

// ExportMetricsToFile periodically exports metrics to a file
func ExportMetricsToFile(metrics *MetricsCollector, logger *ValidationLogger, filepath string, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		m := metrics.GetMetrics()

		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			logger.Error("Failed to marshal metrics", err)
			continue
		}

		if err := os.WriteFile(filepath, data, 0644); err != nil {
			logger.Error("Failed to write metrics file", err)
			continue
		}

		logger.Debug("Metrics exported to file", map[string]interface{}{
			"file": filepath,
		})
	}

	return nil
}
