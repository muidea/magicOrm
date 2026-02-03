# MagicORM 监控系统

MagicORM 的监控系统，专注于数据收集，支持 ORM 操作、验证系统和数据库执行的监控。

## 概述

监控系统提供：

1. **简洁的数据收集**：专注于收集监控数据，不负责导出和管理
2. **ORM 操作监控**：跟踪 CRUD 操作、事务和性能
3. **验证系统监控**：监控验证性能和缓存效果
4. **数据库执行监控**：跟踪数据库查询、连接和事务
5. **性能指标**：延迟、吞吐量、错误率和资源使用情况

## 架构设计

**核心原则**：MagicORM 提供简单、独立的监控数据收集器，专注于数据收集而不负责导出和管理。监控数据由外部系统处理。

**新架构特点**：
1. **模块化设计**：每个组件（ORM、验证、数据库）有独立的收集器
2. **简单接口**：清晰的API，易于使用和集成
3. **无全局状态**：避免测试中的状态污染问题
4. **向后兼容**：保持与现有代码的兼容性

**文件结构**：
```
monitoring/
├── collector.go                    # 顶层接口和类型定义
├── init.go                         # 初始化集成
├── e2e_test.go                     # 端到端测试
├── core/                           # 核心类型和简单收集器实现
├── orm/                            # ORM监控收集器
├── validation/                     # 验证监控收集器
├── database/                       # 数据库监控收集器
└── example/                        # 使用示例
```

## 快速开始

### 基本使用

```go
import (
    "time"
    
    "github.com/muidea/magicOrm/monitoring"
    "github.com/muidea/magicOrm/monitoring/orm"
    "github.com/muidea/magicOrm/monitoring/validation"
    "github.com/muidea/magicOrm/monitoring/database"
)

// 创建独立的收集器
ormCollector := orm.NewCollector()
valCollector := validation.NewCollector()
dbCollector := database.NewCollector()

// 记录ORM操作
ormCollector.RecordOperation(
    monitoring.OperationInsert,
    "User",
    time.Now(),
    nil,
    map[string]string{"database": "postgresql"},
)

// 记录验证操作
valCollector.RecordValidation(
    "validate_user",
    "User",
    "insert",
    time.Now(),
    nil,
    map[string]string{"field_count": "5"},
)

// 记录数据库操作
dbCollector.RecordQuery(
    "postgresql",
    "SELECT",
    10, // rowsAffected
    time.Now(),
    nil,
    map[string]string{"table": "users"},
)
```

### ORM 监控

```go
import "github.com/muidea/magicOrm/monitoring/orm"

// 创建ORM收集器
collector := orm.NewCollector()

// 记录ORM操作
collector.RecordOperation(
    monitoring.OperationInsert,
    "User",
    time.Now(),
    nil,
    map[string]string{"test": "example"},
)

// 获取收集的指标
metrics, err := collector.GetMetrics()
if err != nil {
    // 处理错误
}
```

### 数据库监控

```go
import "github.com/muidea/magicOrm/monitoring/database"

// 创建数据库收集器
collector := database.NewCollector()

// 记录数据库查询
collector.RecordQuery(
    "postgresql",
    "SELECT",
    10, // rowsAffected
    time.Now(),
    nil,
    map[string]string{"table": "users"},
)

// 记录数据库事务
collector.RecordTransaction(
    "postgresql",
    "BEGIN",
    time.Now(),
    nil,
    map[string]string{"test": "transaction"},
)
```

### 验证监控

```go
import "github.com/muidea/magicOrm/monitoring/validation"

// 创建验证收集器
collector := validation.NewCollector()

// 记录验证操作
collector.RecordValidation(
    "validate_user",
    "User",
    "insert",
    time.Now(),
    nil,
    map[string]string{"field_count": "5"},
)

// 记录缓存访问
collector.RecordCacheAccess(
    "User",
    "insert",
    true, // hit
    time.Now(),
    map[string]string{"cache_type": "memory"},
)
```

## 监控数据类型

### ORM 操作监控

- **操作类型**: Insert, Update, Delete, Query, BatchQuery
- **指标**: 成功率、延迟、错误类型
- **标签**: 模型名称、数据库类型、操作类型

### 验证系统监控

