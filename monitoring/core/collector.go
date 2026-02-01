package core

import (
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	// CounterMetric represents a cumulative metric that only increases
	CounterMetric MetricType = "counter"
	// GaugeMetric represents a metric that can go up and down
	GaugeMetric MetricType = "gauge"
	// HistogramMetric represents a metric that samples observations
	HistogramMetric MetricType = "histogram"
	// SummaryMetric represents a metric that calculates quantiles
	SummaryMetric MetricType = "summary"
)

// Metric represents a single metric with labels and value
type Metric struct {
	Name        string            `json:"name"`
	Type        MetricType        `json:"type"`
	Value       float64           `json:"value"`
	Labels      map[string]string `json:"labels"`
	Timestamp   time.Time         `json:"timestamp"`
	Description string            `json:"description,omitempty"`
}

// MetricDefinition defines a metric's structure and behavior
type MetricDefinition struct {
	Name       string              `json:"name"`
	Type       MetricType          `json:"type"`
	Help       string              `json:"help"`
	LabelNames []string            `json:"label_names"`
	Buckets    []float64           `json:"buckets,omitempty"`    // For histograms
	Objectives map[float64]float64 `json:"objectives,omitempty"` // For summaries
	MaxAge     time.Duration       `json:"max_age,omitempty"`
}

// Collector collects and manages metrics
type Collector struct {
	mu sync.RWMutex

	config      *MonitoringConfig
	metrics     map[string][]Metric
	definitions map[string]MetricDefinition

	// Performance optimization
	batchBuffer []Metric
	batchSize   int
	batchMutex  sync.Mutex

	// Statistics
	stats CollectorStats
}

// CollectorStats holds collector statistics
type CollectorStats struct {
	MetricsCollected int64         `json:"metrics_collected"`
	MetricsDropped   int64         `json:"metrics_dropped"`
	BatchOperations  int64         `json:"batch_operations"`
	LastCollection   time.Time     `json:"last_collection"`
	Uptime           time.Duration `json:"uptime"`
	StartTime        time.Time     `json:"start_time"`
}

// NewCollector creates a new metric collector
func NewCollector(config *MonitoringConfig) *Collector {
	if config == nil {
		defaultConfig := DefaultMonitoringConfig()
		config = &defaultConfig
	}

	collector := &Collector{
		config:      config,
		metrics:     make(map[string][]Metric),
		definitions: make(map[string]MetricDefinition),
		batchBuffer: make([]Metric, 0, config.BufferSize),
		batchSize:   config.BatchSize,
		stats: CollectorStats{
			StartTime: time.Now(),
		},
	}

	// Register default metric definitions
	collector.registerDefaultDefinitions()

	// Start background tasks if async collection is enabled
	if config.AsyncCollection {
		go collector.startBackgroundTasks()
	}

	return collector
}

// RegisterDefinition registers a new metric definition
func (c *Collector) RegisterDefinition(def MetricDefinition) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.definitions[def.Name]; exists {
		return &MetricError{Name: def.Name, Message: "metric already defined"}
	}

	// Validate definition
	if err := c.validateDefinition(def); err != nil {
		return err
	}

	c.definitions[def.Name] = def
	return nil
}

