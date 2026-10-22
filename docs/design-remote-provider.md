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
| `UniqueConstraints` | 数据库唯一键声明，JSON 字段为 `uniqueConstraints` |
| `Indexes` | 普通索引声明，JSON 字段为 `indexes` |
| `valueValidator` | 赋值时使用的 value validator |
| `viewSpec` | 当前复制视图，用于控制赋值与导出边界 |

`Object.Copy(viewSpec)` 会保留模型身份，并按视图初始化字段值集合：

- `MetaView`：非指针字段初始化为零值，指针字段保持未赋值；
- `DetailView` / `LiteView`：只有声明了对应 view 的字段被初始化；
- `OriginView`：保留原始字段赋值状态。

### 2.2 Schema 元数据

`UniqueConstraints` 和 `Indexes` 随 `Object` JSON 传递，并在 `Copy(...)` 时保留。每项都由名称和基础字段名列表组成，例如：

```json
{
  "uniqueConstraints": [{"name": "uq_namespace_token", "fields": ["namespace", "tokenHash"]}],
  "indexes": [{"name": "idx_namespace_expire", "fields": ["namespace", "expireAt"]}]
}
```

它们是受模型校验的 schema 声明，不是调用方可注入的 SQL，也不参与普通查询过滤。`Create` 会应用这些声明；`Reconcile` 仅补建新增声明。详情见 [design-database.md](~/codespace/magicOrm/docs/design-database.md)。

2026-10-17 关系组合唯一主线：`UniqueConstraint.Verify` 接受基础字段与单值指针引用，继续拒绝 owned、多值关联及缺失类型。参与唯一声明的单值引用仅由主表同名列保存目标标量主键，列不继承目标主键的 auto/default；可选引用保持 nullable，`req` 继续由既有校验实施。普通引用仍使用链接表，不双写，不增加客户端替代字符串键。local struct 通过 `GetUniqueConstraints() []models.UniqueConstraint` 声明，local Copy 和 Helper 转 remote 保留防御复制后的声明。

PostgreSQL/MySQL 的 Create/Describe/AddColumn、Insert/Update/UpdateWithFilter、关系筛选和单项/批量回读共用该存储判定。写入前检查非 nil 引用目标主键，null 更新写主表 NULL，写宿主后不再写链接或目标；删除宿主不删除引用目标。查询仍使用既有关系投影和有界展开，只将关系键来源改为主表列。Schema Create/Drop/Inspect/Preflight/Repair 不再拥有该链接表。已有字段链接表与主表间的切换在 Reconcile 预检时要求显式迁移，不能以补列冒充数据迁移；新增可空引用列可先补列再建约束。

当前采用干净候选及配套组件，不搬迁旧链接数据、不自动回填。通用 ORM 不检查业务目标存在性、入向引用删除保护或树无环，这些由执行 owner 在相同事务边界实施。新增[主线回归](~/codespace/magicOrm/orm/inline_reference_test.go)覆盖双 provider、双数据库生成器、单宿主写、null、无主键拒绝、回读及增量恢复；[隔离数据库并发入口](~/codespace/magicOrm/orm/inline_reference_database_test.go)必须显式设置 `MAGICORM_INLINE_REFERENCE_TEST_ALLOW=isolated` 和 `MAGICORM_INLINE_REFERENCE_TEST_SERVER/DATABASE/SCHEMA/USER/PASSWORD`，使用随机表前缀及 ORM Create/Inspect/Insert/Count/Drop 验证两并发单成功、不同目标同时间合法、独立实例唯一范围，未授权时不连接。PostgreSQL 默认 build tag，MySQL 使用 `-tags=mysql`；本轮显式禁用该入口，真实数据库唯一冲突、并发及 Live 均为 `not-run`。

批量格式化被工具拒绝，未换工具覆盖；五仓 `upgrade_vendor.sh magicBase magicRunner magicCas magicFile magicModulesRepo` 也因重建现有 vendor 被拒绝，未手改 vendor 或使用模块模式绕过。因此最新消费者验证未执行，不复用旧资源批次的通过记录。该权限阻塞与真实数据库未验分别保留。

