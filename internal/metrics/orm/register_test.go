package orm

import "testing"

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