// Record records a metric value
func (c *Collector) Record(name string, value float64, labels map[string]string) error {
	if !c.config.ShouldSample() {
		c.stats.MetricsDropped++
		return nil
	}

	def, err := c.getDefinition(name)
	if err != nil {
		return err
	}

	// Validate labels
	if err := c.validateLabels(def, labels); err != nil {
		return err
	}

	metric := Metric{
		Name:      name,
		Type:      def.Type,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	if c.config.AsyncCollection {
		return c.recordAsync(metric)
	}

	return c.recordSync(metric)
}

// RecordWithTimestamp records a metric with a specific timestamp
func (c *Collector) RecordWithTimestamp(name string, value float64, labels map[string]string, timestamp time.Time) error {
	if !c.config.ShouldSample() {
		c.stats.MetricsDropped++
		return nil
	}

	def, err := c.getDefinition(name)
	if err != nil {
		return err
	}

	// Validate labels
	if err := c.validateLabels(def, labels); err != nil {
		return err
	}

	metric := Metric{
		Name:      name,
		Type:      def.Type,
		Value:     value,
		Labels:    labels,
		Timestamp: timestamp,
	}

	if c.config.AsyncCollection {
		return c.recordAsync(metric)
	}

	return c.recordSync(metric)
}

// Increment increments a counter metric
func (c *Collector) Increment(name string, labels map[string]string) error {
	return c.Record(name, 1, labels)
}

// Decrement decrements a gauge metric
func (c *Collector) Decrement(name string, labels map[string]string) error {
	return c.Record(name, -1, labels)
}

// Observe observes a value for a histogram or summary metric
func (c *Collector) Observe(name string, value float64, labels map[string]string) error {
	return c.Record(name, value, labels)
}

// GetMetrics returns all collected metrics
func (c *Collector) GetMetrics() map[string][]Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid concurrent modification
	result := make(map[string][]Metric)
	for name, metrics := range c.metrics {
		metricsCopy := make([]Metric, len(metrics))
		copy(metricsCopy, metrics)
		result[name] = metricsCopy
	}

	return result
}

// GetMetric returns metrics for a specific name
func (c *Collector) GetMetric(name string) ([]Metric, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics, exists := c.metrics[name]
	if !exists {
		return nil, &MetricError{Name: name, Message: "metric not found"}
	}

	// Return a copy
	result := make([]Metric, len(metrics))
	copy(result, metrics)
	return result, nil
}

// GetStats returns collector statistics
func (c *Collector) GetStats() CollectorStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Uptime = time.Since(stats.StartTime)
	return stats
}

// Reset clears all collected metrics
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = make(map[string][]Metric)
	c.batchBuffer = make([]Metric, 0, c.config.BufferSize)
	c.stats.MetricsCollected = 0
	c.stats.MetricsDropped = 0
	c.stats.BatchOperations = 0
	c.stats.LastCollection = time.Now()
}

// Cleanup removes old metrics based on retention period
func (c *Collector) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-c.config.RetentionPeriod)

	for name, metrics := range c.metrics {
		var filtered []Metric
		for _, metric := range metrics {
			if metric.Timestamp.After(cutoff) {
				filtered = append(filtered, metric)
			}
		}
		c.metrics[name] = filtered
	}
}

// Private methods

func (c *Collector) recordSync(metric Metric) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics[metric.Name] = append(c.metrics[metric.Name], metric)
	c.stats.MetricsCollected++
	c.stats.LastCollection = time.Now()

	return nil
}

func (c *Collector) recordAsync(metric Metric) error {
	c.batchMutex.Lock()
	defer c.batchMutex.Unlock()

	c.batchBuffer = append(c.batchBuffer, metric)

	// Flush if buffer is full
	if len(c.batchBuffer) >= c.batchSize {
		return c.flushBatch()
	}

	return nil
}

func (c *Collector) flushBatch() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.batchBuffer) == 0 {
		return nil
	}

	// Group metrics by name for efficient storage
	for _, metric := range c.batchBuffer {
		c.metrics[metric.Name] = append(c.metrics[metric.Name], metric)
	}

	c.stats.MetricsCollected += int64(len(c.batchBuffer))
	c.stats.BatchOperations++
	c.stats.LastCollection = time.Now()

	// Clear buffer
	c.batchBuffer = c.batchBuffer[:0]

	return nil
}

func (c *Collector) startBackgroundTasks() {
	// Cleanup task
	cleanupTicker := time.NewTicker(c.config.RetentionPeriod / 2)
	defer cleanupTicker.Stop()

	// Batch flush task
	flushTicker := time.NewTicker(c.config.CollectionInterval)
	defer flushTicker.Stop()

	for {
		select {
		case <-cleanupTicker.C:
			c.Cleanup()
		case <-flushTicker.C:
			c.batchMutex.Lock()
			if len(c.batchBuffer) > 0 {
				c.flushBatch()
			}
			c.batchMutex.Unlock()
		}
	}
}

