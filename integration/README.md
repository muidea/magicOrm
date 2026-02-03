# MagicORM 与 magicCommon/monitoring 集成

本文档介绍如何将 MagicORM 的监控系统与 magicCommon 的监控系统集成。

## 概述

MagicORM 提供了简洁的监控系统，专注于数据收集。通过与 magicCommon/monitoring 集成，可以实现：

1. **统一监控管理**：使用 magicCommon 的统一监控管理器
2. **丰富的指标类型**：利用 magicCommon 的指标定义和收集功能
3. **灵活的导出选项**：支持 Prometheus、JSON 等多种导出格式
4. **提供者架构**：可扩展的监控提供者系统

## 架构设计

### 集成组件

```
integration/
├── magiccommon_integration.go    # 主要集成逻辑
├── integration_test.go           # 集成测试
└── README.md                     # 本文档
```

### 数据流

```
MagicORM 监控操作
        ↓
MagicCommonCollectorAdapter (适配器)
        ↓
magicCommon/monitoring 收集器
        ↓
magicCommon 监控管理器
        ↓
外部监控系统 (Prometheus, Datadog, 等)
```

## 快速开始

### 基本集成

```go
import (
    "github.com/muidea/magicOrm/integration"
    "github.com/muidea/magicCommon/monitoring/core"
    "time"
)

func main() {
    // 创建 magicCommon 监控配置
    config := core.DefaultMonitoringConfig()
    config.Namespace = "magicorm"
    config.AsyncCollection = true
    config.CollectionInterval = 30 * time.Second

    // 创建集成实例
    integration, err := integration.NewMagicCommonIntegration(&config)
    if err != nil {
        log.Fatal(err)
    }
    defer integration.Stop()

    // 获取 MagicORM 收集器适配器
    collector := integration.GetMagicORMCollector()

    // 使用收集器记录操作
    collector.RecordORMOperation(
        monitoring.OperationInsert,
        "User",
        true,
        150*time.Millisecond,
        nil,
        map[string]string{"database": "postgresql"},
    )
}
```

### 与 MagicORM 组件集成

```go
import (
    "github.com/muidea/magicOrm/integration"
    "github.com/muidea/magicOrm/monitoring/orm"
    "github.com/muidea/magicOrm/monitoring/validation"
    "github.com/muidea/magicOrm/monitoring/database"
)

func setupMonitoredComponents() {
    // 创建集成
    integration, _ := integration.NewMagicCommonIntegration(nil)
    collector := integration.GetMagicORMCollector()

    // 创建带监控的 ORM 装饰器
    ormDecorator := orm.NewDecorator(collector)

    // 创建带监控的验证适配器
    validationAdapter := validation.NewAdapter(collector)

    // 创建带监控的数据库适配器
    databaseAdapter := database.NewAdapter(collector, "postgresql")

    // 使用这些组件...
}
```

## 配置选项

### 监控配置

```go
config := core.DefaultMonitoringConfig()

// 基本配置
config.Namespace = "magicorm"           // 指标命名空间
config.Enabled = true                   // 启用监控
config.SamplingRate = 1.0               // 采样率 (0.0-1.0)

// 性能配置
config.AsyncCollection = true           // 异步收集
config.CollectionInterval = 30 * time.Second  // 收集间隔
config.BatchSize = 1000                 // 批处理大小
config.BufferSize = 5000                // 缓冲区大小

// 数据保留
config.RetentionPeriod = 1 * time.Hour  // 数据保留时间

// 导出配置
config.ExportConfig.Enabled = true      // 启用导出
config.ExportConfig.Port = 9090         // Prometheus 端口
config.ExportConfig.Path = "/metrics"   // 指标路径
```

### 环境特定配置

```go
// 开发环境配置
devConfig := core.DevelopmentConfig()
devConfig.Namespace = "magicorm_dev"
devConfig.SamplingRate = 1.0  // 100% 采样

// 生产环境配置
prodConfig := core.ProductionConfig()
prodConfig.Namespace = "magicorm_prod"
prodConfig.SamplingRate = 0.1  // 10% 采样

// 高负载环境配置
highLoadConfig := core.HighLoadConfig()
highLoadConfig.Namespace = "magicorm_highload"
highLoadConfig.SamplingRate = 0.01  // 1% 采样
```

