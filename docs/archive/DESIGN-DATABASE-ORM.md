# magicOrm 数据库 ORM 操作设计文档

本文档描述基于 **models.Model** 及相关定义实现的**数据库 ORM 操作**的设计，与 **DESIGN-CONSISTENCY.md**（数据一致性与 Provider 模型抽象）互补：一致性文档约定 Model/Field/Type/Value 的接口与转换规则，本文档约定**如何用这些抽象完成建表、增删改查与事务**。

---

## 1. 文档范围与前置依赖

### 1.1 范围

- **在库**：基于 `models.Model`、`models.Field`、`models.Type`、`models.Value`、`models.Filter` 的数据库操作（Create/Drop、Insert/Update/Delete/Query、Count、BatchQuery、事务）。
- **不在库**：具体存储引擎、连接串格式、部署与运维；这些由其他文档或配置说明。

### 1.2 前置依赖

- **models 抽象**：以 **DESIGN-CONSISTENCY.md** 为准。ORM 层**仅**通过以下接口访问模型与字段，不依赖 Local/Remote 具体实现：
  - **Model**：`GetName/GetPkgPath/GetPkgKey`、`GetFields`、`GetPrimaryField`、`SetFieldValue/SetPrimaryFieldValue`、`GetField(name)`、`Interface(ptrValue)`、`Copy(viewSpec)`、`Reset`。
  - **Field**：`GetName`、`GetType`、`GetSpec`、`GetValue`、`SetValue`；`GetSliceValue`、`AppendSliceValue`、`Reset`（切片字段）；以及 `IsValidField`、`IsAssignedField` 等语义。
  - **Type**：`GetValue`（TypeDeclare）、`IsPtrType`、`Elem()`（元素类型）、`GetPkgKey`。
  - **Value**：`Get`、`Set`、`IsValid`、`IsZero`、`UnpackValue`（切片时）。
- **Provider**：ORM 通过 **provider.Provider** 获取类型模型、编解码值（`GetTypeModel`、`EncodeValue`、`DecodeValue`、`SetModelValue` 等），从而与 DESIGN-CONSISTENCY 中的 Local/Remote 行为一致。

---

## 2. 当前实现架构（按实现整理）

### 2.1 总体分层

