package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Exporter exports metrics via HTTP in various formats
type Exporter struct {
	mu sync.RWMutex

	collector *Collector
	config    *ExportConfig
	server    *http.Server

	// Custom labels applied to all metrics
	customLabels map[string]string

	// Statistics
	stats ExporterStats

	// Cache for formatted metrics
	cache struct {
		prometheus string
		json       string
		timestamp  time.Time
		mu         sync.RWMutex
	}
}

// ExporterStats holds exporter statistics
type ExporterStats struct {
	RequestsTotal  int64            `json:"requests_total"`
	RequestsByPath map[string]int64 `json:"requests_by_path"`
	ErrorsTotal    int64            `json:"errors_total"`
	LastRequest    time.Time        `json:"last_request"`
	StartTime      time.Time        `json:"start_time"`
	Uptime         time.Duration    `json:"uptime"`
	CacheHits      int64            `json:"cache_hits"`
	CacheMisses    int64            `json:"cache_misses"`
}

// NewExporter creates a new metrics exporter
func NewExporter(collector *Collector, config *ExportConfig) *Exporter {
	if config == nil {
		defaultConfig := DefaultExportConfig()
		config = &defaultConfig
	}

	exporter := &Exporter{
		collector:    collector,
		config:       config,
		customLabels: make(map[string]string),
		stats: ExporterStats{
			StartTime:      time.Now(),
			RequestsByPath: make(map[string]int64),
		},
	}

	// Initialize cache
	exporter.cache.timestamp = time.Time{}

	return exporter
}

