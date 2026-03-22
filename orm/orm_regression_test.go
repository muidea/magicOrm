package orm

import "testing"

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