## 监控指标

### ORM 指标

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `magicorm_operations_total` | Counter | ORM 操作总数 | `model`, `operation`, `success` |
| `magicorm_operation_duration_seconds` | Gauge | ORM 操作延迟 | `model`, `operation`, `success` |
| `magicorm_errors_total` | Counter | ORM 错误总数 | `model`, `operation`, `error_type`, `error_code` |

### 验证指标

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `magicorm_validation_operations_total` | Counter | 验证操作总数 | `validator`, `model`, `scenario` |
| `magicorm_validation_duration_seconds` | Gauge | 验证延迟 | `validator`, `model`, `scenario` |
| `magicorm_validation_errors_total` | Counter | 验证错误总数 | `validator`, `model`, `scenario`, `error_type`, `error_code` |

### 数据库指标

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `magicorm_database_queries_total` | Counter | 数据库查询总数 | `database`, `query_type`, `success` |
| `magicorm_database_query_duration_seconds` | Gauge | 数据库查询延迟 | `database`, `query_type`, `success` |
| `magicorm_database_rows_affected` | Gauge | 影响行数 | `database`, `query_type`, `success` |
| `magicorm_database_errors_total` | Counter | 数据库错误总数 | `database`, `query_type`, `error_type`, `error_code` |

## 高级用法

### 自定义指标提供者

```go
// 自定义监控提供者
type CustomProvider struct {
    name string
}

func (p *CustomProvider) Name() string {
    return p.name
}

func (p *CustomProvider) Init(collector *core.Collector) *types.Error {
    // 注册自定义指标定义
    definitions := []types.MetricDefinition{
        types.NewCounterDefinition(
            "custom_operations_total",
            "Custom operations total",
            []string{"custom_label"},
            nil,
        ),
    }
    
    for _, def := range definitions {
        if err := collector.RegisterDefinition(def); err != nil {
            return err
        }
    }
    
    return nil
}

func (p *CustomProvider) Metrics() []types.MetricDefinition {
    // 返回指标定义
    return []types.MetricDefinition{
        types.NewCounterDefinition(
            "custom_operations_total",
            "Custom operations total",
            []string{"custom_label"},
            nil,
        ),
    }
}

func (p *CustomProvider) Collect() ([]types.Metric, *types.Error) {
    // 收集自定义指标
    return []types.Metric{
        {
            Name:   "custom_operations_total",
            Value:  1.0,
            Labels: map[string]string{"custom_label": "test"},
        },
    }, nil
}

func (p *CustomProvider) Shutdown() *types.Error {
    return nil
}

// 注册自定义提供者
func registerCustomProvider(integration *integration.MagicCommonIntegration) {
    collector := integration.GetManager().GetCollector()
    customProvider := &CustomProvider{name: "custom_provider"}
    collector.RegisterProvider(customProvider)
}
```

### 全局集成实例

```go
// 初始化全局集成
func init() {
    config := core.DefaultMonitoringConfig()
    config.Namespace = "magicorm"
    
    if err := integration.InitializeGlobalIntegration(&config); err != nil {
        log.Printf("初始化全局集成失败: %v", err)
    }
}

// 在应用中使用
func someFunction() {
    integration := integration.GetGlobalIntegration()
    if integration == nil {
        return
    }
    
    collector := integration.GetMagicORMCollector()
    collector.RecordORMOperation(...)
}

// 应用退出时清理
func cleanup() {
    integration.ShutdownGlobalIntegration()
}
```

### 性能监控

