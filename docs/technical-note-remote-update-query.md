# Remote / Update / Query 技术变更说明

## 1. 目标

本说明面向研发维护者，描述这轮修改解决了哪些实现偏差，以及当前代码的稳定语义。

核心目标：

- 让 remote provider 能可靠表达字段是否“参与本次写入”
- 让 `orm.Update` 在关系字段上按需更新，而不是大面积重建
- 让 `orm.Query` 的过滤逻辑与返回列逻辑解耦

---

## 2. 背景问题

### 2.1 Remote 没法稳定区分三种状态

变更前，remote 对单值字段容易混淆：

- 未赋值
- 显式 `nil`
- 显式零值

直接后果：

- `nil` pointer relation 容易被误解释成“清空关系”
- `0` / `false` / `""` / `[]` 容易在 helper 或 JSON 往返后丢失
- query 输入对象会把默认零值误带进过滤条件

### 2.2 Update 对包含关系过于粗暴

变更前，包含关系一旦变化，常见路径是：

- 删除旧关系
- 删除旧子对象
- 再重新插入

这会导致：

- 无变化也发生重建
- 局部变化也重建整组
- 关系更新精度不足

### 2.3 Query 把“过滤模型”当成“返回列模型”

变更前，`Query(model)` 使用查询模型本身决定 SELECT 列：

- 输入对象只带少数字段时，查询结果也只回填少数字段
- 对 `Reference` 这类包含大量 basic / slice / pointer 字段的对象，回填结果容易残缺

---

## 3. 关键实现调整

### 3.1 `models.IsAssignedField`

当前优先读取 value 自身的赋值状态：

- 如果 value 实现了 `IsAssigned()`，优先用它
- 否则回退到传统 `!IsZero()` 判定

这为 remote 和 local 共用统一入口提供了基础。

### 3.2 `remote.ValueImpl` 与 `remote.FieldValue`

新增并贯穿的状态：

- `ValueImpl.assigned`
- `FieldValue.Assigned`

当前规则：

- `ValueImpl.Set(nil)` 会保留“显式清空”
- `FieldValue.Assigned` 会跟随 JSON 编解码
- `Object.Interface(...)` 导出 `ObjectValue` 时会保留字段 assigned 状态

### 3.3 helper 导出规则

`provider/helper/remote.go` 当前约定：

- `nil` pointer / `nil` slice：未赋值
- 非 `nil` pointer：`Assigned:true`
- 非 `nil` slice：`Assigned:true`

这一步保证：

- 从本地 struct 到 remote `ObjectValue` 的语义不会在第一步就丢失

### 3.4 remote `SetModelValue(...)`

当前消费规则不是“看字段是否存在”，而是：

- `Assigned:true` 或者值本身非零：参与写入
- `Assigned:false` 且零值：跳过

这样可以同时支持：

- helper 导出的默认零值查询对象
- 手工构造的显式零值更新对象

### 3.5 `orm.Update`

当前 update 关系语义：

- 引用关系：只维护关系表差异
- 包含关系：
  - 先查当前数据库关系
  - 未变化：跳过
  - 单值包含且主键相同：对子对象原地 `Update`
  - 集合包含：按主键做增删改
  - 无法稳定识别时才回退整组替换

同时补了：

- remote 单值引用 `Assigned:true, Value:nil` 清空关系
- typed nil relation compare 防 panic

### 3.6 `orm.Query`

当前 query 分两层：

1. `getModelFilter(vModel, ...)`
   - 负责把查询模型转成过滤条件
2. `buildFullQueryMaskModel(...)`
   - 负责构造“用于拉取列”的 query mask

当前 query mask 规则：

- 自动补齐所有非指针 basic 字段
- 不自动补齐指针 basic / 指针 slice

原因：

- 非指针 basic 是“实体基础面”
- 指针字段通常承载“是否请求该字段”语义，不能无条件扩

---

## 4. 本轮重点修复的典型回归

### 4.1 Remote `Compose` / `OnlineEntity`

问题：

- helper 导出的 nil / zero 字段在 remote provider 中被误当成已赋值

结果：

- insert/update/query 条件被收紧
- query 结果或关系更新异常

修复：

- `Assigned` 贯穿 helper、JSON、provider

### 4.2 `Reference` 查询不一致

问题：

- `Query(model)` 返回列不完整
- 空 slice / 零值指针在 helper 导出后被误当成未赋值

修复：

- query mask 只补齐非指针 basic
- helper 对非 nil slice / pointer 保留 `Assigned:true`

### 4.3 Local typed nil panic

问题：

- 精准比较 relation 时，typed nil 进入反射路径导致 panic

修复：

- relation compare 先用 `models.IsValidField(...)` 判断

---

## 5. 当前稳定约定

后续修改应尽量保持以下约定稳定：

- remote 的“是否参与写入”由 `Assigned + ZeroValue` 共同决定
- helper 生成对象时，`nil` 和“非 nil 但为空”必须严格区分
- `Query(model)` 是“按模型过滤、顶层固定按 DetailView 返回”，不是“按输入对象字段或输入 view 裁剪返回列”
- 如果需要精确裁剪顶层返回列，走 `BatchQuery(filter)` + `Filter.ValueMask(...)`
- 包含/引用的子对象始终统一收敛到 `lite`

---

## 6. 不建议再走的方向

- 不要再把 `FieldValue.Assigned` 退回成纯内存态，否则 JSON 往返会再次丢语义
- 不要再让 `Query(model)` 受输入模型自身 view 或字段赋值形状控制顶层返回列，否则 `Reference` / `OnlineEntity` 一类问题会重现
- 不要把包含关系更新重新收回到“统一 delete + insert”，除非明确放弃精细 diff

---

## 7. 后续演进建议

- 如果后续需要 patch 风格 API，可直接沿用 `Assigned` 作为字段参与标记
- 如果要做更深的 update diff，可继续把“一对多包含”的主键 diff 扩展到字段级脏检查
- 如果要做外部协议文档，可从本说明和 `design-remote-provider.md` 直接抽出一份 Remote 调用指南
