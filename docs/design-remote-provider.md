# Remote Provider 设计

**功能块**：`provider/remote`、`provider/helper/remote.go`、`test/vmi/`  
**依据**：`provider/remote/*.go`、`provider/helper/remote.go`、`provider/vmi_remote_provider_test.go`、`provider/remote/*_test.go`。

[← 返回设计文档索引](README.md)

---

## 1. 目标与范围

本文档描述 `magicOrm` 当前实现下 `remoteProvider` 的真实设计语义，重点覆盖：

- `Object` / `ObjectValue` / `SliceObjectValue` 如何对齐 `models.Model` / `models.Value`；
- `test/vmi` 中 remote Object 定义如何映射到运行时模型；
- `provider/helper/remote.go` 中本地 struct 与 remote 对象值之间的往返规则；
- `filter` / `mask` / `codec` / `orm runner` / `database builder` 已验证的行为边界；
- 当前已固化的约定与限制，便于后续知识库沉淀。

本文档只描述**当前代码已经实现并验证通过**的行为，不扩展尚未落地的协议或抽象。

---

## 2. 核心对象模型

### 2.1 `Object`

`Object` 是 remote 侧的运行时模型，实现 `models.Model`，同时承担两类职责：

- **模型声明**：字段元信息、类型、约束、视图；
- **运行时值容器**：字段当前值、视图上下文、校验器。

关键字段：

| 字段 | 含义 |
|------|------|
| `Name` / `PkgPath` | 模型身份 |
| `Fields []*Field` | 字段定义与当前值 |
| `valueValidator` | 赋值时使用的 value validator |
| `viewSpec` | 当前复制视图，用于控制赋值与导出边界 |

`Object.Copy(viewSpec)` 会保留模型身份，并按视图初始化字段值集合：

- `MetaView`：非指针字段初始化为零值，指针字段保持未赋值；
- `DetailView` / `LiteView`：只有声明了对应 view 的字段被初始化；
- `OriginView`：保留原始字段赋值状态。

### 2.2 `ObjectValue`

`ObjectValue` 是 remote 侧的**值载体**，用于：

- 序列化/反序列化；
- provider 与 helper 间传递实体值；
- `filter.ValueMask`、`SetModelValue`、`QueryRunner` 回填等链路。

关键字段：

| 字段 | 含义 |
|------|------|
| `ID string` | 主键的字符串表达 |
| `Name` / `PkgPath` | 值所属模型身份 |
| `Fields []*FieldValue` | 字段值集合 |

当前实现约定：

- `Object.Interface(...)` 导出值时，主键同时保存在 `ObjectValue.ID` 和主键字段对应的 `FieldValue` 中；
- `CompareObjectValue` 判等时同时比较 `ID`、`Name`、`PkgPath`、字段名和值；
- 未赋值且无有效值壳的字段不会导出到 `Fields`；
- 显式赋值为 `nil` 的字段会导出，便于 update 场景表达“清空关系”；
- `FieldValue.Assigned` 是 remote 协议状态，会跟随 JSON 一起编码/解码，并影响 `SetModelValue(...)` 如何解释零值和 `Value:nil`。

### 2.3 `SliceObjectValue`

`SliceObjectValue` 表示 remote 侧的 struct slice 值，常用于关系字段或顶层对象集合。

关键字段：

| 字段 | 含义 |
|------|------|
| `Name` / `PkgPath` | slice 元素模型身份 |
| `Values []*ObjectValue` | 元素对象值集合 |

---

## 3. 关键语义

### 3.1 `nil` 与 `[]` 的区分

这是 remote 侧最重要的运行时语义之一。

从当前实现开始，remote 值同时区分：

- **未赋值**：字段未参与本次写入；
- **显式清空**：字段参与写入，但值被明确设为 `nil`；
- **显式空集合**：字段参与写入，集合值为 `[]`。

其中单值字段通过 `ValueImpl.assigned` 区分“未赋值”和“显式 `nil`”；导出为 `ObjectValue` 后，对应状态保存在 `FieldValue.Assigned`。