```go
// 监控集成性能
func monitorIntegrationPerformance(integration *integration.MagicCommonIntegration) {
    manager := integration.GetManager()
    
    // 获取管理器统计
    stats := manager.GetStats()
    log.Printf("管理器统计: %+v", stats)
    
    // 获取提供者健康状态
    health := manager.GetProviderHealth()
    for name, status := range health {
        log.Printf("提供者 %s 健康状态: %v", name, status)
    }
    
    // 获取收集器统计
    collector := integration.GetManager().GetCollector()
    collectorStats := collector.GetStats()
    log.Printf("收集器统计: %+v", collectorStats)
    
    // 监控缓冲区使用率
    bufferUsage := collector.GetBufferUsage()
    if bufferUsage > 0.8 {
        log.Printf("警告: 缓冲区使用率过高: %.2f%%", bufferUsage*100)
    }
}
```

## 测试

### 运行集成测试

```bash
# 运行所有集成测试
go test ./integration/... -v

# 运行特定测试
go test ./integration/... -v -run TestMagicCommonIntegration

# 运行基准测试
go test ./integration/... -bench=. -benchmem
```

### 测试覆盖率

```bash
# 生成测试覆盖率报告
go test ./integration/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## 故障排除

### 常见问题

#### 1. 集成初始化失败

**症状**: `NewMagicCommonIntegration` 返回错误

**解决方案**:
```go
// 检查配置
config := core.DefaultMonitoringConfig()
if err := config.Validate(); err != nil {
    log.Fatalf("配置验证失败: %v", err)
}

// 尝试使用默认配置
integration, err := integration.NewMagicCommonIntegration(nil)
```

#### 2. 指标没有记录

**症状**: 操作后获取不到指标

**解决方案**:
```go
// 检查采样率
config.SamplingRate = 1.0  // 设置为 100% 采样

// 检查异步收集
config.AsyncCollection = false  // 禁用异步收集进行测试

// 手动触发收集
collector := integration.GetManager().GetCollector()
collector.ForceFlush()
```

#### 3. 内存使用过高

**症状**: 应用内存持续增长

**解决方案**:
```go
// 减少数据保留时间
config.RetentionPeriod = 30 * time.Minute

// 减少缓冲区大小
config.BufferSize = 1000

// 降低采样率
config.SamplingRate = 0.1  // 10% 采样
```

#### 4. 性能影响明显

**症状**: 应用响应时间变慢

**解决方案**:
```go
// 启用异步收集
config.AsyncCollection = true
config.CollectionInterval = 60 * time.Second

// 增加批处理大小
config.BatchSize = 5000

// 使用高负载配置
config := core.HighLoadConfig()
```

### 调试日志

```go
// 启用详细日志
func enableDebugLogging() {
    // magicCommon 可能提供日志配置
    // 检查文档获取详细信息
}

// 手动检查状态
func checkIntegrationStatus(integration *integration.MagicCommonIntegration) {
    manager := integration.GetManager()
    
    // 检查是否运行
    if !manager.IsRunning() {
        log.Println("监控管理器未运行")
    }
    
    // 检查是否初始化
    if !manager.IsInitialized() {
        log.Println("监控管理器未初始化")
    }
    
    // 获取详细统计
    stats := manager.GetStats()
    log.Printf("详细统计: %+v", stats)
}
```

## 最佳实践

### 1. 环境特定配置

```go
func getMonitoringConfig(env string) *core.MonitoringConfig {
    switch env {
    case "development":
        config := core.DevelopmentConfig()
        config.Namespace = "magicorm_dev"
        config.SamplingRate = 1.0
        return &config
        
    case "production":
        config := core.ProductionConfig()
        config.Namespace = "magicorm_prod"
        config.SamplingRate = 0.1
        return &config
        
    case "highload":
        config := core.HighLoadConfig()
        config.Namespace = "magicorm_highload"
        config.SamplingRate = 0.01
        return &config
        
    default:
        config := core.DefaultMonitoringConfig()
        config.Namespace = "magicorm"
        return &config
    }
}
```

### 2. 优雅关闭

```go
func main() {
    // 创建集成
    integration, err := integration.NewMagicCommonIntegration(nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // 设置信号处理
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    // 等待信号
    <-sigChan
    
    // 优雅关闭
    log.Println("收到关闭信号，开始优雅关闭...")
    
    if err := integration.Stop(); err != nil {
        log.Printf("关闭集成时出错: %v", err)
    }
    
    log.Println("集成已关闭")
}
```

### 3. 监控集成本身

```go
// 监控集成健康状况
func monitorIntegrationHealth(integration *integration.MagicCommonIntegration) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        manager := integration.GetManager()
        
        // 检查运行状态
        if !manager.IsRunning() {
            log.Println("警告: 监控管理器未运行")
        }
        
        // 检查提供者健康
        health := manager.GetProviderHealth()
        for name, status := range health {
            if !status.Healthy {
                log.Printf("警告: 提供者 %s 不健康: %v", name, status.LastError)
            }
        }
        
        // 检查收集器状态
        collector := integration.GetManager().GetCollector()
        stats := collector.GetStats()
        
        if stats.MetricsDropped > 0 {
            log.Printf("信息: 丢弃了 %d 个指标", stats.MetricsDropped)
        }
        
        // 检查缓冲区
        if collector.IsBufferFull() {
            log.Println("警告: 收集器缓冲区已满")
        }
    }
}
```

## 迁移指南

### 从独立监控迁移到集成监控

#### 之前（独立监控）:
```go
import "github.com/muidea/magicOrm/monitoring"

