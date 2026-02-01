# MagicORM 验证系统监控模块

## 概述

监控模块为MagicORM验证系统提供全面的监控、日志和指标收集功能。支持结构化日志、性能指标收集、Prometheus导出和健康检查。

## 功能特性

### 1. 结构化日志
- 多级别日志（DEBUG, INFO, WARN, ERROR, FATAL）
- 结构化字段支持
- 终端彩色输出和JSON格式
- 多输出目标（控制台、文件、网络）

### 2. 性能指标收集
- 验证操作统计（总数、成功率、平均时间）
- 缓存效率指标（命中率、访问次数）
- 错误分类统计
- 各验证层性能分析
- 资源使用监控（内存、并发数）

### 3. 指标导出
- Prometheus格式导出
- JSON API端点
- 健康检查端点
- 自定义标签支持

### 4. 集成支持
- 与验证管理器无缝集成
- 可配置的监控级别
- 生产环境就绪

## 快速开始

### 基本使用

```go
import "github.com/muidea/magicOrm/validation/monitoring"

// 1. 创建监控组件
metrics := monitoring.NewMetricsCollector()
logger := monitoring.NewValidationLogger("info", nil).WithMetrics(metrics)

// 2. 记录验证操作
logger.LogValidation(
    "insert",
    "User",
    "insert",
    50*time.Millisecond,
    nil, // 无错误
    map[string]interface{}{
        "user_id": 123,
        "fields":  []string{"name", "email"},
    },
)

// 3. 获取指标
currentMetrics, _ := logger.GetMetrics()
fmt.Printf("Total validations: %d\n", currentMetrics.TotalValidations)
fmt.Printf("Cache hit rate: %.2f%%\n", currentMetrics.CacheHitRate*100)
```

### 集成到验证系统

```go
import (
    "github.com/muidea/magicOrm/validation"
    "github.com/muidea/magicOrm/validation/monitoring"
)

func SetupValidationWithMonitoring() validation.ValidationManager {
    // 创建监控组件
    metrics := monitoring.NewMetricsCollector()
    logger := monitoring.NewValidationLogger("info", nil).WithMetrics(metrics)
    
    // 配置验证管理器
    config := validation.ValidationConfig{
        EnableMetrics: true,
        Logger:        logger,
        EnableCaching: true,
        CacheTTL:      10 * time.Minute,
    }
    
    // 创建验证管理器
    factory := validation.NewValidationFactory()
    manager := factory.CreateValidationManager(config)
    
    // 启动指标导出器（可选）
    exporter, _ := monitoring.StartDefaultExporter(metrics, logger)
    
    // 在应用关闭时停止导出器
    // defer exporter.Stop()
    
    return manager
}
```

## 组件详解

### 1. MetricsCollector（指标收集器）

收集和统计验证系统的各种指标。

```go
// 创建收集器
metrics := monitoring.NewMetricsCollector()

// 记录验证操作（通常通过Logger自动记录）
metrics.RecordValidation(50*time.Millisecond, nil)

// 记录缓存访问
metrics.RecordCacheHit()
metrics.RecordCacheMiss()

// 记录层性能
metrics.RecordLayerTime("type", 10*time.Millisecond)

// 获取当前指标
currentMetrics := metrics.GetMetrics()
```

**收集的指标包括**：
- 验证总数和速率
- 平均验证时间
- 缓存命中率和访问次数
- 错误总数和分类
- 各验证层性能
- 并发验证数
- 内存使用量
- 系统运行时间

### 2. ValidationLogger（验证日志器）

提供结构化日志记录，自动集成指标收集。

```go
// 创建日志器
logger := monitoring.NewValidationLogger("info", os.Stdout)

// 附加指标收集器
logger.WithMetrics(metrics)

// 添加全局字段
logger.WithFields(map[string]interface{}{
    "service": "validation",
    "env":     "production",
})

// 记录验证操作
logger.LogValidation(
    operation string,
    modelName string,
    scenario string,
    duration time.Duration,
    err error,
    fields map[string]interface{},
)

// 记录缓存访问
logger.LogCacheAccess(
    operation string,
    cacheType string,
    key string,
    hit bool,
    duration time.Duration,
    fields map[string]interface{},
)

// 记录层性能
logger.LogLayerPerformance(
    layer string,
    operation string,
    duration time.Duration,
    success bool,
    fields map[string]interface{},
)
```

