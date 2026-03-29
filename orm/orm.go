package orm

import (
	"context"
	"fmt"
	"sync"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/types"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/metrics/metricsdb"
	metricsorm "github.com/muidea/magicOrm/metrics/orm"
	metricsvalidation "github.com/muidea/magicOrm/metrics/validation"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/validation"
	verrors "github.com/muidea/magicOrm/validation/errors"
	"log/slog"
)

const maxDeepLevel = 3

// Orm orm interface
type Orm interface {
	Create(entity models.Model) *cd.Error
	Drop(entity models.Model) *cd.Error
	Insert(entity models.Model) (models.Model, *cd.Error)
	Update(entity models.Model) (models.Model, *cd.Error)
	Delete(entity models.Model) (models.Model, *cd.Error)
	Query(entity models.Model) (models.Model, *cd.Error)
	Count(filter models.Filter) (int64, *cd.Error)
	BatchQuery(filter models.Filter) ([]models.Model, *cd.Error)
	BeginTransaction() *cd.Error
	CommitTransaction() *cd.Error
	RollbackTransaction() *cd.Error
	Release()
}

var (
	name2Pool                  sync.Map
	name2PoolInitializeOnce    sync.Once
	name2PoolUninitializedOnce sync.Once
	name2LastPoolStats         sync.Map

	ormMetricProvider  *metricsorm.ORMMetricProvider
	ormMetricCollector *metricsorm.ORMMetricsCollector
)

func poolStatsChanged(prev, curr database.PoolStats) bool {
	return prev.MaxOpenConnections != curr.MaxOpenConnections ||
		prev.OpenConnections != curr.OpenConnections ||
		prev.InUse != curr.InUse ||
		prev.Idle != curr.Idle ||
		prev.WaitCount != curr.WaitCount ||
		prev.WaitDuration != curr.WaitDuration
}

func logPoolStats(owner, event string, pool database.Pool, force bool) {
	if owner == "" || pool == nil {
		return
	}

	stats := pool.GetStats()
	prevVal, ok := name2LastPoolStats.Load(owner)
	prevStats, _ := prevVal.(database.PoolStats)
	shouldLog := force || !ok || poolStatsChanged(prevStats, stats) ||
		stats.InUse > 1 || stats.OpenConnections > 1 || stats.WaitCount > 0

	name2LastPoolStats.Store(owner, stats)
	if !shouldLog {
		return
	}

	slog.Info("ORM pool stats",
		"owner", owner,
		"event", event,
		"max_open", stats.MaxOpenConnections,
		"open", stats.OpenConnections,
		"in_use", stats.InUse,
		"idle", stats.Idle,
		"wait_count", stats.WaitCount,
		"wait_ms", stats.WaitDuration.Milliseconds(),
	)
}

func defaultValidationConfig(enableCaching bool) validation.ValidationConfig {
	cfg := validation.DefaultConfig()
	cfg.EnableCaching = enableCaching
	return cfg
}

// Initialize InitOrm
func Initialize() {
	name2PoolInitializeOnce.Do(func() {
		name2Pool = sync.Map{}

		// 总是创建metrics收集器，但只在GlobalManager存在时注册provider
		registerORMMetrics()
		metricsdb.RegisterDatabaseMetrics()
		metricsvalidation.RegisterValidationMetrics()
	})
}

// registerORMMetrics 注册ORM监控provider
func registerORMMetrics() {
	// 创建全局metrics收集器（无论GlobalManager是否存在都创建）
	ormMetricCollector = metricsorm.NewORMMetricsCollector()

	// 只有在GlobalManager存在时才注册provider
	if mgr := monitoring.GetGlobalManager(); mgr != nil {
		// 创建provider并传递collector
		ormMetricProvider = metricsorm.NewORMMetricProvider(ormMetricCollector)

		// 尝试注册ORMMetricProvider
		if err := monitoring.RegisterGlobalProvider(
			"magicorm_orm",
			func() types.MetricProvider {
				return ormMetricProvider
			},
			true, // 自动初始化
			100,  // 优先级
		); err != nil {
			ormMetricProvider = nil
			// 记录错误但不影响ORM初始化
			slog.Warn("Failed to register ORM metrics provider", "error", err.Error())
		} else {
			slog.Info("ORM metrics provider registered successfully")
		}
	} else {
		// GlobalManager不存在，只创建collector不注册provider
		slog.Debug("GlobalManager not available, ORM metrics collector created but provider not registered")
	}
}

