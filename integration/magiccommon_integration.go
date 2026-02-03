// 简单的 magicCommon/monitoring 集成
package integration

import (
	"fmt"
	"log"
	"sync"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/core"
	"github.com/muidea/magicCommon/monitoring/types"
)

// 全局注册表避免重复创建
var (
	integrationRegistry = make(map[string]*SimpleIntegration)
	registryMutex       sync.RWMutex
)

// SimpleIntegration 提供简单的 magicCommon 监控集成
type SimpleIntegration struct {
	manager   *monitoring.Manager
	collector *core.Collector
	namespace string
}

// NewSimpleIntegration 创建新的简单集成
func NewSimpleIntegration(config *core.MonitoringConfig) (*SimpleIntegration, *cd.Error) {
	if config == nil {
		defaultConfig := core.DefaultMonitoringConfig()
		defaultConfig.Namespace = "magicorm"
		config = &defaultConfig
	}

	// 检查是否已存在相同命名空间的集成
	registryMutex.RLock()
	if existing, exists := integrationRegistry[config.Namespace]; exists {
		registryMutex.RUnlock()
		return existing, nil
	}
	registryMutex.RUnlock()

	// 创建 magicCommon 监控管理器
	manager, err := monitoring.NewManager(config)
	if err != nil {
		return nil, err
	}

	// 初始化管理器
	if err := manager.Initialize(); err != nil {
		return nil, err
	}

	// 启动监控系统
	if err := manager.Start(); err != nil {
		return nil, err
	}

	integration := &SimpleIntegration{
		manager:   manager,
		collector: manager.GetCollector(),
		namespace: config.Namespace,
	}

	// 注册 magicOrm 监控提供者
	if err := integration.registerProviders(); err != nil {
		manager.Stop()
		return nil, err
	}

	// 注册到全局注册表
	registryMutex.Lock()
	integrationRegistry[config.Namespace] = integration
	registryMutex.Unlock()

	return integration, nil
}

// registerProviders 注册 magicOrm 监控提供者
func (i *SimpleIntegration) registerProviders() *cd.Error {
	// 检查是否已注册
	providers := i.collector.GetProviders()

	// 注册 ORM 监控提供者
	if _, exists := providers["magicorm_orm"]; !exists {
		ormProvider := NewSimpleORMProvider()
		if err := i.collector.RegisterProvider(ormProvider); err != nil {
			return err
		}
	}

	// 注册验证监控提供者
	if _, exists := providers["magicorm_validation"]; !exists {
		valProvider := NewSimpleValidationProvider()
		if err := i.collector.RegisterProvider(valProvider); err != nil {
			return err
		}
	}

	// 注册数据库监控提供者
	if _, exists := providers["magicorm_database"]; !exists {
		dbProvider := NewSimpleDatabaseProvider()
		if err := i.collector.RegisterProvider(dbProvider); err != nil {
			return err
		}
	}

	return nil
}

// RecordORMOperation 记录 ORM 操作
func (i *SimpleIntegration) RecordORMOperation(
	operation string,
	modelName string,
	success bool,
	latency time.Duration,
	err error,
	labels map[string]string,
) {
	// 转换标签
	commonLabels := convertLabels(labels)
	commonLabels["model"] = modelName
	commonLabels["operation"] = operation
	commonLabels["success"] = fmt.Sprintf("%v", success)

	// 记录操作计数
	metricName := "magicorm_operations_total"
	if err := i.collector.Increment(metricName, commonLabels); err != nil {
		log.Printf("记录 ORM 操作失败: %v", err)
	}

	// 记录延迟
	latencyMetricName := "magicorm_operation_duration_seconds"
	if err := i.collector.Record(latencyMetricName, latency.Seconds(), commonLabels); err != nil {
		log.Printf("记录 ORM 延迟失败: %v", err)
	}

	// 记录错误（如果有）
	if err != nil {
		errorLabels := make(map[string]string)
		for k, v := range commonLabels {
			errorLabels[k] = v
		}
		errorLabels["error_type"] = "database"
		errorLabels["error_code"] = "UNKNOWN"

		errorMetricName := "magicorm_errors_total"
		if err := i.collector.Increment(errorMetricName, errorLabels); err != nil {
			log.Printf("记录 ORM 错误失败: %v", err)
		}
	}
}