最新主线源码验证：models/provider/codec 全包、ORM `-short` 全包、PostgreSQL/MySQL 无数据库 builder 专项、MySQL build tag 的 models/provider/ORM、models/provider/ORM race、全仓 vet 与默认/MySQL 两种 build 均通过；最后补齐条件引用写、Schema 恢复和隔离 DB 测试入口后，相关检查再次通过。数据库测试显式禁用，未消费连接配置。只读 `gofmt -l` 仍报告 `~/codespace/magicOrm/database/codec/reference_column.go`、`~/codespace/magicOrm/orm/inline_reference_test.go`、`~/codespace/magicOrm/orm/inline_reference_database_test.go`，因覆盖拒绝未格式化；其余本批路径无格式差异。

后续用户已刷新 vendor，Base/Runner 在该基线的全仓 test/vet/build 通过，之前“消费者未执行”仅对应历史批次。随后 ORM 新增显式可序列化事务能力供执行 owner 在首读前使用，host/connection 两类 executor 均传递 `sql.LevelSerializable`；已有事务不能升级，失败不降级、不自动重放，Base 保持业务规则 owner。双后端内存 driver 的选项/嵌套/取消/失败恢复、ORM short 双 tag、全仓 vet/build 与事务 race 通过；这些不是实际数据库并发证据。新能力的五仓 vendor 同步再次被权限检查拒绝，未绕过，当前 Base 受保护写入因能力未配套明确拒绝，不能沿用上一基线的消费者结果。新建事务/规则文件格式化成功不解除上述三个旧文件的覆盖限制。

前一轮物理列拒绝准入验证记录：models/provider 全包、provider 全包 race、PostgreSQL/MySQL 无数据库 builder/Schema 声明专项、mysql build tag 的 models/provider、全仓 vet/build 均通过；补 local 字段对称与拒绝时无部分 DDL 断言后，models/provider、models race/mysql tag 和 vet/build 再次通过。未运行全仓数据库测试、物理 Schema catalog 测试或任何真实连接；没有设置或消费数据库 DSN。本批新增测试初次错误引用 builder 类型/codec 参数，修正接口后测试发现 typed-nil Type 会 panic；增加 nil 类型识别后负例通过。此项发生在 owner 源码，未操作共享数据库或覆盖 vendor。

### 2.3 `ObjectValue`

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

#### 数字的远程传输

`FieldValue` 的 JSON 解码保留数字原文为 `json.Number`，直到已知字段类型时再转换。
整数不会先转成 `float64`；`int64`、`uint64` 的合法边界和 `2^53` 附近的相邻值
在请求、响应、指针、集合及嵌套对象中保持精确。编码仍使用 JSON 数字，不要求改用字符串。

按声明的符号和位宽拒绝越界，浮点输入拒绝 NaN/Inf；普通数值转换保留已有的小数截断行为。
零值、显式赋值、复制和整数判等同样支持 `json.Number`，相邻大整数不能因浮点舍入被判为相同。
直接读取尚未转换的 `FieldValue.Value` 时，应使用类型转换工具，不能断言它必为 `float64`。

该规则只属于 remote 字段值协议；不修改任意业务 DTO 的 JSON 解码，也不包含租户、令牌、
流程恢复或数据库补偿规则。回归见 `provider/remote/integer_wire_test.go` 和 `utils/number_test.go`。

### 2.4 `SliceObjectValue`

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

### 6.3 值图循环与写响应复制边界

2026-10-16 源码补充：本地 struct 的 `GetObjectValue` / `GetSliceObjectValue` 转换使用每次调用独立的活动路径，按实际指针身份及切片底层地址/长度识别回环，退出分支后移除。直接环、两节点环和集合回环返回 `IllegalParam` 且不返回部分结果；共享但无环的子对象、有限同类型链、nil 与显式空集合保持原合同。活动指针/切片超过 256 个明确拒绝。`orm:"-"` 字段仍不进入该转换。

[写响应](~/codespace/magicOrm/orm/write_response.go)在 `Copy` 之前检查原始实体值图，避免复制已丢失输入身份后才试图检测循环；适用于 local 模型及有限远端元数据携带的 `ObjectValue` / `SliceObjectValue`。响应检查覆盖 local 深复制实际经过的导出字段（包括 `orm:"-"` 内存组合），也检查 map 的键和值；递归嵌套超过 256 层明确拒绝，不静默裁剪为成功。共享无环值不按全局 visited 误拒绝。

