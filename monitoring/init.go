// Package monitoring provides initialization for MagicORM monitoring.
// This is a standalone initialization that doesn't create import cycles.
package monitoring

import (
	"github.com/muidea/magicCommon/monitoring"
)

// InitializeWithManager initializes MagicORM monitoring with an external monitoring manager.
// This is a simplified version that doesn't create import cycles.
func InitializeWithManager(manager *monitoring.Manager) error {
	if manager == nil {
		return nil
	}

	// Note: In the actual implementation, this would register providers.
	// For now, this is a placeholder that doesn't create import cycles.
	return nil
}

// GetSimpleCollector returns a simple collector for testing.
func GetSimpleCollector() Collector {
	return NewSimpleCollector()
}
