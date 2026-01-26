package models

type Key string

/*
| 指令 | 参数 | 描述 | 业务应用建议 |
| :--- | :--- | :--- | :--- |
| **`req`** | 无 | **Required**: 必填/必传 | 校验值不能为零值（0, "", nil）。 |
| **`ro`** | 无 | **Read-Only**: 只读 | 输出接口展示，但更新接口忽略此字段。 |
| **`wo`** | 无 | **Write-Only**: 只写 | 敏感字段（如密码），禁止在展示接口输出。 |
| **`imm`** | 无 | **Immutable**: 不可变 | 仅允许在 Create 时赋值，Update 时视为只读。 |
| **`opt`** | 无 | **Optional**: 可选 | 若字段为空，则跳过后续所有校验指令。 |
*/
// 预定义核心指令（内置部分）
const (
	KeyRequired  Key = "req"
	KeyReadOnly  Key = "ro"
	KeyWriteOnly Key = "wo"
	KeyImmutable Key = "imm"
	KeyMin       Key = "min"
	KeyMax       Key = "max"
	KeyIn        Key = "in"
)

// Directive 表达单个约束及其参数
type Directive interface {
	Key() Key
	Args() []string
	HasArgs() bool
}

type Constraints interface {
	Has(key Key) bool
	Get(key Key) (Directive, bool)
}
