// Package metrics provides MagicORM-specific metric providers.
// Providers are automatically registered when subpackages are imported.
//
// Usage:
//
//	import _ "github.com/muidea/magicOrm/metrics" // Auto-registers metric providers
//
//	// Then use magicCommon/monitoring to manage and collect metrics
package metrics

import (
	// Import subpackages to trigger metric provider auto-registration
	_ "github.com/muidea/magicOrm/metrics/database"
	_ "github.com/muidea/magicOrm/metrics/orm"
	_ "github.com/muidea/magicOrm/metrics/validation"
)