```
┌─────────────────────────────────────────────────────────────────┐
│  Orm 接口 (orm/orm.go)                                           │
│  Create / Drop / Insert / Update / Delete / Query / Count /      │
│  BatchQuery / BeginTransaction / CommitTransaction / Rollback    │
└───────────────────────────────┬─────────────────────────────────┘
                                │ 使用
┌───────────────────────────────▼─────────────────────────────────┐
│  Runner 层 (orm/*.go)                                            │
│  CreateRunner / InsertRunner / UpdateRunner / DeleteRunner /      │
│  QueryRunner；baseRunner 持有 vModel, executor, modelCodec        │
└───────────────────────────────┬─────────────────────────────────┘
                                │ 使用
┌───────────────────────────────▼─────────────────────────────────┐
│  database.Builder (database/postgres|mysql/builder*.go)         │
│  将 Model/Filter 转为 SQL + 参数：BuildCreateTable, BuildInsert, │
│  BuildUpdate, BuildDelete, BuildQuery, BuildCount, 关系表等      │
└───────────────────────────────┬─────────────────────────────────┘
                                │ 使用
┌───────────────────────────────▼─────────────────────────────────┐
│  database.Codec (database/codec/codec.go)                       │
│  Model/Field/Value ↔ 可写入 DB 的值：PackedBasicFieldValue,       │
│  PackedStructFieldValue, PackedSliceStructFieldValue,            │
│  ExtractBasicFieldValue；表名/关系表名构造                        │
└───────────────────────────────┬─────────────────────────────────┘
                                │ 使用
┌───────────────────────────────▼─────────────────────────────────┐
│  provider.Provider + database.Executor                          │
│  Provider：GetTypeModel, EncodeValue, DecodeValue, SetModelValue │
│  Executor：Execute, ExecuteInsert, Query, Next, GetField, 事务    │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 核心接口与类型（当前实现）

**Orm 接口**（`orm/orm.go`）

```go
type Orm interface {
    Create(entity models.Model) *cd.Error
    Drop(entity models.Model) *cd.Error
    Insert(entity models.Model) (models.Model, *cd.Error)
    Update(entity models.Model) (models.Model, *cd.Error)
    Delete(entity models.Model) (models.Model, *cd.Error)
    Query(entity models.Model) (models.Model, *cd.Error)
    Count(filter models.Filter) (int64, *cd.Error)
    BatchQuery(filter models.Filter) ([]models.Model, *cd.Error)
    BeginTransaction() *cd.Error
    CommitTransaction() *cd.Error
    RollbackTransaction() *cd.Error
    Release()
}
```

- 入参/出参均为 **models.Model** 或 **models.Filter**，不暴露具体 Provider 或数据库类型。
- Insert/Update/Delete/Query 返回的均为 Model，保证上层仅依赖 Model 接口。

**database.Executor**（`database/executor.go`）

```go
type Executor interface {
    Release()
    BeginTransaction() *cd.Error
    CommitTransaction() *cd.Error
    RollbackTransaction() *cd.Error
    Query(sql string, needCols bool, args ...any) (ret []string, err *cd.Error)
    Next() bool
    Finish()
    GetField(value ...any) *cd.Error
    Execute(sql string, args ...any) (rowsAffected int64, err *cd.Error)
    ExecuteInsert(sql string, pkValOut any, args ...any) (err *cd.Error)
    CheckTableExist(tableName string) (bool, *cd.Error)
}
```

- 与具体数据库（PostgreSQL/MySQL）解耦，由 `database/postgres`、`database/mysql` 实现。

**database.Builder**（`database/builder.go`）

- 所有方法均以 **models.Model** 或 **models.Model + models.Filter** 或 **models.Field** 为输入，输出 `Result`（SQL + Args）。
- 表名、关系表名通过 **database.Codec** 的 `ConstructModelTableName`、`ConstructRelationTableName` 生成，依赖 Model/Field 的 GetName/GetPkgPath/GetPkgKey，不依赖具体 Provider 实现。

**database.Codec**（`database/codec/codec.go`）

- **ConstructModelTableName(vIdentifier)**：主表名（含可选前缀）。
- **ConstructRelationTableName(vModel, vField)**：关系表名（一对一/一对多等由 Field 的 Type 推导）。
- **PackedBasicFieldValue(vField, vVal)**：基础字段（含基础类型切片）的 Model 值 → 可写入 DB 的值（通过 Provider.EncodeValue，如切片 JSON 序列化）。
- **PackedStructFieldValue / PackedSliceStructFieldValue**：结构体/结构体切片 → 外键或关联值（通过 GetTypeModel、SetModelValue、再取主键值）。
- **ExtractBasicFieldValue(vField, eVal)**：从 DB 读出的值 → Model 可接受的值（通过 Provider.DecodeValue）。

约束：**所有对“模型值”的读写与类型转换均通过 Model/Field/Value 与 Provider 完成**，与 DESIGN-CONSISTENCY 一致。

### 2.3 操作流程（当前实现摘要）

| 操作 | 入口 | 主要步骤 |
|------|------|----------|
| **Create** | `impl.Create(vModel)` | 校验 context → CreateRunner.Create() → BuildCreateTable(vModel) → 主表 Execute；对非基础字段 BuildCreateRelationTable → 建关系表 |
| **Drop** | `impl.Drop(vModel)` | 校验 context → DropRunner.Drop() → 先删关系表再删主表（BuildDropRelationTable / BuildDropTable） |
| **Insert** | `impl.Insert(vModel)` | 校验 context、vModel 非 nil → 场景校验（ScenarioInsert）→ BeginTransaction → InsertRunner.Insert()：主表 insertHost（BuildInsert + ExecuteInsert，自增/UUID/DateTime 等由 Spec 在 Runner 内填值）→ 非基础字段 insertRelation（关系表插入）→ Commit/Rollback |
| **Update** | `impl.Update(vModel)` | 校验 context → 场景校验（ScenarioUpdate）→ UpdateRunner.Update()：updateHost(BuildUpdate + Execute) → 各已赋值非基础字段 updateRelation（引用关系：查当前 right IDs → 与新 right IDs 差集 → 仅删/插差异链接；包含关系：先 deleteRelation 再 insertRelation）→ 返回 vModel |
| **Delete** | `impl.Delete(vModel)` | 校验 context → 场景校验（ScenarioDelete）→ DeleteRunner.Delete()：先 deleteRelation 再 deleteHost（BuildDelete + Execute） |
| **Query** | `impl.Query(vModel)` | 校验 context、vModel 非 nil → vModel.Copy(OriginView) → getModelFilter(vModel) 得到 Filter（主键或已赋值字段转 Equal/In）→ QueryRunner.Query(vFilter)：BuildQuery + executor.Query + Next + GetField → assignBasicField 等写回 Model → 返回单条 Model，若无结果返回 cd.NotFound |
| **Count** | `impl.Count(vFilter)` | 校验 context → BuildCount(vFilter.MaskModel(), vFilter) → Execute → 解析 count 结果 |
| **BatchQuery** | `impl.BatchQuery(vFilter)` | 校验 context → BuildQuery + 循环 Next/GetField → 每条赋到 Model 后加入列表返回 |

所有写库的“值”均来自 **Model 的 Field/Value**，经 **Codec.Packed*** 与 **Provider.EncodeValue** 得到；读库经 **Executor.GetField** 与 **Codec.ExtractBasicFieldValue**、**Provider.DecodeValue** 写回 **Field.SetValue**，从而保证**仅通过 Model/Field/Value 接口与 Provider 访问数据**。

### 2.4 验证与事务

- **验证**：Insert/Update/Delete 前通过 `validateModel(vModel, scenario)` 做场景校验（Insert/Update/Query/Delete 对应不同 Scenario），内部使用 `validation.ValidationManager`，与 Model 的约束、Spec 一致。
- **事务**：Insert/Update/Delete 在 impl 层显式 BeginTransaction，defer 中根据 err  Commit 或 Rollback；也可由调用方 BeginTransaction → 多次操作 → CommitTransaction/RollbackTransaction。

### 2.5 关系与表结构约定（当前实现）

- **主表**：一 Model 一主表，表名由 Codec.ConstructModelTableName(Model) 得到（如首字母大写的 Name + 可选前缀）。
- **关系**：非基础类型字段（结构体/结构体切片）对应关系表，表名由 ConstructRelationTableName(vModel, vField) 得到；关系类型由 `getFieldRelation(vField)` 根据 Type 的指针/切片推导（Has1v1、Has1vn、Ref1v1、Ref1vn）。
- **主键**：依赖 Model 的 GetPrimaryField()，单主键；Insert 的 RETURNING 主键写回通过 Codec.ExtractBasicFieldValue + pkField.SetValue。

### 2.6 Orm 契约与实现约定

以下为 Orm 接口及实现的**正式约定**，实现与对接方须遵守。

**入参与错误码**

- 所有接受 **entity（models.Model）** 的方法（Create、Drop、Insert、Update、Delete、Query）：当 `entity == nil` 时，返回 `cd.IllegalParam`，错误信息建议为 "illegal model value" 或等价描述。
- **Count**、**BatchQuery**：当 `filter == nil` 时，返回 `cd.IllegalParam`，错误信息建议为 "illegal filter value" 或 "filter is nil"。
- 其他非法入参（如验证失败、context 已取消等）按实现返回相应错误码（如 `cd.IllegalParam`、`cd.Unexpected`）；数据库执行失败可为 `cd.DatabaseError` 或由底层返回的错误码。

**Update 关系更新策略**

- **关系类型区分**：Update 时按字段的 `Elem().IsPtrType()` 区分两种关系，采用不同策略：
  - **引用关系**（`Elem().IsPtrType() == true`，如 `*Child`、`[]*Role`）：**按差异增量更新**——先查询当前关系表中该 host 字段的 right ID 集合，与本次要写入的新 right ID 集合做差集，仅删除需移除的链接（`BuildDeleteRelationByRights`）、仅插入需新增的链接（`BuildInsertRelation`），不变的不动。不删除关联实体本身（关联实体独立存在）。
  - **包含关系**（`Elem().IsPtrType() == false`，如 `Child`、`[]Tag`）：**以新换旧**——先对该字段执行 DeleteRelation（删除关系表行并级联删除关联实体），再按当前 Model 中该字段的值执行 InsertRelation（创建新关联实体并插入关系表行）。行为与原有一致。
- **引用关系约束**：引用关系下，关联实体必须已存在于库中且有有效主键；若主键为空视为非法入参，返回 `cd.IllegalParam`。
- 未赋值的非基础字段视为不更新该关系（在 `Update()` 循环中通过 `IsAssignedField` 过滤跳过）。
- **slice 赋值语义**：属性类型为 slice 时，**nil=未赋值**、**[]=已赋值（size 0）**，用于引用关系「清空」等场景；实现见 Value 层 `IsZero()` 与 MetaView 下 slice 重置，详见 **DESIGN-UPDATE-RELATION-DIFF.md** 第 1.4 节。
- **Query 选列与 slice 一致**：Query 时基础列中仅「已赋值」或「值类型 slice」（如 `[]int`）参与 SELECT 与赋值，指针型未赋值不拉取，与 slice 语义一致；详见 **docs/QUERY-SLICE-SEMANTICS-FIX.md**。
- 详细实现方案与实现清单见 **DESIGN-UPDATE-RELATION-DIFF.md**（已实现并归档）。

**无结果语义**

- **Query**：若无匹配记录，返回 `(nil, *cd.Error)`，错误码为 `cd.NotFound`，错误信息中可包含 model pkgKey、filter 等便于排查。
- **Count**：无匹配时返回 `(0, nil)`，即数量为 0、无错误。
- **BatchQuery**：无匹配时返回 `([]models.Model{}, nil)`，即空切片、无错误。

**Filter 与 Model**

- **Filter** 与某一 **Model** 对应（同一 GetPkgKey）；由 Model 通过 getModelFilter 得到的 Filter 与该 Model 对应。
- **BuildQuery**、**BuildCount** 等需要「模型结构」时，使用 `filter.MaskModel()` 取得 Model，Builder 仅依赖 Model 接口，不直接依赖 Filter 的内部表示。

**事务边界**

- **单次 Insert/Update/Delete**：默认在**自身事务**中执行；impl 内部 BeginTransaction，在 defer 中根据 err 执行 Commit 或 Rollback。
- **调用方显式事务**：若调用方先调用 BeginTransaction，则后续多次操作共用一个事务；由调用方在适当时机调用 CommitTransaction 或 RollbackTransaction，impl 的 Insert/Update/Delete 在已有事务的 executor 上执行，不再单独 Begin。

**Codec 与 Builder 职责**

- **Codec**：不生成 SQL；仅负责「值 ↔ 可写入/读出 DB 的值」的转换，以及「表名/关系表名」的构造（ConstructModelTableName、ConstructRelationTableName）。所有对 Model 值的读写与类型转换均通过 Model/Field/Value 与 Provider 完成。
- **Builder**：不解析 Model 的原始类型（如 reflect）；仅通过 Codec 与 Model/Field/Value 接口获取表名、可绑定参数与占位符，生成 SQL 与参数列表，便于扩展多数据源或多方言。

**日志与可观测性**

- 关键失败路径（如 BuildInsert/BuildQuery、Execute、Validate、GetModelFilter）的日志须带齐：**method/operation**（如 Insert、insertHost、BuildQuery）、**model/filter 标识**（如 pkgKey、field）、**error**（err.Error()），与 DESIGN-CONSISTENCY 验证后的日志风格一致，便于排查。

### 2.7 操作语义与错误码速查

| 操作 | 成功语义 | 无结果/空 | 常见失败与错误码 |
|------|----------|------------|------------------|
| **Create** | 主表及关系表创建成功 | — | entity==nil → IllegalParam；建表/执行失败 → DatabaseError 或底层错误 |
| **Drop** | 关系表及主表删除成功 | — | entity==nil → IllegalParam；执行失败 → DatabaseError 或底层错误 |
| **Insert** | 返回带主键等写回后的 Model | — | entity==nil → IllegalParam；校验/执行失败 → IllegalParam/DatabaseError 等 |
| **Update** | 返回更新后的 Model（同一引用） | — | entity==nil → IllegalParam；校验/执行失败 → IllegalParam/DatabaseError 等 |
| **Delete** | 返回被删的 Model（同一引用） | — | entity==nil → IllegalParam；校验/执行失败 → IllegalParam/DatabaseError 等 |
| **Query** | 返回单条匹配的 Model | 无匹配 → (nil, NotFound) | entity==nil → IllegalParam；getModelFilter/执行失败 → 对应错误码 |
| **Count** | 返回匹配条数 int64 | 无匹配 → 0 | filter==nil → IllegalParam；执行失败 → DatabaseError 等 |
| **BatchQuery** | 返回匹配的 Model 切片 | 无匹配 → [] | filter==nil → IllegalParam；执行失败 → 对应错误码 |

---

## 3. 设计评估与完善建议

### 3.1 与 models.Model 及 DESIGN-CONSISTENCY 的符合度

| 维度 | 评估 | 说明 |
|------|------|------|
| 仅通过 Model/Field/Value 访问数据 | **符合** | Builder/Codec/Runner 均只使用 Model 的 GetFields、GetField、GetPrimaryField、GetValue/SetValue、GetType、GetSpec 等，无直接 reflect 或 Remote 专有类型。 |
| 编解码经 Provider 统一 | **符合** | Codec 的 Packed*/Extract* 均通过 provider.EncodeValue/DecodeValue、GetTypeModel、SetModelValue 完成，与 DESIGN-CONSISTENCY 的 Local/Remote 行为一致。 |
| 表名/关系名仅依赖 Model 元数据 | **符合** | 表名、关系表名仅用 GetName、GetPkgPath、GetPkgKey 及 Field 的 Type，无业务数据依赖。 |

结论：**当前数据库 ORM 设计在“基于 models.Model 及相关定义”这一点上与 DESIGN-CONSISTENCY 一致，可视为该文档在数据库侧的落地实现。**

### 3.2 可完善点（设计层面）

1. **入参校验与错误码**
   - **现状**：Create/Drop/Count/BatchQuery 未在文档或实现中统一约定对 `entity == nil` / `filter == nil` 的校验与错误码（Insert/Update/Delete/Query 部分已有 nil 校验）。
   - **建议**：在 Orm 接口的语义说明中明确：所有接受 Model 的方法在 `entity == nil` 时返回 `cd.IllegalParam`；Count/BatchQuery 在 `filter == nil` 时返回 `cd.IllegalParam`；实现上已满足的保持，未满足的补齐并与 helper 层错误码风格统一。

2. **Update 关系更新策略**
   - **现状**：已按 **DESIGN-UPDATE-RELATION-DIFF.md** 实现——引用关系按差异增量更新（仅刷新链接）、包含关系以新换旧；slice 语义（nil=未赋值、[]=已赋值）与 Query 选列已与实现一致，见 2.6 与 docs/QUERY-SLICE-SEMANTICS-FIX.md。

3. **Query 无结果时的契约**
   - **现状**：Query 无记录时返回 `cd.NotFound` 与描述信息，符合常见契约。
   - **建议**：在 Orm 接口说明中显式写出：Query 返回 (nil, cd.NotFound) 表示无匹配记录；Count 与 BatchQuery 的“无结果”语义（Count=0、BatchQuery=空切片）一并写明。

4. **Filter 与 Model 的对应关系**
   - **现状**：BatchQuery/Count 使用 Filter；Query 使用 Model 转 Filter（getModelFilter）。Filter 的 MaskModel() 用于 Builder 需要“模型结构”的场景。
   - **建议**：在文档中明确：Filter 与某 Model 对应（同一 GetPkgKey）；BuildQuery/BuildCount 需要“结构”时使用 filter.MaskModel()，保证 Builder 只依赖 Model 接口。

5. **事务边界与 Runner 复用**
   - **现状**：Insert/Update/Delete 在 impl 内各自 BeginTransaction + defer Commit/Rollback；Runner 不持有事务状态，由 impl 传入 executor。
   - **建议**：在文档中说明“单次 Insert/Update/Delete 默认在自身事务中执行”；若调用方先 BeginTransaction，则多次操作共用一个事务，由调用方 Commit/Rollback。

6. **Codec 与 Builder 的职责边界**
   - **现状**：Codec 负责“值 ↔ DB 值”与“表名/关系表名”；Builder 负责“Model/Filter → SQL 与参数”，并调用 Codec 取表名和字段值。
   - **建议**：在文档中固定为：**Codec** 不生成 SQL，只做标识符与值的转换；**Builder** 不解析 Model 的原始类型，只通过 Codec 和 Model/Field/Value 接口获取表名与可绑定参数，便于后续扩展多数据源或多方言。

7. **日志与可观测性**
   - **现状**：Runner/impl 中部分错误路径仍使用通用 "operation failed" 等日志。
   - **建议**：在“实现约定”中要求关键路径（如 BuildInsert/BuildQuery、Execute、Validate）的失败日志带齐 method、operation、model/filter 标识（如 pkgKey）、error，与 DESIGN-CONSISTENCY 验证后的日志风格一致，便于排查。

### 3.3 文档与实现同步

- 本文档以**当前代码实现**为准整理 Orm、Executor、Builder、Codec、Runner 的职责与数据流；若实现调整（如新增批量 Insert、软删除、多主键），应在本文档中同步更新“当前实现”与“操作流程”两节，并刷新“设计评估与完善建议”中受影响项。

---

## 4. 需改进项汇总（按当前实现评估）

以下为结合**当前设计**与**当前实现**的改进清单，便于按优先级落地。

### 4.1 设计/契约层面（见 3.2，保持与实现同步）

| 序号 | 内容 | 状态 |
|------|------|------|
| D1 | 入参校验与错误码：Orm 各方法对 nil 的约定（entity/filter → IllegalParam）在文档中写清 | **已成文** 见 2.6；实现已校验 entity/filter |
| D2 | Update 关系策略：引用关系按差异增量、包含关系按以新换旧 | **已成文** 见 2.6；**已实施**，详细方案与清单见 DESIGN-UPDATE-RELATION-DIFF.md |
| D3 | Query/Count/BatchQuery 无结果语义：Query 返回 NotFound、Count=0、BatchQuery=空切片 | **已成文** 见 2.6、2.7 |
| D4 | Filter 与 Model 对应关系、MaskModel 用途 | **已成文** 见 2.6 |
| D5 | 事务边界说明（单次操作事务 vs 调用方显式事务） | **已成文** 见 2.6 |
| D6 | Codec 与 Builder 职责边界成文 | **已成文** 见 2.6 |
| D7 | 关键路径日志规范（method、operation、model/filter、error） | **已成文** 见 2.6；实现见 4.2 |

### 4.2 实现层面（代码改进）

| 序号 | 位置 | 问题 | 建议 | 状态 |
|------|------|------|------|------|
| I1 | **orm 全包** | 大量 `slog.Error("operation failed", "error", ...)` 无 method/operation/model 上下文 | 与 provider 层一致：失败处打 slog 时带上方法名（如 Insert/Update、insertHost/buildInsert）、必要时 pkgKey/filter、以及 err，便于排查 | **已实施** |
| I2 | **orm/base.go** | `checkContext` / `CheckContext` 中 `slog.Error("message")` 无有效信息 | 改为如 `slog.Error("orm context invalid or cancelled")` 或带 "method","CheckContext","error", err.Error() | **已实施** |
| I3 | **orm/util.go** | `getModelFilter` 失败分支 `slog.Error("operation failed", ...)` | 改为带 "method","getModelFilter"、"error", err.Error()，必要时带 pkgKey | **已实施** |
| I4 | **orm/filter.go** | BatchQuery 在 filter==nil 时返回错误文案为 "illegal model value" | 改为 "illegal filter value" 或 "filter is nil"，与 Count 一致，错误码保持 IllegalParam | **已实施** |
| I5 | **orm/filter.go** | filter==nil 时 `slog.Error("message")` | 改为带 "method","BatchQuery","error", err.Error() 或明确 "filter is nil" | **已实施** |
| I6 | **orm/orm.go** | NewOrm / GetOrm / AddDatabase 等失败路径 "operation failed" | 改为带 "method"（NewOrm/GetOrm/AddDatabase）、"error", err.Error()；涉及连接池时带 owner 等标识 | **已实施** |
| I7 | **database 层** | postgres/mysql 的 executor/builder/util 中部分 slog 文案与 key 错误（如 "error", dsn、格式串未用） | 关键失败路径带 operation、dsn/field、error；修正 slog 键值对 | **已实施** |

### 4.3 实施记录（按第 4 节改进项落地情况）

- **日期**：按当前实现刷新。
- **I1**：orm 包内 create.go、drop.go、insert.go、update.go、delete.go、query.go、count.go、filter.go、util.go、orm.go、base.go 中所有原 `"operation failed"` 或 `"message"` 的 slog 已改为带方法/操作名、必要时 pkgKey/field、以及 error 的上下文日志（如 CreateRunner createHost BuildCreateTable failed、Insert InsertRunner.Insert failed、Query getModelFilter failed 等）。
- **I2**：base.go 中 checkContext / CheckContext 的 slog 已改为 "orm context invalid or cancelled" 与 "CheckContext: context invalid or cancelled"。
- **I3**：util.go getModelFilter 各失败分支已改为 getModelFilter GetModelFilter failed、getModelFilter PackedBasicFieldValue failed、getModelFilter Equal failed 等，并带 field、pkgKey、error。
- **I4、I5**：filter.go BatchQuery 在 filter==nil 时错误文案改为 "filter is nil"，slog 改为 "BatchQuery: filter is nil"；BatchQuery 调用 Query 失败时改为 "BatchQuery QueryRunner.Query failed"。
- **I6**：orm.go 中 AddDatabase、NewOrm、GetOrm、BeginTransaction、CommitTransaction、RollbackTransaction、finalTransaction 及 Insert 的失败 slog 已改为带方法名与 error（或 owner）；Create/Drop 的 impl 层失败改为带 pkgKey 与 error。
- **I7**：database 层已做修正：postgres/mysql executor 中 open/ping 日志改为带 dsn、error，去掉错误键值（"error", dsn）；Pool.connect 文案统一为 "Pool connect open/ping ..."；mysql executor Query 失败处改为 operation + sql + error；mysql builder/util 中 "%s" 格式串未用、键值错位已改为 "BuildInsert failed" 等 + field/operation/error。
- **D1～D7 成文**：新增 2.6「Orm 契约与实现约定」、2.7「操作语义与错误码速查」；4.1 设计层面状态更新为「已成文 见 2.6/2.7」。
- **database Warn**：postgres/mysql executor 中 Close/Rollback/Release 的 slog.Warn 已统一为 `"error", err.Error()`。

### 4.4 优先级建议

- **P0、P1**：已实施。
- **P2**：D1～D7 中尚未成文的在文档正文中补全（见第 5 节）。

---

## 5. 后续可完善项汇总

基于当前设计与实现，以下为**仍可继续完善**的内容（可选）。

| 类型 | 内容 | 状态 | 说明 |
|------|------|------|------|
| **文档** | D1～D7 成文 | **已实施** | 已在第 2.6 节「Orm 契约与实现约定」中写清入参 nil、Update 策略、无结果语义、Filter 与 Model、事务边界、Codec/Builder 职责、日志规范。 |
| **文档** | 操作语义与错误码速查 | **已实施** | 已在第 2.7 节「操作语义与错误码速查」中给出各操作成功/无结果/常见失败与错误码。 |
| **代码** | database 层 Warn 统一 | 已实施 | executor 中 Close/Rollback/Release 的 slog.Warn 已改为 `"error", err.Error()`，与 Error 路径风格一致。 |
| **功能** | Update 关系差异更新 | **已实施** | 已按 **DESIGN-UPDATE-RELATION-DIFF.md** 第 7 节清单落地（Builder、orm/update_diff.go、slice 语义、Query 选列）；测试见 test/update_relation_diff_local_test.go，符合性见 docs/UPDATE-TEST-DESIGN-COMPLIANCE.md。 |

当前**必须项**与 P2 文档成文、database Warn 统一、Update 关系差异更新均已完成。

---

## 6. 小结

- **数据库 ORM 操作**完全基于 **models.Model** 及 **models.Field / Type / Value / Filter**，通过 **provider.Provider** 做编解码与类型模型获取，与 **DESIGN-CONSISTENCY.md** 的模型抽象与数据一致性约定一致。
- **当前实现**的分层（Orm → Runner → Builder → Codec + Executor + Provider）清晰，表名与关系表名、字段值的读写均不绕过 Model/Field/Value 与 Provider。
- **评估结论**：设计符合“基于 Model 及相关定义实现数据库 ORM”的目标。**实现层面** I1～I7、database 层 Warn 统一、**Update 关系差异更新**（含 slice 语义与 Query 选列）已落地；**契约与操作语义**已在第 2.6、2.7 节成文；**改进项状态与实施记录**见第 4 节「需改进项汇总」；**后续可完善项**见第 5 节。

---

**相关文档**：模型抽象与数据一致性见 **DESIGN-CONSISTENCY.md**；本文档仅描述基于该抽象之上的数据库 ORM 操作设计。Update 关系差异更新与 slice 语义见 **DESIGN-UPDATE-RELATION-DIFF.md**（已实现并归档）；Query 选列与 slice 语义修复见 **docs/QUERY-SLICE-SEMANTICS-FIX.md**；Update 测试符合性见 **docs/UPDATE-TEST-DESIGN-COMPLIANCE.md**。