#### `assigned == false`

表示**字段未参与本次写入**：

- `models.IsAssignedField(field)` 返回 `false`；
- `Update` 不会处理该字段；
- `Object.Interface(...)` 只有在字段本身已有有效值壳时才会导出。

#### `assigned == true && value == nil`

表示**字段被显式赋值为 `nil`**：

- 对单值引用关系，表示清空关系；
- 对单值包含关系，表示删除旧子对象并清空关系；
- 在 `Object.Interface(...)` 导出时会保留该字段，`FieldValue.Value` 为 `nil`；
- 仅当 `FieldValue.Assigned == true && FieldValue.Value == nil` 时，`SetModelValue(...)` 才会把它解释成“显式清空”。

#### `FieldValue.Assigned == false && FieldValue.Value == nil`

表示**未赋值的 nil 字段**：

- 常见于 helper 从本地 struct 导出 remote `ObjectValue` 时的 nil 指针 relation；
- `SetModelValue(...)` 对 relation 字段会跳过，不会误解释成清空；
- 这样可以同时支持“按需更新跳过”和“显式 nil 清空”。

#### `FieldValue.Assigned == false && Value 为零值`

表示**helper 导出的默认零值，不作为显式写入**：

- 适用于 `0`、`false`、`""` 等 basic 零值，以及未赋值 relation shell；
- `SetModelValue(...)` 会跳过这些字段，避免 query/update 被默认零值误收紧；
- 如果调用方需要把字段显式更新为零值，必须传入 `FieldValue{Assigned:true, Value:<zero>}`。

helper 从本地 struct 导出 remote `ObjectValue` 时，当前进一步采用以下规则：

- `nil` 指针 / `nil` slice：导出为未赋值；
- 非 `nil` 指针：即使其指向零值，也导出为 `Assigned:true`；
- 非 `nil` slice：即使长度为 `0`，也导出为 `Assigned:true`。

这样可以在 remote 场景下区分：

- “字段未出现”
- “字段显式传入了空切片”
- “字段显式传入了非 nil 的零值指针”

#### `SliceObjectValue.Values == nil`

表示**未赋值 / 不覆盖目标**：

- 在 `ObjectValue` 中表示字段存在 relation shell，但没有显式赋值；
- 在 helper 回填本地实体时，不覆盖本地原值；
- 在 runner 写路径中，不触发“清空关系”。

#### `SliceObjectValue.Values != nil && len == 0`

表示**显式赋值为空 / 清空目标**：

- 在 helper 更新本地 struct slice 时，目标会被清空；
- 在 runner 写路径中，关系表会被视为需要清空。

该语义已经在字段级和顶层 slice 两层保持一致。

### 3.2 relation shell

当前实现允许 remote relation 字段保持“壳对象”：

- 单值 relation：`&ObjectValue{Name, PkgPath, Fields:nil}`
- 切片 relation：`&SliceObjectValue{Name, PkgPath, Values:nil}`

这类 shell 在 `EncodeObjectValue` / `DecodeObjectValue` / `ConvertObjectValue` 的往返过程中会被保留，不再退化成 `nil`。

补充约定：

- relation shell 本身不等于“显式清空”；
- 只有 `assigned == true && value == nil` 才表示单值字段清空；
- `SliceObjectValue.Values == nil` 仍表示未赋值，不表示清空。

### 3.3 指针语义

`TypeImpl.Interface` 当前语义如下：

- **basic / basic slice**：当 `IsPtr == true` 且输入非空时，保留指针语义；
- **struct / slice-struct**：统一转换为 `*ObjectValue` / `*SliceObjectValue`，不额外再包一层指针；
- **nil 输入**：返回该类型的初始化值壳。

### 3.4 视图边界

`Object` 会记录当前 `viewSpec`，并在以下场景遵守视图边界：

- `SetFieldValue`
- `SetPrimaryFieldValue`
- `filter.MaskModel()`
- `GetEntityFilter(..., LiteView/DetailView)`