// 创建独立收集器
collector := monitoring.NewCollector()

// 记录操作
collector.RecordORMOperation(...)
```

#### 之后（集成监控）:
```go
import (
    "github.com/muidea/magicOrm/integration"
    "github.com/muidea/magicCommon/monitoring/core"
)

// 创建集成
config := core.DefaultMonitoringConfig()
integration, _ := integration.NewMagicCommonIntegration(&config)

// 获取适配的收集器
collector := integration.GetMagicORMCollector()

// 记录操作（API 兼容）
collector.RecordORMOperation(...)
```

### 迁移步骤

1. **添加依赖**: 确保项目依赖 magicCommon
2. **更新导入**: 将独立监控导入改为集成导入
3. **初始化集成**: 在应用启动时初始化集成
4. **获取收集器**: 使用集成提供的收集器
5. **测试验证**: 运行测试确保功能正常

## 性能考虑

### 内存使用

- **默认配置**: 适合大多数应用
- **高负载环境**: 使用 `HighLoadConfig` 并降低采样率
- **内存限制**: 调整 `BufferSize` 和 `RetentionPeriod`

### CPU 使用

- **异步收集**: 减少对业务逻辑的影响
- **批处理**: 使用合适的 `BatchSize`
- **采样率**: 根据负载调整采样率

### 网络开销

- **导出频率**: 调整 `CollectionInterval`
- **数据量**: 使用标签过滤不重要的指标
- **压缩**: 考虑启用导出数据的压缩

## 扩展性

### 添加新的监控维度

```go
// 扩展集成以支持新的监控类型
type ExtendedIntegration struct {
    *integration.MagicCommonIntegration
    // 添加扩展字段
}

func NewExtendedIntegration(config *core.MonitoringConfig) (*ExtendedIntegration, *types.Error) {
    base, err := integration.NewMagicCommonIntegration(config)
    if err != nil {
        return nil, err
    }
    
    return &ExtendedIntegration{
        MagicCommonIntegration: base,
    }, nil
}

// 添加新的监控方法
func (e *ExtendedIntegration) RecordCustomOperation(
    operationType string,
    customLabels map[string]string,
) {
    collector := e.GetMagicORMCollector().(*integration.MagicCommonCollectorAdapter)
    
    // 使用底层的 magicCommon 收集器
    // 记录自定义指标...
}
```

## 支持与反馈

### 获取帮助

- 查看 magicCommon 文档
- 检查集成测试示例
- 查看监控指标定义

### 报告问题

如果遇到问题，请提供：

1. 使用的配置
2. 错误信息
3. 复现步骤
4. 环境信息

### 贡献

欢迎贡献代码改进、文档更新或新的集成功能。