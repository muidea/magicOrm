# 标签参考

**依据**：`models/const.go`、`models/constraint.go`、README 与各设计文档中的 struct 标签约定。

[← 返回设计文档索引](README.md)

---

## 1. orm 标签

用于结构体字段，当前本地模型的稳定格式为：`` `orm:"<名称> [key] [auto|uuid|snowflake|datetime]"` ``。
`view` 与 `constraint` 是**独立标签**，不写在 `orm:"..."` 内。

### 1.1 字段名

- 第一个标识符为**列名/字段名**（如 `orm:"uid"`、`orm:"name"`），对应数据库列名与 Model 字段访问名。
- 当前实现不会自动做驼峰转下划线；数据库列名与模型字段名直接取 `orm` 标签中的第一个标识符。

### 1.2 key

- 表示该字段为**主键**。一个模型有且仅有一个主键字段。
- 示例：`orm:"id key"`、`orm:"uid key auto"`。

### 1.3 auto

- 表示**自增主键**，仅对数值类型主键有效；数据库侧自动生成值。
- 示例：`orm:"id key auto"`。
- 同一位置还支持 `uuid`、`snowflake`、`datetime`，示例：`orm:"uid key uuid"`、`orm:"createdAt datetime"`。

### 1.4 关系字段

- 关系字段仍使用 `orm:"<名称>"`，类型为 `*T`、`T`、`[]*T`、`[]T` 等；关系类型判定见 [design-relation.md](design-relation.md)，不依赖额外标签。
- 当前不支持通过额外 orm 标签显式指定「引用/包含」或自定义关系表名；关系语义完全由字段类型决定。

### 1.5 视图声明（view）

- 格式：`view:"detail,lite"` 等，表示该字段在哪些视图下可见/参与序列化。
- 当前本地 struct tag 解析稳定识别 `detail`、`lite`；其它值不是稳定的 struct tag 输入，写入 `view:"..."` 时会被忽略。运行时内部仍保留 `origin`、`meta` 两类内部视图。
- 未带 view 的字段在默认/原始视图中仍可被访问；view 用于控制 Copy(viewSpec) 时的输出范围，并作为默认查询响应字段收敛规则的一部分。
- 当前稳定查询约定：
  - `Query(model)` 不处理 `ValueMask`，顶层对象固定按 `DetailView` 返回；
  - `BatchQuery(filter)` 的顶层对象遵循固定优先级：`ValueMask > view`；
  - `BatchQuery(filter)` 若设置了 `ValueMask`，则顶层字段以 `ValueMask` 为准；未设置时才按 `view` 返回；
  - 包含/引用的子对象默认统一收敛到 `lite`，不因为父对象是 `detail` 或 `ValueMask` 中声明了嵌套子字段而自动扩成子对象 `detail`；
  - 业务若需要子对象详细信息，应基于子对象主键单独查询，而不是在一次列表/详情查询中继续放大多层结构。

### 1.6 子对象查询约束

- `magicOrm` 不再提供公开的 `relationView` tag，也不承诺通过字段级配置放大子对象默认查询层级。
- 默认查询规则属于框架内置特性：
  - `Query(model)` 顶层对象固定按 `DetailView` 裁剪；
  - `BatchQuery(filter)` 顶层对象按 `ValueMask > view` 的稳定优先级裁剪；
  - 包含/引用的子对象统一收敛到 `lite`；
  - 这一规则同时适用于默认视图 mask 生成和远端 schema/spec 表达。
- 业务若需要子对象详情，必须基于子对象主键单独查询，而不是在父对象查询里声明额外关系层级。

---

## 2. constraint 标签

用于约束校验，键与 [design-models.md](design-models.md) 附录一致。

| 键 | 含义 |
|----|------|
| req | 必填 |
| ro | 只读（如 Update 时忽略） |
| wo | 只写（如 Query 结果中不暴露） |
| min | 最小值/最小长度 |
| max | 最大值/最大长度 |
| range | 数值闭区间 [min, max] |
| in | 枚举约束 |
| re | 正则约束 |

示例：`` `constraint:"req,min=3,max=50"` ``、`` `constraint:"ro"` ``。约束在验证层使用，见 [design-validation.md](design-validation.md)。

---

## 3. 标签组合示例

```go
type User struct {
    ID     int      `orm:"uid key auto" view:"detail,lite"`
    Name   string   `orm:"name" constraint:"req,min=3,max=50" view:"detail,lite"`
    EMail  string   `orm:"email" view:"detail,lite"`
    Status *Status  `orm:"status" view:"detail,lite"`
    Group  []*Group `orm:"group" view:"detail,lite"`
}
```

---

## 4. 参考

- 模型与视图：[design-models.md](design-models.md)
- 约束与验证：[design-validation.md](design-validation.md)
- 关联关系：[design-relation.md](design-relation.md)
- 类型与主键：[type-mapping.md](type-mapping.md)
