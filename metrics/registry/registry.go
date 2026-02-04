// Package registry provides a simple provider registration mechanism for MagicORM metrics.
// This package handles registration of MagicORM metric providers with magicCommon/monitoring.
package registry

import (
	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/types"
)

// Register registers a provider with the monitoring system.
// This is a simplified version that directly registers with the global manager.
func Register(name string, factory types.ProviderFactory, autoInitialize bool, priority int) error {
	return monitoring.RegisterGlobalProvider(name, factory, autoInitialize, priority)
}