源码检查已通过：provider/orm 全目录在默认 PostgreSQL 和 `mysql` build-tag 下测试通过，helper/orm race、全仓 vet 及两种构建通过；真实数据库的 opt-in DSN 显式留空，未运行真实 DB 或 `~/codespace/magicOrm/test/` 的数据库集成集。另有 local/remote 单值自引用注册、GetTypeModel、单主表/单关系表、Create/Drop/Inspect/Preflight 计划专项通过；执行器仍是夹具，不能证明 DDL 已实际完成，也不关闭 Base/业务树门槛。

本项不修改持久模型、主键、关系所有权、Schema、权限、租户或业务规则，不改变写入顺序；写响应失败仍可能发生在数据提交后，必须回查而非盲目重写。Helper 对现有 remote 值的直接透传入口、Provider 注册/复制及其它 CRUD/查询入口并未因此获得全链循环保护；不据此放宽 Base 自引用准入。单值引用自环的注册、Schema、有限数据读写、删除保护及真实数据库回归仍需独立完成。

验证分为源码和现场：覆盖本地反射转换、local/remote 写响应的直接环/两节点环/集合环、共享无环、有限链、nil、过深链和导出内存组合回环；双后端 build-tag 回归只验证源码路径，不当作真实 DDL 或持久化证据。真实 DB/Live 未运行。后续用户已授权脚本并完成首次六仓依赖同步；随后 Base 源码变更，第二次同步被工具权限拒绝，最新消费者最终验证未完成。

### 6.4 引用主键与单值自引用验证

2026-10-16 继续收口：`InsertRunner` 在宿主写入前逐一解析全部单值指针及指针元素集合引用，要求目标具有有效非零主键；不会为引用目标分配 ID 或自动插入它。此前无主键引用可返回成功，新增 local/remote 负例先复现，再以通用校验修复；包含关系的递归创建与主键分配保持原合同。此校验只证明主键形状，不证明目标记录实际存在或允许关联。

Helper 对非 nil struct 指针和非 nil struct slice 输出 `Assigned:true`，包括字段全零的空对象和显式空集合；nil 仍未赋值。该意图经 remote 编解码保留，避免空引用对象被误当省略并绕过主键校验。`UpdateRunner.prepareNewRightIDs` 仅把无有效对象值的 nil 当作清空；非 nil 空对象继续检查主键并拒绝，不再因 `IsZero` 静默删除旧链接。显式 null 清空、空集合清空及有限包含对象语义保持。

[单值自引用执行回归](~/codespace/magicOrm/orm/self_reference_crud_test.go)使用 local/remote provider 与执行夹具，覆盖省略/null 根记录、自动主键后链接已有目标、以有限对象表示同主键引用、换绑/解除引用、只删除宿主及其出向链接、链接失败回滚、引用集合先验完整检查，以及无主键更新不修改旧链接。查询覆盖根、有限链、超出内置展开深度的链、同主键与双主键循环，单查/批查均有界返回可序列化值，并校验有限链精确祖先；数据库主键关系循环不是 Go 内存指针环，也不等于业务树允许循环。查询仍遵循既有展开深度，不提供完整树闭包。

最终源码验证：默认 PostgreSQL 与 `mysql` build-tag 的 provider/orm 全目录、helper/orm race、全仓 vet 和两种 build 通过；文档布局和相关仓库 diff 检查通过。负例首次失败分别暴露缺少引用主键及空对象当作清空，修复后重跑，不将首次失败隐藏为通过。真实数据库 opt-in DSN 显式为空，数据库集成目录未运行。后续已备份六仓依赖及差异并成功运行首次 `~/codespace/upgrade_vendor.sh`；同步后的 Base 全仓首先暴露查询空关联占位回归，已在 Base 查询边界修复，不撤销 ORM 写入赋值保护。Runner、CAS、File、共享模块首次同步后的 test/vet/build 通过；最新 Base 源码的第二次同步被工具权限拒绝，最终消费者基线尚未对齐。

本批没有新增 Module/Block、EventHub 合同、持久字段或 Schema；只改 ORM owner 内通用赋值和引用写入校验，不识别平台身份。引用删除不级联目标的回归**不是入向引用删除保护**，后者与树并发防环仍归业务执行合同配套；Base 已精确放行单值指针类型自引用的源码准入，业务树门槛不变。双 build-tag、race、vet/build 属源码证据，执行夹具不冒充实际提交或 DDL。直接消费者最新版本须在再次获准同步后另验；未执行数据库或 Live。

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