**日志级别**：
- `DEBUG`: 调试信息，详细操作日志
- `INFO`: 常规信息，成功操作
- `WARN`: 警告信息，非关键问题
- `ERROR`: 错误信息，操作失败
- `FATAL`: 致命错误，系统无法继续

### 3. MetricsExporter（指标导出器）

通过HTTP导出指标，支持Prometheus和JSON格式。

```go
// 创建导出器
config := monitoring.ExportConfig{
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

exporter := monitoring.NewMetricsExporter(metrics, logger, config)

// 添加自定义标签
exporter.WithLabels(map[string]string{
    "application": "magicorm",
    "component":   "validation",
    "instance":    "instance-1",
})

// 启动导出器
exporter.Start()

// 停止导出器（在应用关闭时）
defer exporter.Stop()
```

**提供的端点**：
- `GET /metrics`: Prometheus格式指标
- `GET /metrics/json`: JSON格式指标
- `GET /health`: 健康检查
- `GET /`: 信息页面

## 配置指南

### 日志配置

```go
// 控制台日志（默认）
logger := monitoring.NewValidationLogger("info", os.Stdout)

// 文件日志
fileLogger, err := monitoring.FileLogger("debug", "/var/log/validation.log")

// 多输出日志
multiLogger := monitoring.MultiLogger("info", os.Stdout, fileWriter, networkWriter)

// 设置日志级别
logger.SetLevel("debug")  // DEBUG, INFO, WARN, ERROR, FATAL
```

### 导出器配置

```go
config := monitoring.ExportConfig{
    Enabled:          true,           // 启用导出器
    Port:             9090,           // 监听端口
    Path:             "/metrics",     // Prometheus端点路径
    HealthCheckPath:  "/health",      // 健康检查路径
    MetricsPath:      "/metrics/json", // JSON端点路径
    EnablePrometheus: true,           // 启用Prometheus导出
    EnableJSON:       true,           // 启用JSON导出
    RefreshInterval:  30 * time.Second, // 指标刷新间隔
    Timeout:          10 * time.Second, // HTTP超时时间
}
```

### 生产环境配置示例

```go
func ProductionSetup() {
    // 1. 创建监控组件
    metrics := monitoring.NewMetricsCollector()
    
    // 2. 创建文件和控制台日志器
    fileLogger, _ := monitoring.FileLogger("warn", "/var/log/magicorm/validation.log")
    consoleLogger := monitoring.NewValidationLogger("error", os.Stderr)
    multiLogger := monitoring.MultiLogger("warn", consoleLogger, fileLogger)
    
    // 3. 附加指标
    logger := multiLogger.WithMetrics(metrics)
    
    // 4. 添加环境标签
    logger.WithFields(map[string]interface{}{
        "environment": "production",
        "region":      "us-west-2",
        "version":     "1.0.0",
    })
    
    // 5. 配置导出器
    exporterConfig := monitoring.ExportConfig{
        Enabled:          true,
        Port:             9100, // 避免与其它服务冲突
        Path:             "/validation/metrics",
        HealthCheckPath:  "/validation/health",
        MetricsPath:      "/validation/metrics/json",
        EnablePrometheus: true,
        EnableJSON:       true,
        RefreshInterval:  15 * time.Second,
        Timeout:          5 * time.Second,
    }
    
    exporter := monitoring.NewMetricsExporter(metrics, logger, exporterConfig)
    exporter.WithLabels(map[string]string{
        "app":       "magicorm",
        "component": "validation",
        "env":       "prod",
    })
    
    // 6. 启动
    exporter.Start()
    
    // 返回配置好的日志器供验证系统使用
    return logger
}
```

## 监控指标详解

### Prometheus指标

验证系统导出以下Prometheus指标：