// Start starts the HTTP server for metric export
func (e *Exporter) Start() error {
	if !e.config.Enabled {
		return nil
	}

	mux := http.NewServeMux()

	// Register handlers
	if e.config.EnablePrometheus {
		mux.HandleFunc(e.config.Path, e.prometheusHandler)
	}

	if e.config.EnableJSON {
		mux.HandleFunc(e.config.MetricsPath, e.jsonHandler)
	}

	mux.HandleFunc(e.config.HealthCheckPath, e.healthHandler)
	mux.HandleFunc(e.config.InfoPath, e.infoHandler)

	// Add middleware for authentication and statistics
	handler := e.withMiddleware(mux)

	e.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", e.config.Port),
		Handler:      handler,
		ReadTimeout:  e.config.ScrapeTimeout,
		WriteTimeout: e.config.ScrapeTimeout,
	}

	// Start server in background
	go func() {
		var err error
		if e.config.EnableTLS {
			err = e.server.ListenAndServeTLS(e.config.TLSCertPath, e.config.TLSKeyPath)
		} else {
			err = e.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			// Log error (in production, use proper logging)
			fmt.Printf("Exporter server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the HTTP server
func (e *Exporter) Stop() error {
	if e.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return e.server.Shutdown(ctx)
}

// WithLabel adds a custom label to all exported metrics
func (e *Exporter) WithLabel(key, value string) *Exporter {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.customLabels[key] = value
	e.invalidateCache()
	return e
}

// WithLabels adds multiple custom labels to all exported metrics
func (e *Exporter) WithLabels(labels map[string]string) *Exporter {
	e.mu.Lock()
	defer e.mu.Unlock()
	for k, v := range labels {
		e.customLabels[k] = v
	}
	e.invalidateCache()
	return e
}

// GetStats returns exporter statistics
func (e *Exporter) GetStats() ExporterStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stats := e.stats
	stats.Uptime = time.Since(stats.StartTime)
	return stats
}

// ResetStats resets exporter statistics
func (e *Exporter) ResetStats() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stats = ExporterStats{
		StartTime:      time.Now(),
		RequestsByPath: make(map[string]int64),
	}
}

// Handler methods

func (e *Exporter) prometheusHandler(w http.ResponseWriter, r *http.Request) {
	e.recordRequest(r.URL.Path)

	// Check cache first
	e.cache.mu.RLock()
	cacheValid := !e.cache.timestamp.IsZero() && time.Since(e.cache.timestamp) < e.config.RefreshInterval
	if cacheValid && e.cache.prometheus != "" {
		e.cache.mu.RUnlock()
		e.stats.CacheHits++
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Write([]byte(e.cache.prometheus))
		return
	}
	e.cache.mu.RUnlock()

	e.stats.CacheMisses++

	// Generate metrics
	metrics := e.collector.GetMetrics()
	prometheusMetrics := e.formatPrometheus(metrics)

	// Update cache
	e.cache.mu.Lock()
	e.cache.prometheus = prometheusMetrics
	e.cache.timestamp = time.Now()
	e.cache.mu.Unlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write([]byte(prometheusMetrics))
}

func (e *Exporter) jsonHandler(w http.ResponseWriter, r *http.Request) {
	e.recordRequest(r.URL.Path)

	// Check cache first
	e.cache.mu.RLock()
	cacheValid := !e.cache.timestamp.IsZero() && time.Since(e.cache.timestamp) < e.config.RefreshInterval
	if cacheValid && e.cache.json != "" {
		e.cache.mu.RUnlock()
		e.stats.CacheHits++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(e.cache.json))
		return
	}
	e.cache.mu.RUnlock()

	e.stats.CacheMisses++

	// Generate metrics
	metrics := e.collector.GetMetrics()
	stats := e.collector.GetStats()
	exporterStats := e.GetStats()

	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"metrics":   metrics,
		"stats": map[string]interface{}{
			"collector": stats,
			"exporter":  exporterStats,
		},
		"labels": e.getCustomLabels(),
	}

	// Add system info
	response["system"] = map[string]interface{}{
		"go_version":    "1.24+", // This would be dynamic in real implementation
		"export_config": e.config,
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		e.recordError()
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	// Update cache
	e.cache.mu.Lock()
	e.cache.json = string(data)
	e.cache.timestamp = time.Now()
	e.cache.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (e *Exporter) healthHandler(w http.ResponseWriter, r *http.Request) {
	e.recordRequest(r.URL.Path)

	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(e.stats.StartTime).String(),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(healthStatus); err != nil {
		e.recordError()
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (e *Exporter) infoHandler(w http.ResponseWriter, r *http.Request) {
	e.recordRequest(r.URL.Path)

	info := map[string]interface{}{
		"name":        "MagicORM Monitoring Exporter",
		"version":     "1.0.0",
		"description": "Unified monitoring system for MagicORM",
		"endpoints": map[string]string{
			"prometheus": e.config.Path,
			"json":       e.config.MetricsPath,
			"health":     e.config.HealthCheckPath,
		},
		"config": map[string]interface{}{
			"port":             e.config.Port,
			"tls_enabled":      e.config.EnableTLS,
			"auth_enabled":     e.config.EnableAuth,
			"refresh_interval": e.config.RefreshInterval.String(),
		},
		"custom_labels": e.getCustomLabels(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		e.recordError()
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Private methods

func (e *Exporter) withMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authentication
		if e.config.EnableAuth && !e.authenticateRequest(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Host filtering
		if len(e.config.AllowedHosts) > 0 && !e.isHostAllowed(r) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func (e *Exporter) authenticateRequest(r *http.Request) bool {
	if !e.config.EnableAuth {
		return true
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	// Simple token-based authentication
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token == e.config.AuthToken
}

func (e *Exporter) isHostAllowed(r *http.Request) bool {
	host := r.Host
	if host == "" {
		host = r.RemoteAddr
	}

	for _, allowedHost := range e.config.AllowedHosts {
		if host == allowedHost || strings.HasPrefix(host, allowedHost+":") {
			return true
		}
	}

	return false
}

func (e *Exporter) formatPrometheus(metrics map[string][]Metric) string {
	var builder strings.Builder

	// Get custom labels string
	customLabelsStr := e.formatCustomLabels()

	// Sort metric names for consistent output
	metricNames := make([]string, 0, len(metrics))
	for name := range metrics {
		metricNames = append(metricNames, name)
	}
	sort.Strings(metricNames)

	for _, metricName := range metricNames {
		metricList := metrics[metricName]
		if len(metricList) == 0 {
			continue
		}

		// Get definition for help text
		def, err := e.collector.getDefinition(metricName)
		if err != nil {
			// Use default help text if definition not found
			def.Help = "Metric " + metricName
		}

		// Write help and type lines
		builder.WriteString("# HELP ")
		builder.WriteString(metricName)
		builder.WriteString(" ")
		builder.WriteString(def.Help)
		builder.WriteString("\n")

		builder.WriteString("# TYPE ")
		builder.WriteString(metricName)
		builder.WriteString(" ")
		builder.WriteString(string(def.Type))
		builder.WriteString("\n")

		// Write each metric value
		for _, metric := range metricList {
			builder.WriteString(metricName)

			// Combine custom labels and metric labels
			labels := e.combineLabels(metric.Labels)
			if len(labels) > 0 {
				builder.WriteString("{")
				first := true

				// Sort label keys for consistent output
				labelKeys := make([]string, 0, len(labels))
				for k := range labels {
					labelKeys = append(labelKeys, k)
				}
				sort.Strings(labelKeys)

				for _, key := range labelKeys {
					if !first {
						builder.WriteString(",")
					}
					builder.WriteString(key)
					builder.WriteString("=\"")
					builder.WriteString(labels[key])
					builder.WriteString("\"")
					first = false
				}
				builder.WriteString("}")
			}

			builder.WriteString(" ")
			builder.WriteString(fmt.Sprintf("%v", metric.Value))

			// Add timestamp if supported
			if !metric.Timestamp.IsZero() {
				builder.WriteString(" ")
				builder.WriteString(fmt.Sprintf("%d", metric.Timestamp.UnixMilli()))
			}

			builder.WriteString("\n")
		}
	}

	// Add exporter metrics
	builder.WriteString(e.formatExporterMetrics(customLabelsStr))

	return builder.String()
}

func (e *Exporter) formatCustomLabels() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.customLabels) == 0 {
		return ""
	}

	var labels []string
	for k, v := range e.customLabels {
		labels = append(labels, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	sort.Strings(labels)

	return "{" + strings.Join(labels, ",") + "}"
}

func (e *Exporter) combineLabels(metricLabels map[string]string) map[string]string {
	e.mu.RLock()
	customLabels := e.customLabels
	e.mu.RUnlock()

	combined := make(map[string]string)

	// Add custom labels first
	for k, v := range customLabels {
		combined[k] = v
	}

	// Add metric labels (override custom labels if same key)
	for k, v := range metricLabels {
		combined[k] = v
	}

	return combined
}

func (e *Exporter) formatExporterMetrics(customLabels string) string {
	var builder strings.Builder

	stats := e.GetStats()

	// Exporter statistics
	builder.WriteString("# HELP monitoring_exporter_requests_total Total requests to exporter\n")
	builder.WriteString("# TYPE monitoring_exporter_requests_total counter\n")
	builder.WriteString(fmt.Sprintf("monitoring_exporter_requests_total%s %d\n", customLabels, stats.RequestsTotal))

	builder.WriteString("# HELP monitoring_exporter_errors_total Total exporter errors\n")
	builder.WriteString("# TYPE monitoring_exporter_errors_total counter\n")
	builder.WriteString(fmt.Sprintf("monitoring_exporter_errors_total%s %d\n", customLabels, stats.ErrorsTotal))

	builder.WriteString("# HELP monitoring_exporter_cache_hits_total Total cache hits\n")
	builder.WriteString("# TYPE monitoring_exporter_cache_hits_total counter\n")
	builder.WriteString(fmt.Sprintf("monitoring_exporter_cache_hits_total%s %d\n", customLabels, stats.CacheHits))

	builder.WriteString("# HELP monitoring_exporter_cache_misses_total Total cache misses\n")
	builder.WriteString("# TYPE monitoring_exporter_cache_misses_total counter\n")
	builder.WriteString(fmt.Sprintf("monitoring_exporter_cache_misses_total%s %d\n", customLabels, stats.CacheMisses))

	builder.WriteString("# HELP monitoring_exporter_uptime_seconds Exporter uptime in seconds\n")
	builder.WriteString("# TYPE monitoring_exporter_uptime_seconds gauge\n")
	builder.WriteString(fmt.Sprintf("monitoring_exporter_uptime_seconds%s %.0f\n", customLabels, stats.Uptime.Seconds()))

	// Requests by path
	for path, count := range stats.RequestsByPath {
		pathLabels := customLabels
		if pathLabels == "" {
			pathLabels = "{path=\"" + path + "\"}"
		} else {
			pathLabels = strings.TrimSuffix(pathLabels, "}")
			pathLabels += ",path=\"" + path + "\"}"
		}

		builder.WriteString("# HELP monitoring_exporter_requests_by_path_total Requests by path\n")
		builder.WriteString("# TYPE monitoring_exporter_requests_by_path_total counter\n")
		builder.WriteString(fmt.Sprintf("monitoring_exporter_requests_by_path_total%s %d\n", pathLabels, count))
	}

	return builder.String()
}

func (e *Exporter) recordRequest(path string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stats.RequestsTotal++
	e.stats.RequestsByPath[path]++
	e.stats.LastRequest = time.Now()
}

func (e *Exporter) recordError() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stats.ErrorsTotal++
}

func (e *Exporter) invalidateCache() {
	e.cache.mu.Lock()
	defer e.cache.mu.Unlock()
	e.cache.timestamp = time.Time{}
	e.cache.prometheus = ""
	e.cache.json = ""
}

func (e *Exporter) getCustomLabels() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	labels := make(map[string]string)
	for k, v := range e.customLabels {
		labels[k] = v
	}
	return labels
}

// Convenience functions

// StartDefaultExporter starts an exporter with default configuration
func StartDefaultExporter(collector *Collector) (*Exporter, error) {
	config := DefaultExportConfig()
	exporter := NewExporter(collector, &config)
	return exporter, exporter.Start()
}

// StartExporterWithConfig starts an exporter with custom configuration
func StartExporterWithConfig(collector *Collector, config *ExportConfig) (*Exporter, error) {
	exporter := NewExporter(collector, config)
	return exporter, exporter.Start()
}
