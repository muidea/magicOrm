package models

type Key string

/*
| 指令 (Key) | 常量定义 | 业务语义 | 校验逻辑描述 |
| :--- | :--- | :--- | :--- |
| **`req`** | `KeyRequired` | **必填** | 值不能为类型的零值（如 0, "", nil）。 |
| **`ro`** | `KeyReadOnly` | **只读** | 该字段仅用于展示，在更新/写入操作时应被忽略。 |
| **`wo`** | `KeyWriteOnly` | **只写** | 敏感字段（如密码），在视图展示或序列化时应隐藏。 |
| **`imm`** | `KeyImmutable` | **不可变** | 仅允许在创建时赋值，后续更新操作应禁止修改。 |
| **`min`** | `KeyMin` | **最小值/长度** | 数字比较大小；字符串、数组、Map 比较长度。 |
| **`max`** | `KeyMax` | **最大值/长度** | 数字比较大小；字符串、数组、Map 比较长度。 |
| **`range`** | `KeyRange` | **区间约束** | 数值必须在闭区间 `[min, max]` 内。 |
| **`in`** | `KeyIn` | **枚举约束** | 值必须存在于指定的参数列表中。 |
| **`re`** | `KeyRegexp` | **正则约束** | 值必须匹配指定的正则表达式。 |
*/
// 预定义核心指令（内置部分）
const (
	KeyRequired  Key = "req"
	KeyReadOnly  Key = "ro"
	KeyWriteOnly Key = "wo"
	KeyImmutable Key = "imm"
	KeyMin       Key = "min"
	KeyMax       Key = "max"
	KeyRange     Key = "range"
	KeyIn        Key = "in"
	KeyRegexp    Key = "re"
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
	Directives() []Directive
}

type ValidatorFunc func(val any, args []string) error

type ValueValidator interface {
	Register(k Key, fn ValidatorFunc)
	ValidateValue(val any, directives []Directive) error
}