#### 性能指标
```
# HELP validation_total_validations Total number of validations performed
# TYPE validation_total_validations counter
validation_total_validations{app="magicorm",component="validation"} 1234

# HELP validation_rate_validations_per_second Validation rate in validations per second  
# TYPE validation_rate_validations_per_second gauge
validation_rate_validations_per_second{app="magicorm",component="validation"} 5.67

# HELP validation_average_duration_seconds Average validation duration in seconds
# TYPE validation_average_duration_seconds gauge
validation_average_duration_seconds{app="magicorm",component="validation"} 0.045
```

#### 缓存指标
```
# HELP validation_cache_hits_total Total number of cache hits
# TYPE validation_cache_hits_total counter
validation_cache_hits_total{app="magicorm",component="validation"} 890

# HELP validation_cache_misses_total Total number of cache misses
# TYPE validation_cache_misses_total counter
validation_cache_misses_total{app="magicorm",component="validation"} 110

# HELP validation_cache_hit_rate Cache hit rate (0-1)
# TYPE validation_cache_hit_rate gauge
validation_cache_hit_rate{app="magicorm",component="validation"} 0.89
```

#### 错误指标
```
# HELP validation_errors_total Total number of validation errors
# TYPE validation_errors_total counter
validation_errors_total{app="magicorm",component="validation"} 45

# HELP validation_error_rate Validation error rate (0-1)
# TYPE validation_error_rate gauge
validation_error_rate{app="magicorm",component="validation"} 0.036

# HELP validation_errors_by_type_total Total errors by type
# TYPE validation_errors_by_type_total counter
validation_errors_by_type_total{app="magicorm",component="validation",error_type="type_validation"} 15
validation_errors_by_type_total{app="magicorm",component="validation",error_type="constraint_validation"} 25
validation_errors_by_type_total{app="magicorm",component="validation",error_type="database_validation"} 5
```

#### 资源指标
```
# HELP validation_concurrent_validations_current Current number of concurrent validations
# TYPE validation_concurrent_validations_current gauge
validation_concurrent_validations_current{app="magicorm",component="validation"} 8

# HELP validation_concurrent_validations_peak Peak number of concurrent validations
# TYPE validation_concurrent_validations_peak gauge
validation_concurrent_validations_peak{app="magicorm",component="validation"} 42

# HELP validation_memory_usage_bytes Memory usage in bytes
# TYPE validation_memory_usage_bytes gauge
validation_memory_usage_bytes{app="magicorm",component="validation"} 16777216
```

### JSON指标端点

`GET /metrics/json` 返回完整的JSON指标：

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "metrics": {
    "total_validations": 1234,
    "validation_rate": 5.67,
    "average_validation_time": "45ms",
    "cache_hits": 890,
    "cache_misses": 110,
    "cache_hit_rate": 0.89,
    "total_errors": 45,
    "error_rate": 0.036,
    "errors_by_type": {
      "type_validation": 15,
      "constraint_validation": 25,
      "database_validation": 5
    },
    "current_concurrent_validations": 8,
    "peak_concurrent_validations": 42,
    "memory_usage": 16777216,
    "uptime": "2h30m15s"
  },
  "labels": {
    "app": "magicorm",
    "component": "validation",
    "env": "production"
  }
}
```

## 告警规则示例

### Prometheus告警规则

```yaml
groups:
  - name: validation_alerts
    rules:
      # 高错误率告警
      - alert: HighValidationErrorRate
        expr: validation_error_rate > 0.05  # 错误率超过5%
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High validation error rate"
          description: "Validation error rate is {{ $value | humanizePercentage }}"
      
      # 低缓存命中率告警
      - alert: LowCacheHitRate
        expr: validation_cache_hit_rate < 0.6  # 命中率低于60%
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit rate"
          description: "Cache hit rate is {{ $value | humanizePercentage }}"
      
      # 高验证延迟告警
      - alert: HighValidationLatency
        expr: validation_average_duration_seconds > 0.1  # 平均时间超过100ms
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High validation latency"
          description: "Average validation time is {{ $value }} seconds"
      
      # 高并发告警
      - alert: HighConcurrentValidations
        expr: validation_concurrent_validations_current > 50  # 并发数超过50
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High concurrent validations"
          description: "{{ $value }} concurrent validations detected"
      
      # 系统宕机告警
      - alert: ValidationSystemDown
        expr: up{job="validation"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Validation system is down"
```

## 故障排除

### 常见问题

#### 1. 指标导出器无法启动
**症状**: `Failed to start metrics exporter`
**可能原因**: 端口被占用或权限不足
**解决方案**:
```go
// 检查端口是否可用
config.Port = 9091 // 尝试不同端口

// 或以非特权端口启动
if config.Port < 1024 {
    config.Port = 8080
}
```

#### 2. 内存使用过高
**症状**: 内存持续增长
**可能原因**: 指标收集器未清理旧数据
**解决方案**:
```go
// 定期重置指标
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        logger.ResetMetrics()
    }
}()

