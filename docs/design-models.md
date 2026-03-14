# 模型与查询设计

**功能块**：`models/`  
**依据**：`models/model.go`、`models/filter.go`、`models/constraint.go`、`models/field.go`、`models/const.go`。

[← 返回设计文档索引](README.md)

---

## 1. Model 接口

| 方法 | 说明 |
|------|------|
| GetName / GetShowName / GetPkgPath / GetPkgKey / GetDescription | 元信息 |
| GetFields / GetField(name) / GetPrimaryField | 字段访问 |
| SetFieldValue(name, val) / SetPrimaryFieldValue(val) | 设置字段/主键 |
| Interface(ptrValue bool) | 转成原始实体指针/值；ptrValue 为 true 时返回指针，为 false 时返回值；用于从 Model 取回原始 struct。 |
| Copy(viewSpec ViewDeclare) | 按视图复制 Model：返回一个新的 Model 实例，字段/元信息为浅拷贝；字段值保持原有引用/切片语义（不会深度复制嵌套/关系对象），仅根据视图裁剪「参与本次操作的字段集合」。 |
| Reset() | 重置 |
 
当前实现中，`Copy` 的主要用途是：  
- 在 Provider 内部构造只包含特定视图字段的 Model（如 MetaView/DetailView/LiteView）；  
- 在 Query/BatchQuery 前避免直接修改调用方传入的 Model（实现中会对传入 Model 做一次 `Copy(OriginView)`）。  
`Copy` 不改变底层表结构，也不会创建全新的「深度分离」对象图，如需完全独立的数据副本需在业务层自行实现。

---

## 2. Filter 接口

| 方法 | 说明 |
|------|------|
| Equal / NotEqual / Below / Above | 比较条件 |
| In / NotIn | 集合条件 |
| Like | 模糊匹配 |
| Pagination(pageNum, pageSize) / Sort(fieldName, ascFlag) | 分页与排序 |
| ValueMask(val) | 用实体值填充 Filter 的「掩码」：将 val 对应实体的字段值写入 Filter 内部，用于 BatchQuery/Count 时限定查询的「模型范围」或条件（如按某实体主键查）。val 须与 Filter 绑定的类型一致。 |
| MaskModel() | 返回当前 Filter 对应的 Model 实例（含 ValueMask 写入的掩码值）；用于 Runner 内部解析查询表与条件（如 QueryRunner、CountRunner 使用 MaskModel() 得到要查询的 Model）。 |
| Paginationer() / Sorter() / GetFilterItem(key) | 分页/排序/单项访问 |

**操作符常量**（`models` 包）：EqualOpr、NotEqualOpr、BelowOpr、AboveOpr、InOpr、NotInOpr、LikeOpr。

---

## 3. 约束（constraint 标签）

与 README 一致，支持以下 `models.Key`：

- **访问行为**：`req`（必填）、`ro`（只读）、`wo`（只写）。
- **内容值**：`min`、`max`、`range`、`in`、`re`（正则）。

约束在验证系统中使用，见 [design-validation.md](design-validation.md)。完整标签语法见 [tags-reference.md](tags-reference.md)。

### 3.1 主键与值声明

- **主键**：通过 orm 标签 `key` 指定，一个模型有且仅有一个主键字段。
- **值声明**：当前实现支持 `auto`、`uuid`、`snowflake`、`datetime` 四类 `ValueDeclare`。
- **插入时填充**：`Orm.Insert` 在 basic 字段为零值时，会分别填充自增主键回写值、UUID、雪花 ID 或当前时间；详见 [type-mapping.md](type-mapping.md)、[tags-reference.md](tags-reference.md)。

---

## 4. 视图（ViewDeclare）（评审 MOD-004）

| 视图 | 说明 |
|------|------|
| OriginView | 按 MaskValue 定义字段；MaskValue 为空时等价默认。 |
| MetaView | 元数据视图，包含全部字段，字段值按类型初始化；主要供 Provider 内部使用。 |
| DetailView | 详细视图，需在类型定义中声明（struct 标签 `view:"detail,lite"` 等）。 |
| BasicView | 基础视图常量存在于 `models`，但当前本地 struct tag 解析不会从 `view:"..."` 中识别该值。 |
| LiteView | 精简视图。 |

- **使用场景**：Copy(viewSpec) 后得到的 Model 仅包含该 view 声明的字段，用于控制序列化/输出范围，或在 Provider 内部构造 Filter/Model 的字段子集。字段是否出现在某视图由该字段的 `view:` 标签决定。
- **与查询加载的关系**：当前 `Query/BatchQuery` 的实现会在内部对传入 Model 执行一次 `Copy(OriginView)` 并基于完整字段集构造 SQL 与结果映射，**不会根据视图裁剪 SELECT 列表**。视图主要用于控制调用方使用 Model 进行序列化/输出时的字段集合；如未来在 Runner 层按视图裁剪查询列，将在本设计文档中补充说明。
- **限制**：当前本地 struct tag 解析稳定支持的视图标签值为 `detail` 和 `lite`；`origin`、`meta`、`basic` 属于框架内部/常量层概念，不是稳定的 struct tag 输入。
- **限制**：视图不改变表结构，仅影响内存中 Model 的字段子集与序列化结果。支持的基础类型与映射见 [type-mapping.md](type-mapping.md)。关联关系（一对一、一对多、多对多）见 [design-relation.md](design-relation.md)。

---

## 附录：操作符与约束键

### Filter 操作符（models 包常量）

| 常量 | 含义 |
|------|------|
| EqualOpr | 等于 (=) |
| NotEqualOpr | 不等于 (!=) |
| BelowOpr | 小于 (<) |
| AboveOpr | 大于 (>) |
| InOpr | 在集合内 (in) |
| NotInOpr | 不在集合内 (!in) |
| LikeOpr | 模糊匹配 (like) |

### 约束 Key（models.Key）

| Key | 含义 |
|-----|------|
| req | 必填 |
| ro | 只读 |
| wo | 只写 |
| min | 最小值/最小长度 |
| max | 最大值/最大长度 |
| range | 数值闭区间 [min, max] |
| in | 枚举约束 |
| re | 正则约束 |