func (c *Collector) getDefinition(name string) (MetricDefinition, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	def, exists := c.definitions[name]
	if !exists {
		return MetricDefinition{}, &MetricError{Name: name, Message: "metric not defined"}
	}

	return def, nil
}

func (c *Collector) validateDefinition(def MetricDefinition) error {
	if def.Name == "" {
		return &MetricError{Name: def.Name, Message: "name cannot be empty"}
	}

	if def.Help == "" {
		return &MetricError{Name: def.Name, Message: "help text cannot be empty"}
	}

	switch def.Type {
	case CounterMetric, GaugeMetric, HistogramMetric, SummaryMetric:
		// Valid types
	default:
		return &MetricError{Name: def.Name, Message: "invalid metric type"}
	}

	if def.Type == HistogramMetric && len(def.Buckets) == 0 {
		// Set default buckets if none provided
		def.Buckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	}

	if def.Type == SummaryMetric && len(def.Objectives) == 0 {
		// Set default objectives if none provided
		def.Objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}

	return nil
}

func (c *Collector) validateLabels(def MetricDefinition, labels map[string]string) error {
	// Check for required labels
	for _, labelName := range def.LabelNames {
		if _, exists := labels[labelName]; !exists {
			return &MetricError{
				Name:    def.Name,
				Message: "missing required label: " + labelName,
			}
		}
	}

	// Check for extra labels (optional - could be allowed)
	// For now, we allow extra labels

	return nil
}

func (c *Collector) registerDefaultDefinitions() {
	// System metrics
	c.RegisterDefinition(MetricDefinition{
		Name:       "monitoring_metrics_collected_total",
		Type:       CounterMetric,
		Help:       "Total number of metrics collected",
		LabelNames: []string{},
	})

	c.RegisterDefinition(MetricDefinition{
		Name:       "monitoring_metrics_dropped_total",
		Type:       CounterMetric,
		Help:       "Total number of metrics dropped due to sampling",
		LabelNames: []string{},
	})

	c.RegisterDefinition(MetricDefinition{
		Name:       "monitoring_collector_uptime_seconds",
		Type:       GaugeMetric,
		Help:       "Collector uptime in seconds",
		LabelNames: []string{},
	})

	// Performance metrics
	c.RegisterDefinition(MetricDefinition{
		Name:       "monitoring_batch_operations_total",
		Type:       CounterMetric,
		Help:       "Total number of batch operations performed",
		LabelNames: []string{},
	})

	c.RegisterDefinition(MetricDefinition{
		Name:       "monitoring_buffer_size",
		Type:       GaugeMetric,
		Help:       "Current buffer size",
		LabelNames: []string{},
	})
}

// MetricError represents a metric-related error
type MetricError struct {
	Name    string
	Message string
}

func (e *MetricError) Error() string {
	return "metric error: " + e.Name + ": " + e.Message
}

// Convenience functions for common metric patterns

// RecordDuration records a duration metric
func (c *Collector) RecordDuration(name string, duration time.Duration, labels map[string]string) error {
	return c.Record(name, duration.Seconds(), labels)
}

// RecordOperation records a complete operation with timing
func (c *Collector) RecordOperation(name string, startTime time.Time, success bool, labels map[string]string) error {
	// Record duration
	duration := time.Since(startTime)
	if err := c.RecordDuration(name+"_duration_seconds", duration, labels); err != nil {
		return err
	}

	// Record operation count
	status := "success"
	if !success {
		status = "error"
	}

	opLabels := make(map[string]string)
	for k, v := range labels {
		opLabels[k] = v
	}
	opLabels["status"] = status

	return c.Increment(name+"_total", opLabels)
}

// RecordError records an error occurrence
func (c *Collector) RecordError(name string, errorType string, labels map[string]string) error {
	errLabels := make(map[string]string)
	for k, v := range labels {
		errLabels[k] = v
	}
	errLabels["error_type"] = errorType

	return c.Increment(name+"_errors_total", errLabels)
}
