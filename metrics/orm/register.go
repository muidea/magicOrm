// Package orm provides ORM-specific metric definitions for MagicORM.
package orm

import (
	"github.com/muidea/magicCommon/monitoring/types"
	"github.com/muidea/magicOrm/metrics/registry"
)

func init() {
	// 自动注册到 MagicORM metrics 系统（延迟注册到 GlobalManager）
	registry.Register(
		"magicorm_orm",
		func() types.MetricProvider {
			return NewORMMetricProvider()
		},
		true, // 自动初始化
		100,  // 优先级
	)
}
