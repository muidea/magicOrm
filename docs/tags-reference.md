# 标签参考

**依据**：`models/const.go`、`models/constraint.go`、README 与各设计文档中的 struct 标签约定。

[← 返回设计文档索引](README.md)

---

## 1. orm 标签

用于结构体字段，格式：`` `orm:"<名称> [key] [auto] [view:...]"` ``。

### 1.1 字段名

- 第一个标识符为**列名/字段名**（如 `orm:"uid"`、`orm:"name"`），对应数据库列名与 Model 字段访问名。
- **待确认**：列名与 Go 字段名的映射规则（是否支持驼峰转下划线、是否允许与 Go 字段名不同）是否在实现中统一？

### 1.2 key

- 表示该字段为**主键**。一个模型有且仅有一个主键字段。
- 示例：`orm:"id key"`、`orm:"uid key auto"`。

### 1.3 auto

- 表示**自增主键**，仅对数值类型主键有效；数据库侧自动生成值。
- 示例：`orm:"id key auto"`。
- 其它值声明（如 uuid、snowflake、datetime）见 models 注释，实现支持程度以代码为准。**待确认**：是否在文档中正式列出并说明各值声明？

### 1.4 关系字段

- 关系字段仍使用 `orm:"<名称>"`，类型为 `*T`、`T`、`[]*T`、`[]T` 等；关系类型判定见 [design-relation.md](design-relation.md)，不依赖额外标签。
- **待确认**：是否支持显式指定「引用/包含」或关系表名的标签（如 `orm:"groups ref"`）？

### 1.5 视图声明（view）

- 格式：`view:"detail,lite"` 等，表示该字段在哪些视图下可见/参与序列化。
- 可选值：`origin`、`detail`、`lite`、`basic` 等，与 `ViewDeclare` 常量一致（见 [design-models.md](design-models.md)）。
- 未带 view 的字段在默认/原始视图中仍可被访问；view 用于控制 Copy(viewSpec) 时的输出范围。**待确认**：view 与「查询时是否加载」的对应关系是否在文档中明确？

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
    ID     int     `orm:"uid key auto" view:"detail,lite"`
    Name   string  `orm:"name" constraint:"req,min=3,max=50" view:"detail,lite"`
    EMail  string  `orm:"email" view:"detail,lite"`
    Status *Status `orm:"status" view:"detail,lite"`
    Group  []*Group `orm:"group" view:"detail,lite"`
}
```

---

## 4. 参考

- 模型与视图：[design-models.md](design-models.md)
- 约束与验证：[design-validation.md](design-validation.md)
- 关联关系：[design-relation.md](design-relation.md)
- 类型与主键：[type-mapping.md](type-mapping.md)
