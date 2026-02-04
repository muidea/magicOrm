// Package validation provides validation-specific metric definitions for MagicORM.
package validation

// 移除init()函数中的自动注册
// func init() {
//     // 自动注册到 MagicORM metrics 系统（延迟注册到 GlobalManager）
//     monitoring.RegisterGlobalProvider(
//         "magicorm_validation",
//         func() types.MetricProvider {
//             return NewValidationMetricProvider()
//         },
//         true, // 自动初始化
//         100,  // 优先级
//     )
// }