// EnsureORMMetricProviderRegistered 尝试在GlobalManager可用时注册ORM metrics provider。
// 该方法是幂等的，可在monitoring.InitializeGlobalManager()之后显式调用。
func EnsureORMMetricProviderRegistered() {
	// 如果已经有provider，则不需要重复注册
	if ormMetricProvider != nil {
		return
	}
	if ormMetricCollector == nil {
		// 尚未初始化ORM，保持静默
		return
	}

	if monitoring.GetGlobalManager() == nil {
		// 监控系统尚未初始化，无法注册
		return
	}

	// 创建并注册provider
	ormMetricProvider = metricsorm.NewORMMetricProvider(ormMetricCollector)
	if err := monitoring.RegisterGlobalProvider(
		"magicorm_orm",
		func() types.MetricProvider {
			return ormMetricProvider
		},
		true,
		100,
	); err != nil {
		ormMetricProvider = nil
		slog.Warn("Failed to ensure ORM metrics provider registration", "error", err.Error())
	} else {
		slog.Info("ORM metrics provider ensured and registered successfully")
	}
}

// Uninitialized orm
func Uninitialized() {
	name2PoolUninitializedOnce.Do(func() {
		name2Pool.Range(func(key, val any) bool {
			owner, _ := key.(string)
			pool := val.(database.Pool)
			logPoolStats(owner, "uninitialize_pool", pool, true)
			pool.Uninitialized()
			name2LastPoolStats.Delete(owner)

			return true
		})

		name2Pool = sync.Map{}
	})
}

func AddDatabase(dbServer, dbName, username, password string, maxConnNum int, owner string) (err *cd.Error) {
	config := NewConfig(dbServer, dbName, username, password)

	val, ok := name2Pool.Load(owner)
	if ok {
		pool := val.(database.Pool)
		pool.IncReference()
		err = pool.CheckConfig(config)
		logPoolStats(owner, "reuse_pool", pool, true)
		return
	}

	pool := NewPool()
	err = pool.Initialize(maxConnNum, config)
	if err != nil {
		slog.Error("AddDatabase pool.Initialize failed", "owner", owner, "error", err.Error())
		return
	}

	pool.IncReference()
	name2Pool.Store(owner, pool)
	logPoolStats(owner, "initialize_pool", pool, true)
	return
}

func DelDatabase(owner string) {
	val, ok := name2Pool.Load(owner)
	if !ok {
		return
	}

	pool := val.(database.Pool)
	if pool.DecReference() == 0 {
		logPoolStats(owner, "delete_pool", pool, true)
		pool.Uninitialized()
		name2Pool.Delete(owner)
		name2LastPoolStats.Delete(owner)
	}
}

// NewOrm create new Orm
func NewOrm(provider provider.Provider, cfg database.Config, prefix string) (Orm, *cd.Error) {
	if provider == nil {
		return nil, cd.NewError(cd.IllegalParam, "provider is nil")
	}

	executorVal, executorErr := NewExecutor(cfg)
	if executorErr != nil {
		slog.Error("NewOrm NewExecutor failed", "error", executorErr.Error())
		return nil, cd.NewError(cd.Unexpected, executorErr.Error())
	}

	// Create validation manager
	validationFactory := validation.NewValidationFactory()
	// ORM handlers are request-scoped and released after each DAO call. Keeping
	// validation caching enabled here starts a cleanup goroutine per handler and
	// leaks it because Release only tears down the executor.
	validationConfig := defaultValidationConfig(false)
	validationMgr := validationFactory.CreateValidationManager(validationConfig)

	orm := &impl{
		context:         context.Background(),
		executor:        executorVal,
		modelProvider:   provider,
		modelCodec:      codec.New(provider, prefix),
		validationMgr:   validationMgr,
		validationCache: validationConfig.EnableCaching,
	}
	return orm, nil
}

