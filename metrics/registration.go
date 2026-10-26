// Package metrics integrates MagicORM's runtime metrics with magicCommon/monitoring.
// Collectors, metric keys and sampling state are internal implementation details.
package metrics

import (
	"github.com/muidea/magicOrm/internal/metrics/metricsdb"
	metricsorm "github.com/muidea/magicOrm/internal/metrics/orm"
	metricsvalidation "github.com/muidea/magicOrm/internal/metrics/validation"
)

// Initialize enables ORM, database and validation metric collection. It may be
// used independently of the ORM and preserves collectors initialized earlier.
// Providers are registered immediately when monitoring is already available.
func Initialize() {
	if metricsorm.GetORMMetricsCollector() == nil {
		metricsorm.RegisterORMMetrics(metricsorm.NewORMMetricsCollector())
	}
	if metricsdb.GetDatabaseMetricsCollector() == nil {
		metricsdb.RegisterDatabaseMetrics()
	}
	if metricsvalidation.GetValidationMetricsCollector() == nil {
		metricsvalidation.RegisterValidationMetrics()
	}
	EnsureRegistered()
}

// EnsureRegistered registers initialized ORM, database and validation providers
// when the global monitoring manager becomes available. It does not reset any
// collector and is safe to repeat. Call Initialize or orm.Initialize to enable collection first.
func EnsureRegistered() {
	metricsorm.EnsureORMMetricProviderRegistered()
	metricsdb.EnsureDatabaseMetricProviderRegistered()
	metricsvalidation.EnsureValidationMetricProviderRegistered()
}
