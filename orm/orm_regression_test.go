package orm

import (
	"reflect"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

func TestDefaultValidationConfigDisablesCachingForTransientORM(t *testing.T) {
	cfg := defaultValidationConfig(false)
	if cfg.EnableCaching {
		t.Fatal("expected transient ORM validation caching to be disabled")
	}
}

func TestEnsureORMMetricProviderRegisteredWithoutCollector(t *testing.T) {
	originalProvider := ormMetricProvider
	originalCollector := ormMetricCollector
	t.Cleanup(func() {
		ormMetricProvider = originalProvider
		ormMetricCollector = originalCollector
	})

	ormMetricProvider = nil
	ormMetricCollector = nil

	EnsureORMMetricProviderRegistered()

	if ormMetricProvider != nil {
		t.Fatal("expected ensure registration to stay silent without collector")
	}
}

func TestAcquireValidationManagerReturnsUsableManager(t *testing.T) {
	manager := acquireValidationManager()
	if manager == nil {
		t.Fatal("expected pooled validation manager")
	}

	releaseValidationManager(manager)
}

func TestConfigureValidationDropsPooledManager(t *testing.T) {
	manager := acquireValidationManager()
	ormImpl := &impl{
		validationMgr:       manager,
		pooledValidationMgr: true,
	}

	cfg := defaultValidationConfig(true)
	if err := ormImpl.ConfigureValidation(cfg); err != nil {
		t.Fatalf("ConfigureValidation failed: %v", err)
	}

	if ormImpl.validationMgr == nil {
		t.Fatal("expected ConfigureValidation to install a validation manager")
	}
	if ormImpl.pooledValidationMgr {
		t.Fatal("expected ConfigureValidation to detach from pooled validation manager")
	}
	if !ormImpl.validationCache {
		t.Fatal("expected ConfigureValidation to honor config.EnableCaching")
	}
}

func TestOrmPublicQueryContract(t *testing.T) {
	ormType := reflect.TypeOf((*Orm)(nil)).Elem()

	if _, exists := ormType.MethodByName("QueryByFilter"); exists {
		t.Fatal("QueryByFilter must not be part of the public Orm interface")
	}

	queryMethod, ok := ormType.MethodByName("Query")
	if !ok {
		t.Fatal("Query must remain part of the public Orm interface")
	}
	if queryMethod.Type.NumIn() != 1 || queryMethod.Type.In(0) != reflect.TypeOf((*models.Model)(nil)).Elem() {
		t.Fatalf("unexpected Query input signature: %s", queryMethod.Type.String())
	}
	if queryMethod.Type.NumOut() != 2 ||
		queryMethod.Type.Out(0) != reflect.TypeOf((*models.Model)(nil)).Elem() ||
		queryMethod.Type.Out(1) != reflect.TypeOf((*cd.Error)(nil)) {
		t.Fatalf("unexpected Query output signature: %s", queryMethod.Type.String())
	}

	batchMethod, ok := ormType.MethodByName("BatchQuery")
	if !ok {
		t.Fatal("BatchQuery must remain part of the public Orm interface")
	}
	if batchMethod.Type.NumIn() != 1 || batchMethod.Type.In(0) != reflect.TypeOf((*models.Filter)(nil)).Elem() {
		t.Fatalf("unexpected BatchQuery input signature: %s", batchMethod.Type.String())
	}
	if batchMethod.Type.NumOut() != 2 ||
		batchMethod.Type.Out(0) != reflect.TypeOf([]models.Model{}) ||
		batchMethod.Type.Out(1) != reflect.TypeOf((*cd.Error)(nil)) {
		t.Fatalf("unexpected BatchQuery output signature: %s", batchMethod.Type.String())
	}
}