- **场景**: Insert, Update, Query, Delete
- **指标**: 验证延迟、缓存命中率、错误统计
- **标签**: 验证器名称、模型名称、场景类型

### 数据库执行监控

- **查询类型**: Select, Insert, Update, Delete, Transaction
- **指标**: 查询延迟、返回行数、连接状态
- **标签**: 数据库类型、表名、操作类型

## 标签系统

支持灵活的标签系统，用于分类和过滤监控数据：

```go
// 基本标签
labels := map[string]string{
    "database": "postgresql",
    "table":    "users",
    "operation": "insert",
}

// 合并默认标签
collector.WithDefaultLabels(map[string]string{
    "environment": "production",
    "service":     "user-service",
})

// 记录带标签的操作
collector.RecordORMOperation(
    monitoring.OperationInsert,
    "User",
    true,
    150*time.Millisecond,
    nil,
    labels,
)
```

## 错误处理

监控系统支持详细的错误分类：

```go
// 错误类型定义
type ErrorType string

const (
    ErrorTypeDatabase   ErrorType = "database"
    ErrorTypeValidation ErrorType = "validation"
    ErrorTypeConstraint ErrorType = "constraint"
    ErrorTypeType       ErrorType = "type"
    ErrorTypeSystem     ErrorType = "system"
)

// 记录带错误信息的操作
collector.RecordORMOperation(
    monitoring.OperationInsert,
    "User",
    false, // 操作失败
    150*time.Millisecond,
    &monitoring.ErrorInfo{
        Type:    monitoring.ErrorTypeDatabase,
        Message: "duplicate key value violates unique constraint",
        Code:    "23505",
    },
    labels,
)
```

## 性能优化

监控系统设计为低开销：

1. **异步收集**：默认启用异步收集，减少对业务逻辑的影响
2. **采样率控制**：支持配置采样率，控制监控数据量
3. **内存优化**：使用高效的数据结构，避免内存泄漏
4. **零分配设计**：关键路径避免内存分配

## 测试和验证

```bash
# 运行监控系统测试
go test ./monitoring/... -v

# 运行端到端测试
go test ./monitoring/e2e_test.go -v

# 运行示例程序
cd monitoring/example && go run example.go
```

## 与外部系统集成

监控数据可以通过多种方式导出：

```go
// 获取原始监控数据
data := collector.GetMetrics()

// 转换为JSON格式
jsonData, _ := json.Marshal(data)

// 转换为Prometheus格式
prometheusData := collector.ToPrometheusFormat()

// 自定义导出处理器
collector.SetExportHandler(func(metrics []monitoring.Metric) {
    // 发送到外部监控系统
    sendToExternalSystem(metrics)
})
```

## 最佳实践

1. **合理使用标签**：使用有意义的标签便于数据分析和过滤
2. **控制数据量**：根据需求调整采样率和数据保留策略
3. **错误分类**：使用详细的错误类型便于问题排查
4. **性能监控**：监控监控系统本身的性能
5. **集成测试**：在生产环境前充分测试监控集成

## API 参考

### 核心接口

- `monitoring.Collector`: 通用收集器接口
- `monitoring.OperationType`: ORM 操作类型常量
- `monitoring.Metric`: 监控数据指标定义

### 收集器实现

- `orm.Collector`: ORM 操作收集器
  - `NewCollector()`: 创建新的ORM收集器
  - `RecordOperation()`: 记录ORM操作
  - `GetMetrics()`: 获取收集的指标

- `validation.Collector`: 验证操作收集器
  - `NewCollector()`: 创建新的验证收集器
  - `RecordValidation()`: 记录验证操作
  - `RecordCacheAccess()`: 记录缓存访问
  - `GetMetrics()`: 获取收集的指标

- `database.Collector`: 数据库操作收集器
  - `NewCollector()`: 创建新的数据库收集器
  - `RecordQuery()`: 记录数据库查询
  - `RecordTransaction()`: 记录数据库事务
  - `RecordConnection()`: 记录数据库连接
  - `GetMetrics()`: 获取收集的指标

### 简单收集器

- `core.SimpleCollector`: 简单的通用收集器实现
- `core.NoopCollector`: 无操作收集器（用于测试）
- `core.TestCollector`: 测试收集器（用于单元测试）

## 许可证

MagicORM 项目的一部分。详情请参阅主项目 LICENSE 文件。