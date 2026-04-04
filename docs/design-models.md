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
| ValueMask(val) | 用实体值填充 Filter 的「掩码」：将 val 对应实体的字段值写入 Filter 内部，作为 Filter 的显式 mask 模型；在 `BatchQuery` 中它决定顶层响应字段裁剪，在其它依赖 `MaskModel()` 的路径中则提供显式 mask 形状。`ValueMask` 不放大子对象层级，子对象仍统一收敛到 `lite`。val 须与 Filter 绑定的类型一致。 |
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
| OriginView | 原始视图；不做 detail/lite 裁剪，通常等价于完整字段集；属于框架内部运行时视图，不是稳定的 struct tag 输入。 |
| MetaView | 元数据视图，包含全部字段，字段值按类型初始化；主要供 Provider 内部使用，不是稳定的 struct tag 输入。 |
| DetailView | 详细视图，需在类型定义中声明（struct 标签 `view:"detail,lite"` 等）。 |
| LiteView | 精简视图。 |

- **使用场景**：Copy(viewSpec) 后得到的 Model 仅包含该 view 声明的字段，用于控制序列化/输出范围，或在 Provider 内部构造 Filter/Model 的字段子集。字段是否出现在某视图由该字段的 `view:` 标签决定。
- **与查询加载的关系**：当前 `Query/BatchQuery` 的实现将“查询所需字段”和“最终返回字段”分开处理：
  - SQL 构造仍会基于 query mask 拉取足够的字段完成 host/basic/relation 回填；
  - 最终返回给调用方的模型按以下规则裁剪：
    - `Query(model)` 顶层对象固定按 `DetailView` 裁剪，不处理 `ValueMask`；
    - `BatchQuery(filter)` 若 `Filter.ValueMask(...)` 已指定，则顶层对象以 `ValueMask` 中显式包含的字段为准；
    - `BatchQuery(filter)` 若未指定 `ValueMask(...)`，则顶层对象以 Filter 绑定模型当前 view（如 `DetailView` / `LiteView`）为准；
    - `BatchQuery(filter)` 顶层对象优先级固定为 `ValueMask > view`；
    - relation 子对象统一按其自身 `LiteView` 裁剪，不受父对象 `detail` 或嵌套 `ValueMask` 放大影响；
    - 主键字段始终保留。
- **性能说明**：当前实现先修正“最终返回字段语义”，尚未把 SQL `SELECT` 列完全裁剪到与 `ValueMask/View` 一致；因此该机制首先解决返回契约问题，而不是直接优化数据库读列数。
- **隐式查询条件**：`Query(model)` 仍会把模型中的已赋值字段转成查询条件；但切片字段（如 `[]string`、`[]struct`、`[]*struct`）默认不再自动参与隐式条件构造，避免业务代码用空切片表达“我要返回这个字段”时被误翻译成 `WHERE`。需要按切片字段过滤时，应显式使用 `Filter.In(...)` / `Filter.NotIn(...)` 等操作符。
- **限制**：当前本地/远端视图 tag 解析稳定支持的视图标签值只有 `detail` 和 `lite`；其它值不会成为稳定的 struct tag 输入，写入 tag 时会被忽略。运行时内部仍保留 `origin`、`meta` 两类内部视图。
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
