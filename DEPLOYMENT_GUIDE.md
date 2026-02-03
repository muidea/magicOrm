# MagicORM 监控系统部署指南

本文档提供 MagicORM 监控系统的部署指南和集成示例。

## 目录

1. [系统要求](#系统要求)
2. [安装和配置](#安装和配置)
3. [集成示例](#集成示例)
4. [生产环境部署](#生产环境部署)
5. [性能调优](#性能调优)
6. [故障排除](#故障排除)

## 系统要求

### 软件要求
- Go 1.24+
- PostgreSQL 12+ 或 MySQL 8.0+
- 外部监控系统（可选）：Prometheus, Grafana, Datadog 等

### 硬件要求
- 内存：至少 512MB RAM（监控系统本身）
- 存储：根据监控数据保留策略确定
- CPU：现代多核处理器

## 安装和配置

### 1. 安装 MagicORM

```bash
# 使用 Go modules
go get github.com/muidea/magicOrm

# 或克隆仓库
git clone https://github.com/muidea/magicOrm.git
cd magicOrm
go mod download
```

### 2. 基本配置

#### 数据库配置
```go
// config/database.go
package config

import "github.com/muidea/magicOrm/orm"

var DatabaseConfig = &orm.Options{
    Driver: "postgresql", // 或 "mysql"
    DSN:    "host=localhost port=5432 user=postgres password=secret dbname=mydb sslmode=disable",
    MaxOpenConns: 25,
    MaxIdleConns: 5,
    ConnMaxLifetime: time.Hour,
}
```

#### 监控配置
```go
// config/monitoring.go
package config

import "time"

type MonitoringConfig struct {
    // 启用/禁用监控
    Enabled bool
    
    // 采样率 (0.0 - 1.0)
    SamplingRate float64
    
    // 默认标签
    DefaultLabels map[string]string
    
    // 异步收集配置
    AsyncCollection bool
    CollectionInterval time.Duration
    
    // 导出配置
    ExportHandler func(metrics []monitoring.Metric)
}

var DefaultMonitoringConfig = MonitoringConfig{
    Enabled:           true,
    SamplingRate:      1.0,
    DefaultLabels: map[string]string{
        "environment": "development",
        "service":     "magicorm-service",
    },
    AsyncCollection:   true,
    CollectionInterval: 30 * time.Second,
}
```

## 集成示例

### 示例 1：基本集成

```go
// main.go - 基本监控集成
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/muidea/magicOrm/monitoring"
    "github.com/muidea/magicOrm/orm"
    "github.com/muidea/magicOrm/provider"
    
    "your-project/config"
)

func main() {
    // 初始化监控
    collector := monitoring.NewCollector()
    
    // 设置默认标签
    collector.WithDefaultLabels(config.DefaultMonitoringConfig.DefaultLabels)
    
    // 设置导出处理器
    if config.DefaultMonitoringConfig.ExportHandler != nil {
        collector.SetExportHandler(config.DefaultMonitoringConfig.ExportHandler)
    }
    
    // 初始化ORM
    orm.Initialize()
    defer orm.Uninitialized()
    
    // 创建Provider
    localProvider := provider.NewLocalProvider("default", nil)
    
    // 创建ORM实例
    o, err := orm.NewOrm(localProvider, config.DatabaseConfig, "")
    if err != nil {
        log.Fatal(err)
    }
    defer o.Release()
    
    // 包装为带监控的ORM
    monitoredOrm := monitoring.NewMonitoredOrm(o, collector)
    
    // 使用带监控的ORM
    runApplication(monitoredOrm, collector)
}

func runApplication(o *orm.Orm, collector monitoring.Collector) {
    // 应用逻辑
    // ...
}
```

### 示例 2：完整应用集成

```go
// app/monitored_app.go - 完整监控应用
package app

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"
    
    "github.com/muidea/magicOrm/monitoring"
    "github.com/muidea/magicOrm/orm"
    "github.com/muidea/magicOrm/validation"
)

type MonitoredApplication struct {
    orm        *orm.Orm
    collector  monitoring.Collector
    validator  *validation.ValidationManager
    httpServer *http.Server
}

func NewMonitoredApplication(
    o *orm.Orm,
    collector monitoring.Collector,
    validator *validation.ValidationManager,
) *MonitoredApplication {
    app := &MonitoredApplication{
        orm:       o,
        collector: collector,
        validator: validator,
    }
    
    // 设置HTTP监控端点
    app.setupHTTPMonitoring()
    
    return app
}

func (app *MonitoredApplication) setupHTTPMonitoring() {
    mux := http.NewServeMux()
    
    // 健康检查端点
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
            "time":   time.Now().Format(time.RFC3339),
        })
    })
    
    // 监控数据端点
    mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        metrics := app.collector.GetMetrics()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(metrics)
    })
    
    // Prometheus格式端点
    mux.HandleFunc("/metrics/prometheus", func(w http.ResponseWriter, r *http.Request) {
        prometheusData := app.collector.ToPrometheusFormat()
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte(prometheusData))
    })
    
    app.httpServer = &http.Server{
        Addr:    ":8081",
        Handler: mux,
    }
}

func (app *MonitoredApplication) Start() error {
    // 启动HTTP服务器
    go func() {
        log.Printf("监控服务器启动在 :8081")
        if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("监控服务器错误: %v", err)
        }
    }()
    
    return nil
}

func (app *MonitoredApplication) Stop() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return app.httpServer.Shutdown(ctx)
}
```

### 示例 3：与外部监控系统集成

```go
// monitoring/external_integration.go - 外部监控系统集成
package monitoring

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"
)

// Prometheus集成
type PrometheusExporter struct {
    endpoint string
    client   *http.Client
}

func NewPrometheusExporter(endpoint string) *PrometheusExporter {
    return &PrometheusExporter{
        endpoint: endpoint,
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (e *PrometheusExporter) Export(metrics []Metric) error {
    prometheusData := ToPrometheusFormat(metrics)
    
    req, err := http.NewRequest("POST", e.endpoint, bytes.NewBufferString(prometheusData))
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "text/plain")
    
    resp, err := e.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("导出失败: %s", resp.Status)
    }
    
    return nil
}

// Datadog集成
type DatadogExporter struct {
    apiKey  string
    appKey  string
    site    string
    client  *http.Client
}

func NewDatadogExporter(apiKey, appKey, site string) *DatadogExporter {
    return &DatadogExporter{
        apiKey: apiKey,
        appKey: appKey,
        site:   site,
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (e *DatadogExporter) Export(metrics []Metric) error {
    datadogMetrics := make([]map[string]interface{}, len(metrics))
    
    for i, metric := range metrics {
        datadogMetrics[i] = map[string]interface{}{
            "metric": metric.Name,
            "points": [][]interface{}{
                {float64(time.Now().Unix()), metric.Value},
            },
            "tags": metric.Labels,
            "type": "gauge",
        }
    }
    
    payload := map[string]interface{}{
        "series": datadogMetrics,
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    url := fmt.Sprintf("https://api.%s/api/v1/series", e.site)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("DD-API-KEY", e.apiKey)
    req.Header.Set("DD-APPLICATION-KEY", e.appKey)
    
    resp, err := e.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusAccepted {
        return fmt.Errorf("Datadog导出失败: %s", resp.Status)
    }
    
    return nil
}

// 使用示例
func setupExternalMonitoring() monitoring.Collector {
    collector := NewCollector()
    
    // 设置Prometheus导出
    prometheusExporter := NewPrometheusExporter("http://localhost:9090/api/v1/write")
    collector.SetExportHandler(func(metrics []Metric) {
        if err := prometheusExporter.Export(metrics); err != nil {
            log.Printf("Prometheus导出错误: %v", err)
        }
    })
    
    // 或设置Datadog导出
    datadogExporter := NewDatadogExporter(
        os.Getenv("DATADOG_API_KEY"),
        os.Getenv("DATADOG_APP_KEY"),
        "datadoghq.com",
    )
    
    collector.SetExportHandler(func(metrics []Metric) {
        if err := datadogExporter.Export(metrics); err != nil {
            log.Printf("Datadog导出错误: %v", err)
        }
    })
    
    return collector
}
```

## 生产环境部署

### 1. Docker 部署

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o magicorm-app ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/magicorm-app .
COPY --from=builder /app/config ./config

EXPOSE 8080 8081

CMD ["./magicorm-app"]
```

### 2. Kubernetes 部署

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: magicorm-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: magicorm
  template:
    metadata:
      labels:
        app: magicorm
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8081"
        prometheus.io/path: "/metrics/prometheus"
    spec:
      containers:
      - name: magicorm
        image: your-registry/magicorm-app:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8081
          name: metrics
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: url
        - name: MONITORING_ENABLED
          value: "true"
        - name: MONITORING_SAMPLING_RATE
          value: "0.1"
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 3. 环境变量配置

```bash
# .env.production
DATABASE_URL=postgresql://user:password@host:5432/dbname
MONITORING_ENABLED=true
MONITORING_SAMPLING_RATE=0.1
MONITORING_ASYNC_COLLECTION=true
MONITORING_COLLECTION_INTERVAL=30s

# 外部监控系统配置
PROMETHEUS_ENDPOINT=http://prometheus:9090/api/v1/write
DATADOG_API_KEY=your-datadog-api-key
DATADOG_APP_KEY=your-datadog-app-key
```

## 性能调优

### 1. 监控系统性能优化

```go
// config/performance.go
package config

import "time"

type PerformanceConfig struct {
    // 监控采样率
    MonitoringSamplingRate float64
    
    // 异步收集配置
    AsyncCollection bool
    CollectionInterval time.Duration
    CollectionBatchSize int
    
    // 内存限制
    MaxMetricsInMemory int
    MetricRetentionPeriod time.Duration
    
    // 缓存配置
    EnableMetricCaching bool
    MetricCacheTTL time.Duration
    MaxCacheSize int
}

var ProductionPerformanceConfig = PerformanceConfig{
    MonitoringSamplingRate: 0.1, // 10%采样
    AsyncCollection:        true,
    CollectionInterval:     30 * time.Second,
    CollectionBatchSize:    1000,
    MaxMetricsInMemory:     10000,
    MetricRetentionPeriod:  1 * time.Hour,
    EnableMetricCaching:    true,
    MetricCacheTTL:         5 * time.Minute,
    MaxCacheSize:          1000,
}

var HighLoadPerformanceConfig = PerformanceConfig{
    MonitoringSamplingRate: 0.01, // 1%采样
    AsyncCollection:        true,
    CollectionInterval:     60 * time.Second,
    CollectionBatchSize:    5000,
    MaxMetricsInMemory:     5000,
    MetricRetentionPeriod:  30 * time.Minute,
    EnableMetricCaching:    false, // 高负载时禁用缓存
}
```

### 2. 数据库监控优化

```go
// monitoring/database_optimized.go
package monitoring

import (
    "sync/atomic"
    "time"
)

type OptimizedDatabaseCollector struct {
    collector Collector
    samplingRate float64
    
    // 性能计数器
    operationsProcessed uint64
    operationsSkipped   uint64
    totalLatency        time.Duration
}

func NewOptimizedDatabaseCollector(collector Collector, samplingRate float64) *OptimizedDatabaseCollector {
    return &OptimizedDatabaseCollector{
        collector:    collector,
        samplingRate: samplingRate,
    }
}

func (c *OptimizedDatabaseCollector) RecordDatabaseOperation(
    dbType string,
    queryType QueryType,
    success bool,
    latency time.Duration,
    rowsAffected int,
    errInfo *ErrorInfo,
    labels map[string]string,
) {
    // 应用采样率
    if !shouldSample(c.samplingRate) {
        atomic.AddUint64(&c.operationsSkipped, 1)
        return
    }
    
    atomic.AddUint64(&c.operationsProcessed, 1)
    atomic.AddUint64(&c.totalLatency, uint64(latency))
    
    c.collector.RecordDatabaseOperation(
        dbType,
        queryType,
        success,
        latency,
        rowsAffected,
        errInfo,
        labels,
    )
}

func shouldSample(samplingRate float64) bool {
    // 实现采样逻辑
    return true // 简化示例
}

// 获取性能统计
func (c *OptimizedDatabaseCollector) GetStats() map[string]interface{} {
    processed := atomic.LoadUint64(&c.operationsProcessed)
    skipped := atomic.LoadUint64(&c.operationsSkipped)
    totalLatency := time.Duration(atomic.LoadUint64(&c.totalLatency))
    
    avgLatency := time.Duration(0)
    if processed > 0 {
        avgLatency = totalLatency / time.Duration(processed)
    }
    
    return map[string]interface{}{
        "operations_processed": processed,
        "operations_skipped":   skipped,
        "sampling_rate":        c.samplingRate,
        "avg_latency":          avgLatency.String(),
        "total_latency":        totalLatency.String(),
    }
}
```

## 故障排除

### 常见问题

#### 1. 内存使用过高

**症状**: 应用内存使用持续增长

**解决方案**:
```go
// 减少监控数据保留时间
config.MetricRetentionPeriod = 30 * time.Minute

// 降低采样率
config.MonitoringSamplingRate = 0.01

// 启用内存限制
config.MaxMetricsInMemory = 5000
```

#### 2. 性能影响明显

**症状**: 应用响应时间变慢

**解决方案**:
```go
// 启用异步收集
config.AsyncCollection = true
config.CollectionInterval = 60 * time.Second

// 减少监控频率
config.CollectionBatchSize = 100

// 禁用详细标签
// 使用最小标签集
```

#### 3. 监控数据丢失

**症状**: 部分操作没有监控记录

**解决方案**:
```go
// 检查采样率
if config.MonitoringSamplingRate < 1.0 {
    // 增加采样率或检查采样逻辑
    config.MonitoringSamplingRate = 0.5
}

// 检查异步收集延迟
config.CollectionInterval = 10 * time.Second

// 检查导出处理器
if collector.GetExportHandler() == nil {
    // 设置导出处理器
    collector.SetExportHandler(defaultExportHandler)
}
```

#### 4. 外部监控系统连接问题

**症状**: 监控数据无法导出到外部系统

**解决方案**:
```go
// 增加超时时间
exporter.client.Timeout = 30 * time.Second

// 添加重试逻辑
func exportWithRetry(exporter Exporter, metrics []Metric, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        if err := exporter.Export(metrics); err == nil {
            return nil
        }
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return fmt.Errorf("导出失败，重试 %d 次后放弃", maxRetries)
}

// 添加缓冲队列
type BufferedExporter struct {
    exporter Exporter
    buffer   chan []Metric
    maxSize  int
}

func NewBufferedExporter(exporter Exporter, bufferSize int) *BufferedExporter {
    be := &BufferedExporter{
        exporter: exporter,
        buffer:   make(chan []Metric, bufferSize),
        maxSize:  bufferSize,
    }
    go be.processBuffer()
    return be
}

func (be *BufferedExporter) processBuffer() {
    for metrics := range be.buffer {
        if err := be.exporter.Export(metrics); err != nil {
            log.Printf("导出失败: %v", err)
        }
    }
}
```

### 监控系统健康检查

```go
// monitoring/health_check.go
package monitoring

import (
    "time"
)

type HealthChecker struct {
    collector Collector
    lastCheck time.Time
    metrics   map[string]interface{}
}

func NewHealthChecker(collector Collector) *HealthChecker {
    return &HealthChecker{
        collector: collector,
        metrics:   make(map[string]interface{}),
    }
}

func (hc *HealthChecker) Check() map[string]interface{} {
    now := time.Now()
    
    // 收集健康指标
    hc.metrics["last_check"] = now
    hc.metrics["time_since_last_check"] = now.Sub(hc.lastCheck).String()
    
    // 检查收集器状态
    if hc.collector == nil {
        hc.metrics["collector_status"] = "not_initialized"
    } else {
        hc.metrics["collector_status"] = "active"
        
        // 获取收集器统计
        if stats, ok := hc.collector.(interface{ GetStats() map[string]interface{} }); ok {
            hc.metrics["collector_stats"] = stats.GetStats()
        }
    }
    
    hc.lastCheck = now
    return hc.metrics
}

func (hc *HealthChecker) GetMetrics() map[string]interface{} {
    return hc.metrics
}
```

## 总结

MagicORM 监控系统提供了灵活、高效的监控数据收集能力。通过合理的配置和优化，可以在生产环境中实现：

1. **低开销监控**: 通过采样和异步收集减少性能影响
2. **灵活集成**: 支持多种外部监控系统
3. **可扩展性**: 易于添加新的监控维度和指标
4. **可靠性**: 内置错误处理和故障恢复机制

根据应用的具体需求和负载情况，调整监控配置以获得最佳的性能和监控效果。