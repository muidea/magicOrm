package orm

import (
	"context"
	"fmt"
	"sync"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	metricsorm "github.com/muidea/magicOrm/internal/metrics/orm"
	metrics "github.com/muidea/magicOrm/internal/metrics/shared"
	metricsapi "github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/validation"
	verrors "github.com/muidea/magicOrm/validation/errors"
)

const maxDeepLevel = 3

// Orm retains the public entrypoint name; the stable model contract is defined
// in models and does not depend on database drivers.
type Orm = models.ORM

var (
	name2Pool                  sync.Map
	name2PoolInitializeOnce    sync.Once
	name2PoolUninitializedOnce sync.Once
	name2LastPoolStats         sync.Map
	validationManagerPool      sync.Pool
	ownerPoolLocks             [64]sync.Mutex

	ormMetricCollector *metricsorm.ORMMetricsCollector
)

func poolStatsChanged(prev, curr database.PoolStats) bool {
	return prev.MaxOpenConnections != curr.MaxOpenConnections ||
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
	shouldLog := force || !ok || poolStatsChanged(prevStats, stats) || stats.WaitCount > 0

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

func newDefaultValidationManager() validation.ValidationManager {
	validationFactory := validation.NewValidationFactory()
	return validationFactory.CreateValidationManager(defaultValidationConfig(false))
}

func acquireValidationManager() validation.ValidationManager {
	if val := validationManagerPool.Get(); val != nil {
		manager := val.(validation.ValidationManager)
		manager.ResetStats()
		return manager
	}

	return newDefaultValidationManager()
}

func releaseValidationManager(manager validation.ValidationManager) {
	if manager == nil {
		return
	}

	if resettable, ok := manager.(interface{ Reset() }); ok {
		resettable.Reset()
	} else {
		return
	}
	validationManagerPool.Put(manager)
}

// Initialize InitOrm
func Initialize() {
	name2PoolInitializeOnce.Do(func() {
		name2Pool = sync.Map{}

		// 总是创建metrics收集器，但只在GlobalManager存在时注册provider
		registerORMMetrics()
	})
}

func registerORMMetrics() {
	metricsapi.Initialize()
	ormMetricCollector = metricsorm.GetORMMetricsCollector()
}

// EnsureORMMetricProviderRegistered registers the ORM provider after monitoring startup.
// Use metrics.EnsureRegistered to register all three metric providers together.
func EnsureORMMetricProviderRegistered() {
	metricsorm.EnsureORMMetricProviderRegistered()
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

func ownerPoolLock(owner string) *sync.Mutex {
	var hash uint64
	for _, value := range owner {
		hash = hash*131 + uint64(value)
	}
	return &ownerPoolLocks[hash%uint64(len(ownerPoolLocks))]
}

func AddDatabase(dbServer, dbName, schemaName, username, password string, maxConnNum int, owner string) (err *cd.Error) {
	lock := ownerPoolLock(owner)
	lock.Lock()
	defer lock.Unlock()
	config := newConfig(dbServer, dbName, schemaName, username, password)

	val, ok := name2Pool.Load(owner)
	if ok {
		pool := val.(database.Pool)
		err = pool.CheckConfig(config)
		if err != nil {
			return err
		}
		pool.IncReference()
		logPoolStats(owner, "reuse_pool", pool, true)
		return
	}

	pool := newPool()
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
	lock := ownerPoolLock(owner)
	lock.Lock()
	defer lock.Unlock()
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
func NewOrm(provider provider.Provider, cfg models.DatabaseConfig, prefix string) (Orm, *cd.Error) {
	if provider == nil {
		return nil, cd.NewError(cd.IllegalParam, "provider is nil")
	}
	if cfg == nil {
		return nil, cd.NewError(cd.IllegalParam, "database configuration is nil")
	}

	executorVal, executorErr := newExecutor(cfg)
	if executorErr != nil {
		slog.Error("NewOrm newExecutor failed", "error", executorErr.Error())
		return nil, cd.NewError(cd.Unexpected, executorErr.Error())
	}

	orm := &impl{
		context:             context.Background(),
		executor:            executorVal,
		modelProvider:       provider,
		modelCodec:          codec.New(provider, prefix),
		validationMgr:       acquireValidationManager(),
		validationCache:     false,
		pooledValidationMgr: true,
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

	ret = &impl{
		context:             ctx,
		executor:            executorVal,
		modelProvider:       provider,
		modelCodec:          codec.New(provider, prefix),
		validationMgr:       acquireValidationManager(),
		validationCache:     false,
		pooledValidationMgr: true,
	}
	return
}

// impl orm
type impl struct {
	context             context.Context
	executor            database.Executor
	modelProvider       provider.Provider
	modelCodec          codec.Codec
	validationMgr       validation.ValidationManager
	validationCache     bool
	pooledValidationMgr bool
}

// BeginTransaction begin transaction
func (s *impl) BeginTransaction() (err *cd.Error) {
	if s == nil || s.executor == nil {
		return cd.NewError(cd.InvalidOperation, "transaction requires an available ORM executor")
	}
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
	if s == nil || s.executor == nil {
		return cd.NewError(cd.InvalidOperation, "transaction requires an available ORM executor")
	}
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
	if s == nil || s.executor == nil {
		return cd.NewError(cd.InvalidOperation, "transaction requires an available ORM executor")
	}
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

func (s *impl) finalTransaction(operationErr *cd.Error) *cd.Error {
	if operationErr == nil {
		if err := s.executor.CommitTransaction(); err != nil {
			slog.Error("finalTransaction Commit failed", "error", err.Error())
			return err
		}
		return nil
	}
	if err := s.executor.RollbackTransaction(); err != nil {
		slog.Error("finalTransaction Rollback failed", "error", err.Error())
		return err
	}
	return nil
}

func (s *impl) Release() {
	if s.pooledValidationMgr && s.validationMgr != nil {
		releaseValidationManager(s.validationMgr)
		s.validationMgr = nil
		s.pooledValidationMgr = false
	}

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

// configureValidation configures validation settings
func (s *impl) configureValidation(config validation.ValidationConfig) error {
	if s.pooledValidationMgr && s.validationMgr != nil {
		releaseValidationManager(s.validationMgr)
		s.validationMgr = nil
		s.pooledValidationMgr = false
	}

	// Recreate validation manager with new configuration
	factory := validation.NewValidationFactory()
	s.validationMgr = factory.CreateValidationManager(config)
	s.validationCache = config.EnableCaching
	return nil
}

// OwnerPoolStats returns a read-only pool snapshot, never a database handle.
func OwnerPoolStats(owner string) (models.PoolStats, bool) {
	value, ok := name2Pool.Load(owner)
	if !ok {
		return models.PoolStats{}, false
	}
	return value.(database.Pool).GetStats(), true
}