重要约定：

- 当前实现下，字段是否属于 `LiteView` / `DetailView`，完全由字段 `Spec.ViewDeclare` 决定；
- 主键字段如果需要参与 `LiteView` / `DetailView` 赋值与导出，也必须显式声明相应 view；
- 这与当前仓库中的本地模型和 VMI 定义保持一致。

---

## 4. Provider 入口语义

`provider/remote/provider.go` 当前暴露的核心语义如下。

### 4.1 `GetEntityType`

支持输入：

- `*Object` / `Object`
- `*ObjectValue` / `ObjectValue`
- `*SliceObjectValue` / `SliceObjectValue`

输出规则：

- `Object` / `ObjectValue` -> `TypeStructValue`
- `SliceObjectValue` -> `TypeSliceValue`，`ElemType` 为 struct type

### 4.2 `GetEntityValue`

仅接受：

- `ObjectValue`
- `SliceObjectValue`

返回 `ValueImpl` 包装值。

### 4.3 `GetEntityModel`

仅接受：

- `*Object`
- `Object`

并将传入的 `valueValidator` 绑定到 `Object` 上。

### 4.4 `SetModelValue`

输入分两类：

- `vVal.Get()` 为 `*ObjectValue`：逐字段写回 `Object`；
- 其他有效值：视为主键值，写入 primary field。

当前实现约定：

- 未知字段会被忽略，不报错；
- `nil` `vModel` / `vVal` 会报错；
- 类型断言 panic 会被 recover 并转换为 `*cd.Error`。

### 4.5 `GetModelFilter`

仅接受 `*Object`，返回 `ObjectFilter`。

---

## 5. Filter / Mask 语义

`ObjectFilter` 绑定一个 `Object`，提供与 `models.Filter` 对齐的 remote 过滤能力。

### 5.1 条件语义

支持：

- `Equal`
- `NotEqual`
- `Below`
- `Above`
- `In`
- `NotIn`
- `Like`
- `Pagination`
- `Sort`

其中：

- `In/NotIn` 支持 basic slice，也支持 `*SliceObjectValue`；
- `GetFilterItem` 按 `Equal -> NotEqual -> Below -> Above -> In -> NotIn -> Like` 的顺序返回首个条件。

### 5.2 `ValueMask`

接受：

- `*ObjectValue`
- `ObjectValue`
- `json.RawMessage`

并要求：

- `Name` 与绑定模型一致；
- `PkgPath` 与绑定模型一致。

typed nil 的 `*ObjectValue` 会被忽略；非法类型或模型不匹配会返回错误。

### 5.3 `MaskModel`

`MaskModel()` 的当前实现规则：

- 基于 `bindObject.Copy(models.OriginView)` 复制；
- 复制结果重新挂载原始 `viewSpec`；
- 再把 `MaskValue` 中的字段写入复制对象。

这样可以保证：

- 不污染原始绑定对象；
- 不越过当前视图边界。

---

## 6. Helper 往返转换规则

`provider/helper/remote.go` 提供本地 struct 与 remote 值之间的桥接。

### 6.1 Local -> Remote

入口：

- `GetObject`
- `GetObjectValue`
- `GetSliceObjectValue`

转换规则：

- basic 字段通过 local codec 编码；
- struct 字段转换为 `*ObjectValue`；
- struct slice 转换为 `*SliceObjectValue`；
- `nil` 指针字段保留为 `nil`；
- `nil` struct slice 保留为 `Values:nil`；
- 显式空 struct slice 保留为 `Values:[]`。

### 6.2 Remote -> Local

入口：

- `UpdateEntity`
- `UpdateSliceEntity`

回填规则：

- basic 字段通过 local codec 解码；
- struct 字段递归写回本地 model；
- struct slice 在 `Values == nil` 时不覆盖目标；
- struct slice 在 `Values != nil` 时先清空，再重建；
- 顶层 slice 与字段级 slice 使用同一套 `nil/[]` 语义。