// RecordValidationOperation 记录验证操作
func (i *SimpleIntegration) RecordValidationOperation(
	validatorName string,
	modelName string,
	scenario string,
	latency time.Duration,
	err error,
	labels map[string]string,
) {
	// 转换标签
	commonLabels := convertLabels(labels)
	commonLabels["validator"] = validatorName
	commonLabels["model"] = modelName
	commonLabels["scenario"] = scenario

	// 记录验证计数
	metricName := "magicorm_validation_operations_total"
	if err := i.collector.Increment(metricName, commonLabels); err != nil {
		log.Printf("记录验证操作失败: %v", err)
	}

	// 记录验证延迟
	latencyMetricName := "magicorm_validation_duration_seconds"
	if err := i.collector.Record(latencyMetricName, latency.Seconds(), commonLabels); err != nil {
		log.Printf("记录验证延迟失败: %v", err)
	}

	// 记录验证错误（如果有）
	if err != nil {
		errorLabels := make(map[string]string)
		for k, v := range commonLabels {
			errorLabels[k] = v
		}
		errorLabels["error_type"] = "validation"
		errorLabels["error_code"] = "UNKNOWN"

		errorMetricName := "magicorm_validation_errors_total"
		if err := i.collector.Increment(errorMetricName, errorLabels); err != nil {
			log.Printf("记录验证错误失败: %v", err)
		}
	}
}

// RecordDatabaseOperation 记录数据库操作
func (i *SimpleIntegration) RecordDatabaseOperation(
	dbType string,
	queryType string,
	success bool,
	latency time.Duration,
	rowsAffected int,
	err error,
	labels map[string]string,
) {
	// 转换标签
	commonLabels := convertLabels(labels)
	commonLabels["database"] = dbType
	commonLabels["query_type"] = queryType
	commonLabels["success"] = fmt.Sprintf("%v", success)

	// 记录查询计数
	metricName := "magicorm_database_queries_total"
	if err := i.collector.Increment(metricName, commonLabels); err != nil {
		log.Printf("记录数据库查询失败: %v", err)
	}

	// 记录查询延迟
	latencyMetricName := "magicorm_database_query_duration_seconds"
	if err := i.collector.Record(latencyMetricName, latency.Seconds(), commonLabels); err != nil {
		log.Printf("记录数据库延迟失败: %v", err)
	}

	// 记录影响行数
	if rowsAffected > 0 {
		rowsMetricName := "magicorm_database_rows_affected"
		rowsLabels := make(map[string]string)
		for k, v := range commonLabels {
			rowsLabels[k] = v
		}
		if err := i.collector.Record(rowsMetricName, float64(rowsAffected), rowsLabels); err != nil {
			log.Printf("记录影响行数失败: %v", err)
		}
	}

	// 记录数据库错误（如果有）
	if err != nil {
		errorLabels := make(map[string]string)
		for k, v := range commonLabels {
			errorLabels[k] = v
		}
		errorLabels["error_type"] = "database"
		errorLabels["error_code"] = "UNKNOWN"

		errorMetricName := "magicorm_database_errors_total"
		if err := i.collector.Increment(errorMetricName, errorLabels); err != nil {
			log.Printf("记录数据库错误失败: %v", err)
		}
	}
}

// GetManager 返回 magicCommon 监控管理器
func (i *SimpleIntegration) GetManager() *monitoring.Manager {
	return i.manager
}

// Stop 停止集成
func (i *SimpleIntegration) Stop() *cd.Error {
	// 从注册表中移除
	registryMutex.Lock()
	delete(integrationRegistry, i.namespace)
	registryMutex.Unlock()

	return i.manager.Stop()
}

// SimpleORMProvider 简单的 ORM 监控提供者
type SimpleORMProvider struct {
	*types.BaseProvider
}

// NewSimpleORMProvider 创建新的简单 ORM 监控提供者
func NewSimpleORMProvider() *SimpleORMProvider {
	base := types.NewBaseProvider(
		"magicorm_orm",
		"1.0.0",
		"MagicORM ORM operation monitoring provider",
	)
	base.AddTag("orm")
	base.AddTag("magicorm")

	return &SimpleORMProvider{
		BaseProvider: base,
	}
}

