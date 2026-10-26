package metrics_test

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/core"
	"github.com/muidea/magicOrm/internal/metrics/metricsdb"
	metricsvalidation "github.com/muidea/magicOrm/internal/metrics/validation"
	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/orm"
)

func TestLateRegistrationPreservesAllRuntimeProviders(t *testing.T) {
	metrics.EnsureRegistered() // Before either subsystem starts, stay idle.
	metrics.Initialize()       // Standalone validation/DB collection does not need an ORM handler.
	database := metricsdb.GetDatabaseMetricsCollector()
	validation := metricsvalidation.GetValidationMetricsCollector()
	database.RecordQuery("postgresql", "select", time.Millisecond, nil)
	validation.RecordValidation("create", "Sample", "insert", time.Millisecond, nil)
	metrics.Initialize() // Repeated initialization must retain standalone samples.
	orm.Initialize()     // Joining the ORM must retain the same standalone collectors.
	t.Cleanup(orm.Uninitialized)
	metrics.EnsureRegistered() // A missing manager must not reset collectors.

	config := core.DevelopmentConfig() // No network exporter or background collection.
	if err := monitoring.InitializeGlobalManagerWithConfig(&config); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = monitoring.ShutdownGlobalManager() })
	metrics.EnsureRegistered()
	registry := monitoring.GetGlobalManager().GetRegistry()
	for _, name := range []string{"magicorm_orm", "magicorm_database", "magicorm_validation"} {
		first, err := registry.GetProvider(name)
		if err != nil {
			t.Fatal(err)
		}
		metrics.EnsureRegistered()
		second, err := registry.GetProvider(name)
		if err != nil || first != second {
			t.Fatalf("repeated registration replaced %s: %v", name, err)
		}
	}
	if database != metricsdb.GetDatabaseMetricsCollector() || validation != metricsvalidation.GetValidationMetricsCollector() {
		t.Fatal("late registration replaced runtime collectors")
	}
	for _, name := range []string{"magicorm_database", "magicorm_validation"} {
		provider, err := registry.GetProvider(name)
		if err != nil {
			t.Fatal(err)
		}
		values, err := provider.Collect()
		if err != nil || len(values) == 0 {
			t.Fatalf("runtime samples lost for %s: %v", name, err)
		}
	}
}