---

## 7. VMI 定义对齐

`test/vmi/entity/**/*.json` 是当前 remote 模型设计的事实样例。

已验证的 VMI 对齐点包括：

- 全量 JSON 可以反序列化为 `remote.Object` 并通过 `models.VerifyModel`；
- 跨模型字段类型可通过 `GetTypeModel` 正确解析；
- 典型字段类型语义正确：
  - `product.skuInfo` -> slice struct
  - `product.image` -> basic slice
  - `stockin.goodsInfo` -> slice struct
  - `rewardPolicy.scope` -> struct
  - `rewardPolicy.item` -> slice struct
  - `store.goods` / `store.shelf` -> slice ptr-struct

VMI 定义在当前实现中的角色：

- 既是 remote schema 样本；
- 也是 provider / helper / runner / builder 的端到端回归基线。

---

## 8. 编解码与 SQL 生成语义

### 8.1 remote codec

`provider/remote/codec.go` 只负责 **basic / basic slice** 的编码和解码。

规则：

- struct / slice-struct 不走这里，交给 `ObjectValue` / `SliceObjectValue`；
- pointer basic 和 pointer basic slice 会保留指针语义；
- 非法输入返回 `*cd.Error`。

### 8.2 relation 主键压缩

在 provider / database builder 链路中，relation 字段会压缩为主键值：

- `product.status` -> `status.id`
- `product.skuInfo` -> `skuInfo.sku`

该行为已在 MySQL / Postgres builder 的 VMI 回归中固定。

### 8.3 已验证 SQL 路径

基于 VMI remote 定义，已验证：

- 主表 `BuildCreateTable`
- 主表 `BuildInsert`
- 主表 `BuildUpdate`
- 主表 `BuildDelete`
- relation 表 `BuildCreateRelationTable`
- relation 表 `BuildInsertRelation`
- relation 表 `BuildQueryRelation`
- relation 表 `BuildDeleteRelation`
- relation 表 `BuildDeleteRelationByRights`
- relation filter `BuildQuery` / `BuildCount`

---

## 9. ORM Runner 已验证链路

基于 fake executor 与 VMI 定义，当前已验证以下链路：

- `GetEntityModel -> GetEntityFilter -> ValueMask / MaskModel`
- `QueryRunner` 主表查询与 relation 回填
- `InsertRunner`
- `UpdateRunner`
- `DeleteRunner`

重点已验证语义：

- relation 表无记录时，单值 relation 保持 `nil`；
- relation slice 无记录时，保持未赋值 relation shell；
- 写路径中 `nil` / `[]` 语义与 helper、filter 保持一致。

---

## 10. 当前实现约束

### 10.1 已固化约束

- `remoteProvider` 的运行时模型以 `Object` 为核心，不直接操作任意 map；
- `ObjectValue` / `SliceObjectValue` 是 remote 值传递的标准形态；
- `nil` 与 `[]` 在 relation slice 上必须严格区分；
- 视图裁剪依赖字段级 `ViewDeclare`；
- 远端 relation shell 必须在 JSON 往返中保留。

### 10.2 当前未扩展的行为

- 未定义独立的 remote 传输协议文档，当前知识应以代码和 VMI 样例为准；
- `SetModelValue` 对未知字段当前采取忽略策略，而非严格拒绝；
- `convertRawStruct(int)` 之类明显非法输入当前仍有个别“返回 nil, nil”的旧行为，虽然不会影响已验证主路径，但不是推荐使用方式。

---

## 11. 已完成验证范围

围绕 `test/vmi` 与 `provider/remote`，本轮已经形成以下回归保护网：

- schema 定义校验
- provider wrapper
- Object/ObjectValue/SliceObjectValue 赋值与比较
- view / filter / mask
- helper 双向转换
- remote codec
- ORM query / insert / update / delete runner
- MySQL / Postgres SQL builder

这组回归已经覆盖 remote provider 的主运行链路，可作为后续 remote 相关知识库和回归基线。