// Init 初始化提供者
func (p *SimpleORMProvider) Init(collector interface{}) *cd.Error {
	// 类型断言获取 core.Collector
	c, ok := collector.(*core.Collector)
	if !ok {
		return cd.NewError(cd.IllegalParam, "invalid collector type")
	}

	// 注册 ORM 相关指标定义
	definitions := []types.MetricDefinition{
		types.NewCounterDefinition(
			"magicorm_operations_total",
			"Total number of ORM operations",
			[]string{"model", "operation", "success"},
			nil,
		),
		types.NewGaugeDefinition(
			"magicorm_operation_duration_seconds",
			"ORM operation duration in seconds",
			[]string{"model", "operation", "success"},
			nil,
		),
		types.NewCounterDefinition(
			"magicorm_errors_total",
			"Total number of ORM errors",
			[]string{"model", "operation", "error_type", "error_code"},
			nil,
		),
	}

	for _, def := range definitions {
		if err := c.RegisterDefinition(def); err != nil {
			return err
		}
	}

	p.UpdateHealthStatus(types.ProviderHealthy)
	return nil
}

// Metrics 返回指标定义
func (p *SimpleORMProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		types.NewCounterDefinition(
			"magicorm_operations_total",
			"Total number of ORM operations",
			[]string{"model", "operation", "success"},
			nil,
		),
		types.NewGaugeDefinition(
			"magicorm_operation_duration_seconds",
			"ORM operation duration in seconds",
			[]string{"model", "operation", "success"},
			nil,
		),
		types.NewCounterDefinition(
			"magicorm_errors_total",
			"Total number of ORM errors",
			[]string{"model", "operation", "error_type", "error_code"},
			nil,
		),
	}
}

// Collect 收集指标
func (p *SimpleORMProvider) Collect() ([]types.Metric, *cd.Error) {
	p.UpdateCollectionStats(true, 0, 0)
	return []types.Metric{}, nil
}

// SimpleValidationProvider 简单的验证监控提供者
type SimpleValidationProvider struct {
	*types.BaseProvider
}

// NewSimpleValidationProvider 创建新的简单验证监控提供者
func NewSimpleValidationProvider() *SimpleValidationProvider {
	base := types.NewBaseProvider(
		"magicorm_validation",
		"1.0.0",
		"MagicORM validation monitoring provider",
	)
	base.AddTag("validation")
	base.AddTag("magicorm")

	return &SimpleValidationProvider{
		BaseProvider: base,
	}
}

// Init 初始化提供者
func (p *SimpleValidationProvider) Init(collector interface{}) *cd.Error {
	// 类型断言获取 core.Collector
	c, ok := collector.(*core.Collector)
	if !ok {
		return cd.NewError(cd.IllegalParam, "invalid collector type")
	}

	// 注册验证相关指标定义
	definitions := []types.MetricDefinition{
		types.NewCounterDefinition(
			"magicorm_validation_operations_total",
			"Total number of validation operations",
			[]string{"validator", "model", "scenario"},
			nil,
		),
		types.NewGaugeDefinition(
			"magicorm_validation_duration_seconds",
			"Validation operation duration in seconds",
			[]string{"validator", "model", "scenario"},
			nil,
		),
		types.NewCounterDefinition(
			"magicorm_validation_errors_total",
			"Total number of validation errors",
			[]string{"validator", "model", "scenario", "error_type", "error_code"},
			nil,
		),
	}

	for _, def := range definitions {
		if err := c.RegisterDefinition(def); err != nil {
			return err
		}
	}

	p.UpdateHealthStatus(types.ProviderHealthy)
	return nil
}

// Metrics 返回指标定义
func (p *SimpleValidationProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		types.NewCounterDefinition(
			"magicorm_validation_operations_total",
			"Total number of validation operations",
			[]string{"validator", "model", "scenario"},
			nil,
		),
		types.NewGaugeDefinition(
			"magicorm_validation_duration_seconds",
			"Validation operation duration in seconds",
			[]string{"validator", "model", "scenario"},
			nil,
		),
		types.NewCounterDefinition(
			"magicorm_validation_errors_total",
			"Total number of validation errors",
			[]string{"validator", "model", "scenario", "error_type", "error_code"},
			nil,
		),
	}
}

// Collect 收集指标
func (p *SimpleValidationProvider) Collect() ([]types.Metric, *cd.Error) {
	p.UpdateCollectionStats(true, 0, 0)
	return []types.Metric{}, nil
}

// SimpleDatabaseProvider 简单的数据库监控提供者
type SimpleDatabaseProvider struct {
	*types.BaseProvider
}