// 或限制指标保留时间
```

#### 3. 日志文件过大
**症状**: 日志文件快速增长
**可能原因**: 日志级别过低或日志过多
**解决方案**:
```go
// 提高日志级别
logger.SetLevel("warn") // 只记录警告和错误

// 或使用日志轮转
// 建议使用外部工具如logrotate
```

#### 4. Prometheus指标格式错误
**症状**: Prometheus无法解析指标
**可能原因**: 指标标签包含非法字符
**解决方案**:
```go
// 确保标签值合法
exporter.WithLabels(map[string]string{
    "app":       "magicorm",
    "component": "validation",
    // 避免特殊字符和空格
})
```

### 诊断命令

```bash
# 检查导出器是否运行
curl http://localhost:9090/health

# 获取Prometheus指标
curl http://localhost:9090/metrics

# 获取JSON指标
curl http://localhost:9090/metrics/json

# 检查日志文件
tail -f /var/log/magicorm/validation.log

# 监控内存使用
ps aux | grep magicorm | grep -v grep
```

## 最佳实践

### 1. 生产环境部署
- 使用文件日志而非控制台日志
- 设置适当的日志级别（WARN或ERROR）
- 配置日志轮转
- 使用非特权端口（>1024）
- 添加适当的防火墙规则

### 2. 监控配置
- 根据负载调整指标收集频率
- 设置合理的告警阈值
- 定期审查监控仪表板
- 保留历史指标用于趋势分析

### 3. 性能考虑
- 在高负载环境中禁用详细日志
- 限制并发指标收集
- 使用缓存减少重复计算
- 定期清理旧指标数据

### 4. 安全考虑
- 保护监控端点访问
- 不要记录敏感数据
- 使用HTTPS保护API端点
- 定期审计日志内容

## 扩展和自定义

### 自定义指标

```go
// 扩展MetricsCollector
type CustomMetricsCollector struct {
    *monitoring.MetricsCollector
    customCounter int64
}

func (c *CustomMetricsCollector) RecordCustomEvent() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.customCounter++
}

// 扩展ValidationLogger
type CustomLogger struct {
    *monitoring.ValidationLogger
}

func (l *CustomLogger) LogBusinessEvent(event string, data map[string]interface{}) {
    l.Info("Business event: "+event, data)
}
```

### 集成外部系统

```go
// 集成到现有监控系统
func IntegrateWithExistingMonitoring(existingMetrics prometheus.Registerer) {
    metrics := monitoring.NewMetricsCollector()
    
    // 定期将指标推送到现有系统
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            current := metrics.GetMetrics()
            // 转换并推送到现有系统
            pushMetricsToExistingSystem(current)
        }
    }()
}
```

## 相关文档

- [生产环境指南](../PRODUCTION_GUIDE.md)
- [验证系统架构](../VALIDATION_ARCHITECTURE.md)
- [API参考](../API_REFERENCE.md)
- [示例代码](./example.go)

## 支持

如遇问题，请：
1. 检查日志文件获取详细错误信息
2. 验证配置是否正确
3. 参考本文档的故障排除部分
4. 联系开发团队并提供：
   - 错误日志
   - 相关配置
   - 复现步骤
   - 监控指标截图