// GetOrm get orm from pool
func GetOrm(ctx context.Context, provider provider.Provider, prefix string) (ret Orm, err *cd.Error) {
	if provider == nil {
		err = cd.NewError(cd.IllegalParam, "provider is nil")
		slog.Error("GetOrm: provider is nil")
		return
	}

	val, ok := name2Pool.Load(provider.Owner())
	if !ok {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("can't find orm,name:%s", provider.Owner()))
		slog.Error("GetOrm: pool not found", "owner", provider.Owner())
		return
	}

	pool := val.(database.Pool)
	executorVal, executorErr := pool.GetExecutor(ctx)
	if executorErr != nil {
		err = executorErr
		slog.Error("GetOrm pool.GetExecutor failed", "owner", provider.Owner(), "error", err.Error())
		return
	}
	logPoolStats(provider.Owner(), "acquire_executor", pool, false)

	// Create validation manager
	validationFactory := validation.NewValidationFactory()
	// ORM handlers fetched from the pool are short-lived; per-handler validation
	// caches retain cleanup goroutines and heap state beyond Release.
	validationConfig := defaultValidationConfig(false)
	validationMgr := validationFactory.CreateValidationManager(validationConfig)

	ret = &impl{
		context:         ctx,
		executor:        executorVal,
		modelProvider:   provider,
		modelCodec:      codec.New(provider, prefix),
		validationMgr:   validationMgr,
		validationCache: validationConfig.EnableCaching,
	}
	return
}

// impl orm
type impl struct {
	context         context.Context
	executor        database.Executor
	modelProvider   provider.Provider
	modelCodec      codec.Codec
	validationMgr   validation.ValidationManager
	validationCache bool
}

// BeginTransaction begin transaction
func (s *impl) BeginTransaction() (err *cd.Error) {
	if s.executor != nil {
		err = s.executor.BeginTransaction()
		if err != nil {
			slog.Error("BeginTransaction failed", "error", err.Error())
		}
	}

	// Record transaction metric
	if ormMetricCollector != nil {
		success := err == nil
		ormMetricCollector.RecordTransaction(string(metrics.OperationCreate), success)
	}

	return
}

// CommitTransaction commit transaction
func (s *impl) CommitTransaction() (err *cd.Error) {
	if s.executor != nil {
		err = s.executor.CommitTransaction()
		if err != nil {
			slog.Error("CommitTransaction failed", "error", err.Error())
		}
	}

	// Record transaction metric
	if ormMetricCollector != nil {
		success := err == nil
		ormMetricCollector.RecordTransaction("commit", success)
	}

	return
}

// RollbackTransaction rollback transaction
func (s *impl) RollbackTransaction() (err *cd.Error) {
	if s.executor != nil {
		err = s.executor.RollbackTransaction()
		if err != nil {
			slog.Error("RollbackTransaction failed", "error", err.Error())
		}
	}

	// Record transaction metric
	if ormMetricCollector != nil {
		success := err == nil
		ormMetricCollector.RecordTransaction("rollback", success)
	}

	return
}

func (s *impl) finalTransaction(err *cd.Error) {
	if err == nil {
		err = s.executor.CommitTransaction()
		if err != nil {
			slog.Error("finalTransaction Commit failed", "error", err.Error())
		}
		return
	}

	err = s.executor.RollbackTransaction()
	if err != nil {
		slog.Error("finalTransaction Rollback failed", "error", err.Error())
	}
}

func (s *impl) Release() {
	if s.executor != nil {
		s.executor.Release()
		s.executor = nil
	}
}

// Validation methods

// validateModel validates a model with scenario-aware validation
func (s *impl) validateModel(model models.Model, scenario verrors.Scenario) *cd.Error {
	if s.validationMgr == nil {
		return nil
	}

	// Create validation context without database type
	// Database-specific validation will be handled by the validation system internally
	ctx := validation.NewContext(
		scenario,
		s.getOperationType(scenario),
		nil, // Model adapter would be created from model
		"",  // Database type not specified here
	)

	// Perform validation
	err := s.validationMgr.ValidateModel(model, ctx)
	if err != nil {
		return cd.NewError(cd.IllegalParam, err.Error())
	}

	return nil
}

// getOperationType maps scenario to operation type
func (s *impl) getOperationType(scenario verrors.Scenario) validation.OperationType {
	switch scenario {
	case verrors.ScenarioInsert:
		return validation.OperationCreate
	case verrors.ScenarioUpdate:
		return validation.OperationUpdate
	case verrors.ScenarioQuery:
		return validation.OperationRead
	case verrors.ScenarioDelete:
		return validation.OperationDelete
	default:
		return validation.OperationCreate
	}
}

// GetValidationManager returns the validation manager
func (s *impl) GetValidationManager() validation.ValidationManager {
	return s.validationMgr
}

// ConfigureValidation configures validation settings
func (s *impl) ConfigureValidation(config validation.ValidationConfig) error {
	// Recreate validation manager with new configuration
	factory := validation.NewValidationFactory()
	s.validationMgr = factory.CreateValidationManager(config)
	s.validationCache = config.EnableCaching
	return nil
}