// NewSimpleDatabaseProvider 创建新的简单数据库监控提供者
func NewSimpleDatabaseProvider() *SimpleDatabaseProvider {
	base := types.NewBaseProvider(
		"magicorm_database",
		"1.0.0",
		"MagicORM database monitoring provider",
	)
	base.AddTag("database")
	base.AddTag("magicorm")

	return &SimpleDatabaseProvider{
		BaseProvider: base,
	}
}

// Init 初始化提供者
func (p *SimpleDatabaseProvider) Init(collector interface{}) *cd.Error {
	// 类型断言获取 core.Collector
	c, ok := collector.(*core.Collector)
	if !ok {
		return cd.NewError(cd.IllegalParam, "invalid collector type")
	}

	// 注册数据库相关指标定义
	definitions := []types.MetricDefinition{
		types.NewCounterDefinition(
			"magicorm_database_queries_total",
			"Total number of database queries",
			[]string{"database", "query_type", "success"},
			nil,
		),
		types.NewGaugeDefinition(
			"magicorm_database_query_duration_seconds",
			"Database query duration in seconds",
			[]string{"database", "query_type", "success"},
			nil,
		),
		types.NewGaugeDefinition(
			"magicorm_database_rows_affected",
			"Number of rows affected by database operations",
			[]string{"database", "query_type", "success"},
			nil,
		),
		types.NewCounterDefinition(
			"magicorm_database_errors_total",
			"Total number of database errors",
			[]string{"database", "query_type", "error_type", "error_code"},
			nil,
		),
	}

	for _, def := range definitions {
		if err := c.RegisterDefinition(def); err != nil {
			return err
		}
	}

	p.UpdateHealthStatus(types.ProviderHealthy)
	return nil
}

// Metrics 返回指标定义
func (p *SimpleDatabaseProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		types.NewCounterDefinition(
			"magicorm_database_queries_total",
			"Total number of database queries",
			[]string{"database", "query_type", "success"},
			nil,
		),
		types.NewGaugeDefinition(
			"magicorm_database_query_duration_seconds",
			"Database query duration in seconds",
			[]string{"database", "query_type", "success"},
			nil,
		),
		types.NewGaugeDefinition(
			"magicorm_database_rows_affected",
			"Number of rows affected by database operations",
			[]string{"database", "query_type", "success"},
			nil,
		),
		types.NewCounterDefinition(
			"magicorm_database_errors_total",
			"Total number of database errors",
			[]string{"database", "query_type", "error_type", "error_code"},
			nil,
		),
	}
}

// Collect 收集指标
func (p *SimpleDatabaseProvider) Collect() ([]types.Metric, *cd.Error) {
	p.UpdateCollectionStats(true, 0, 0)
	return []types.Metric{}, nil
}

// convertLabels 转换标签格式
func convertLabels(labels map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range labels {
		// 确保标签键符合 Prometheus 规范
		safeKey := sanitizeLabelKey(k)
		result[safeKey] = v
	}
	return result
}

// sanitizeLabelKey 清理标签键，确保符合 Prometheus 规范
func sanitizeLabelKey(key string) string {
	// 简单的清理逻辑
	// 在实际应用中可能需要更复杂的处理
	return key
}

// 使用示例
func ExampleSimpleIntegration() {
	// 创建集成
	config := core.DefaultMonitoringConfig()
	config.Namespace = "magicorm"
	config.AsyncCollection = true
	config.CollectionInterval = 30 * time.Second

	integration, err := NewSimpleIntegration(&config)
	if err != nil {
		log.Printf("创建集成失败: %v", err)
		return
	}
	defer integration.Stop()

	// 记录示例操作
	integration.RecordORMOperation(
		"insert",
		"User",
		true,
		150*time.Millisecond,
		nil,
		map[string]string{"test": "integration"},
	)

	integration.RecordValidationOperation(
		"validate_user",
		"User",
		"insert",
		50*time.Millisecond,
		nil,
		map[string]string{"test": "integration"},
	)

	integration.RecordDatabaseOperation(
		"postgresql",
		"select",
		true,
		200*time.Millisecond,
		10,
		nil,
		map[string]string{"test": "integration"},
	)

	// 获取管理器统计信息
	manager := integration.GetManager()
	stats := manager.GetStats()
	log.Printf("管理器统计: %+v", stats)

	log.Println("集成示例完成")
}
