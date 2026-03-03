# Update 关系按差异增量更新 — 实现方案（已实现并归档）

本文档描述将 **Update** 中关系字段从「先删后插」改为**按差异增量更新**的实现方案；**已按方案完成实现与测试，与当前代码一致，供归档与维护参考**。

---

## 1. 目标与范围

### 1.1 目标

- **当前行为**：`updateRelation` 对每个非基础字段先 `deleteRelation`（删除该 host 下该字段在关系表中的全部关联行，且对「包含关系」会级联删除关联实体），再 `insertRelation`（按当前 Model 中该字段的值逐条插入关系表）。
- **目标行为**：  
  - **引用关系**（Elem 为指针）：先查出当前关系表中该 host 的关联 right ID 列表，与本次要写入的「新 right ID 集合」做差集，仅**删除**需移除的链接、**插入**需新增的链接，不变的不动；对新集合中尚未落库的关联实体（无主键）先执行 Insert 再参与差集与插入。  
  - **包含关系**（Elem 非指针）：保持「用新关联实例替换旧关联实例」—— 即先删旧（关系行 + 旧关联实体）再插新（新关联实体 + 关系行），行为与当前一致，不做差集优化。

### 1.2 范围

- **在库**：关系字段的 Update 策略按**关系类型**分支（见 1.3）；引用关系仅做「链接级」差异更新，包含关系做「以新换旧」。
- **可选后续**：引用关系下对 toDelete 是否级联删除关联实体可做成可配置，本方案对引用关系仅做链接增删。

### 1.3 关系类型与 Update 策略（重要约定）

**核心规则**：引用关系只刷新关系（增删链接），不处理实体；包含关系要同步处理关系和实体（删旧关系+旧实体，插新实体+新关系）。

按字段类型的 **Elem() 是否指针**区分两种关系，Update 时采用不同策略：

| 关系类型 | 判定 | Update 策略 |
|----------|------|--------------|
| **引用关系** | `vField.GetType().Elem().IsPtrType() == true`（如 `*Child`、`[]*Role`） | **只刷新链接差异**：查当前 right ID 集合，与本次要写入的新 right ID 集合做差集，仅删除需移除的链接、仅插入需新增的链接；不删除关联实体本身。 |
| **包含关系** | `vField.GetType().Elem().IsPtrType() == false`（如 `Child`、`[]Tag`） | **用新关联实例替换旧关联实例**：先删旧（该 host 下该字段对应的关系表行 + 旧关联实体），再按新值插入新关联实体及关系表行；即保持现有「先 deleteRelation 再 insertRelation」的语义，不做链接级差集。 |

- **引用关系**：关联实体独立存在，Update 只维护「谁引用谁」；因此只做关系表的增删差集即可。
- **包含关系**：关联实体从属于 host，Update 视为「用新的一组实例替换旧的一组」；因此沿用先删后插，保证旧实例及其链接被清理、新实例及链接被写入。

下文第 3 节的差异增量更新流程**仅适用于引用关系**；包含关系在实现上仍走现有 `deleteRelation` + `insertRelation`，不再做差集优化。

### 1.4 属性类型为 slice 的赋值语义（与 R3/R7 清空关系一致）

- **nil**：视为**未赋值**（不参与 Update 关系写、Query 时从 DB 加载后仍会覆盖）。
- **[]**（空切片）：视为**已赋值、size 0**；会参与 Update 的 `updateRelation`，用于「清空关系」等场景（如引用单值清空、引用切片清空）。

实现上由 Value 层 `IsZero()` 对 slice 仅以 `IsNil()` 判断；MetaView 下 slice 重置为 nil。

**与 Query 的配合**：Query 时基础列选列与赋值规则与上述语义一致——仅「已赋值」或「值类型 slice」（如 `[]int`）参与 SELECT 与赋值，指针型未赋值不拉取，详见 **docs/QUERY-SLICE-SEMANTICS-FIX.md**。

---

## 2. 现状依赖（详细）

### 2.1 关系表结构

- 关系表包含三列：`id`（自增主键）、`left`（host 主键）、`right`（关联实体主键）。
- 已有 `BuildQueryRelation`（`SELECT right FROM relation WHERE left=?`）、`BuildInsertRelation(left, right)`、`BuildDeleteRelation`（先删关联实体再删关系表 `WHERE left=?`）。

### 2.2 获取当前链接

- `QueryRunner.innerQueryRelationKeys(vModel, vField)` 通过 `BuildQueryRelation` + `Query` + 遍历 `GetField` 得到当前 right ID 列表 `resultItems`（`[]any`）。
- 返回的 `resultItems` 中每个元素是数据库驱动返回的原始类型（如 `int64`、`string` 等），需注意与 Model 中主键值的类型可能不一致。

### 2.3 Update 入口

- `UpdateRunner.Update()` 先 `updateHost`，再对每个已赋值的非基础字段调用 `updateRelation`。
- `updateRelation` 当前为 `deleteRelation` + `SetValue(newVal)` + `insertRelation`。

### 2.4 InsertRunner 中的关系插入逻辑

当前 `insertSingleRelation` 和 `insertSliceRelation` 的核心行为：
- 对于引用关系（`IsPtrType() == true`）：**不**对关联实体执行 Insert（认为已存在于库中），直接 `BuildInsertRelation` 插入关系表行。
- 对于包含关系（`IsPtrType() == false`）：先对关联实体执行 `Insert`（创建实体），再 `BuildInsertRelation` 插入关系表行。

这一行为是差异更新设计的重要前提：引用关系下，新关联实体必须已存在于库中且有有效主键。

### 2.5 UpdateRunner 的 Runner 组合

```go
type UpdateRunner struct {
    baseRunner
    QueryRunner
    InsertRunner
    DeleteRunner
}
```

UpdateRunner 内嵌了 `QueryRunner`、`InsertRunner`、`DeleteRunner`，因此可以直接调用：
- `s.innerQueryRelationKeys(vModel, vField)` — 查询当前关系表中的 right ID 列表
- `s.insertRelation(vModel, vField)` — 插入关系（包含关系路径）
- `s.deleteRelation(vModel, vField, deepLevel)` — 删除关系（包含关系路径）
- `s.insertHost(rModel)` / `s.Insert()` — 插入关联实体（仅包含关系时需要）

---

## 3. 差异增量更新逻辑（详细设计）

**适用范围**：以下流程仅用于**引用关系**（`Elem().IsPtrType() == true`）。包含关系不进入本流程，仍按 1.3 使用「先删后插」替换。

### 3.1 总体流程（每个引用关系字段）

对单个**引用关系**字段，按以下顺序执行：

1. **准备「新 right ID 集合」**（保证新集合中所有项都有主键）
2. **查询「当前 right ID 集合」**（仅关系表）
3. **求差集**：`toDelete = 当前 - 新`，`toInsert = 新 - 当前`
4. **仅删链接**：对 `toDelete` 执行「只删关系表行」的 SQL
5. **仅插链接**：对 `toInsert` 执行「只插关系表行」的 SQL（不再调完整 insertRelation）

以下分别说明各步的入参、出参与对 Model/Builder/Executor 的依赖。

### 3.2 步骤 1：准备「新 right ID 集合」— prepareNewRightIDs

#### 3.2.1 函数签名

```go
func (s *UpdateRunner) prepareNewRightIDs(vModel models.Model, vField models.Field) (ret []any, err *cd.Error)
```

#### 3.2.2 单值关系（非 slice）的处理

```go
// vField 不是 slice 类型
elemType := vField.GetType().Elem()

// 1. 获取关联实体模型
rModel, rErr := s.modelProvider.GetTypeModel(elemType)
if rErr != nil {
    err = rErr
    slog.Error("prepareNewRightIDs GetTypeModel failed", "field", vField.GetName(), "error", err.Error())
    return
}

// 2. 将字段值设置到关联模型
rModel, rErr = s.modelProvider.SetModelValue(rModel, vField.GetValue())
if rErr != nil {
    err = rErr
    slog.Error("prepareNewRightIDs SetModelValue failed", "field", vField.GetName(), "error", err.Error())
    return
}

// 3. 提取主键值
pkField := rModel.GetPrimaryField()
pkValue := pkField.GetValue()
if pkValue.IsZero() {
    // 引用关系下，关联实体的主键不应为空
    // 引用关系的语义是"引用已存在的实体"，主键为空说明数据不合法
    err = cd.NewError(cd.IllegalParam, fmt.Sprintf("reference relation field %s has entity without primary key", vField.GetName()))
    slog.Error("prepareNewRightIDs: reference entity missing primary key", "field", vField.GetName())
    return
}

// 4. 将主键编码为 DB 可比较格式
encodedPK, encErr := s.modelCodec.PackedBasicFieldValue(pkField, pkValue)
if encErr != nil {
    err = encErr
    slog.Error("prepareNewRightIDs PackedBasicFieldValue failed", "field", vField.GetName(), "error", err.Error())
    return
}

ret = []any{encodedPK}
```

#### 3.2.3 切片关系（slice）的处理

```go
// vField 是 slice 类型
fSliceValue := vField.GetSliceValue()
if len(fSliceValue) == 0 {
    ret = []any{}
    return
}

elemType := vField.GetType().Elem()
newRightIDs := make([]any, 0, len(fSliceValue))

for _, fVal := range fSliceValue {
    rModel, rErr := s.modelProvider.GetTypeModel(elemType)
    if rErr != nil {
        err = rErr
        slog.Error("prepareNewRightIDs GetTypeModel failed", "field", vField.GetName(), "error", err.Error())
        return
    }
    rModel, rErr = s.modelProvider.SetModelValue(rModel, fVal)
    if rErr != nil {
        err = rErr
        slog.Error("prepareNewRightIDs SetModelValue failed", "field", vField.GetName(), "error", err.Error())
        return
    }

    pkField := rModel.GetPrimaryField()
    pkValue := pkField.GetValue()
    if pkValue.IsZero() {
        err = cd.NewError(cd.IllegalParam, fmt.Sprintf("reference relation field %s has entity without primary key", vField.GetName()))
        slog.Error("prepareNewRightIDs: reference entity missing primary key", "field", vField.GetName())
        return
    }

    encodedPK, encErr := s.modelCodec.PackedBasicFieldValue(pkField, pkValue)
    if encErr != nil {
        err = encErr
        slog.Error("prepareNewRightIDs PackedBasicFieldValue failed", "field", vField.GetName(), "error", err.Error())
        return
    }
    newRightIDs = append(newRightIDs, encodedPK)
}

ret = newRightIDs
```

#### 3.2.4 关于「引用关系下主键为空」的设计决策

**约定**：引用关系（Elem 为指针）语义上表示"引用已存在的外部实体"。在 Update 场景下，如果引用的实体没有主键，视为非法入参，返回 `cd.IllegalParam`。

**理由**：
1. `insertSingleRelation` / `insertSliceRelation` 中对引用关系（`IsPtrType() == true`）也**不**执行关联实体的 Insert，只插关系表行——这说明现有设计已假设引用关系下的关联实体必须已存在。
2. 若允许 Update 时自动 Insert 缺少主键的引用实体，会模糊「引用」与「包含」的语义边界。
3. 调用方有责任在 Update 前确保引用的实体已落库。

如果后续需要支持「Update 时自动为无主键的引用实体执行 Insert」，可在此处增加分支，但需在文档中明确标注为语义扩展。

### 3.3 步骤 2：查询「当前 right ID 集合」— queryExistingRightIDs

#### 3.3.1 函数签名

```go
func (s *UpdateRunner) queryExistingRightIDs(vModel models.Model, vField models.Field) (ret []any, err *cd.Error)
```

#### 3.3.2 实现

直接复用 `innerQueryRelationKeys`：

```go
func (s *UpdateRunner) queryExistingRightIDs(vModel models.Model, vField models.Field) (ret []any, err *cd.Error) {
    existingIDs, queryErr := s.innerQueryRelationKeys(vModel, vField)
    if queryErr != nil {
        err = queryErr
        slog.Error("queryExistingRightIDs innerQueryRelationKeys failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }
    ret = existingIDs
    return
}
```

#### 3.3.3 注意事项

- `innerQueryRelationKeys` 返回的是数据库驱动层的原始值（如 `int64`），而 `prepareNewRightIDs` 返回的是经 `PackedBasicFieldValue` 编码后的值。
- **两者的类型可能不一致**（如 Go 的 `int` 经编码后为 `int64`，而 DB 返回也是 `int64`，这种情况可对齐；但 `string` UUID 等需确认一致性）。
- 差集比较时需统一规范化，详见 3.4。

### 3.4 步骤 3：求差集 — diffRelationIDs

#### 3.4.1 函数签名

```go
func diffRelationIDs(existing, new []any) (toDelete, toInsert []any)
```

#### 3.4.2 主键判等策略

**核心问题**：`existing` 来自 DB 驱动的 `GetField(&idVal)`（返回 `any`，实际类型由 DB 驱动决定，如 PostgreSQL 返回 `int64`），`new` 来自 `PackedBasicFieldValue`（经 Provider.EncodeValue 编码后的值）。两者的底层类型**可能相同也可能不同**。

**推荐方案：统一转为字符串比较**

```go
func normalizeID(id any) string {
    return fmt.Sprintf("%v", id)
}

func diffRelationIDs(existing, new []any) (toDelete, toInsert []any) {
    existingSet := make(map[string]any, len(existing))
    for _, id := range existing {
        existingSet[normalizeID(id)] = id
    }

    newSet := make(map[string]any, len(new))
    for _, id := range new {
        newSet[normalizeID(id)] = id
    }

    // toDelete = existing - new
    for key, id := range existingSet {
        if _, found := newSet[key]; !found {
            toDelete = append(toDelete, id)
        }
    }

    // toInsert = new - existing
    for key, id := range newSet {
        if _, found := existingSet[key]; !found {
            toInsert = append(toInsert, id)
        }
    }

    return
}
```

#### 3.4.3 为何选择字符串化

| 方案 | 优点 | 缺点 |
|------|------|------|
| `fmt.Sprintf("%v", id)` | 简单通用，int/int64/string/uuid 均可；与 DB 读写的文本表示一致 | 对浮点数可能有精度差异（但 ORM 主键不应是浮点） |
| `reflect.DeepEqual` | 精确比较 | 类型不完全一致时（如 int vs int64）会判定不等，导致误删误插 |
| 自定义 Codec 规范化接口 | 最严谨 | 需扩展 Codec 接口，改动范围大，当前可不必要 |

**结论**：鉴于 ORM 主键类型通常为 int/int64/string/uuid，`fmt.Sprintf("%v")` 足够可靠。后续若引入非常规主键类型，可扩展为自定义规范化。

#### 3.4.4 去重保证

- `prepareNewRightIDs` 不做去重（调用方可能传入重复引用）。
- `diffRelationIDs` 内部使用 map，天然去重。若 `new` 中有重复 ID，差集结果中只会出现一次。
- `existing` 正常情况不应有重复（关系表由 ORM 维护），但 map 也能自然处理。

### 3.5 步骤 4：仅删关系表行（不删关联实体）— deleteRelationLinks

#### 3.5.1 设计思路

当前 `BuildDeleteRelation` 返回两个 Result：`delHost`（删除关联实体）和 `delRelation`（删除关系表行）。对于引用关系的差异更新，我们**只需要删除关系表中特定的 right 行**，不删除关联实体。

需要新增一个 Builder 方法：`BuildDeleteRelationByRights`。

#### 3.5.2 SQL 形态

PostgreSQL：
```sql
DELETE FROM "RelationTable" WHERE "left"=$1 AND "right" IN ($2,$3,...)
```

MySQL：
```sql
DELETE FROM `RelationTable` WHERE `left`=? AND `right` IN (?,?,...)
```

#### 3.5.3 rightIDs 为空时的行为约定

- 若 `rightIDs` 为空或 nil，方法返回 `(nil, nil)`。
- 调用方在收到 nil Result 时跳过 Execute，不执行任何 SQL。

#### 3.5.4 UpdateRunner 中的调用

```go
func (s *UpdateRunner) deleteRelationLinks(vModel models.Model, vField models.Field, toDelete []any) (err *cd.Error) {
    if len(toDelete) == 0 {
        return
    }

    result, buildErr := s.sqlBuilder.BuildDeleteRelationByRights(vModel, vField, toDelete)
    if buildErr != nil {
        err = buildErr
        slog.Error("deleteRelationLinks BuildDeleteRelationByRights failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }

    if result == nil {
        return
    }

    _, err = s.executor.Execute(result.SQL(), result.Args()...)
    if err != nil {
        slog.Error("deleteRelationLinks Execute failed",
            "field", vField.GetName(), "error", err.Error())
    }
    return
}
```

### 3.6 步骤 5：仅插关系表行 — insertRelationLinks

#### 3.6.1 设计思路

对 `toInsert` 中的每个 rightID，需要构造一个只包含主键的 rModel，然后调用已有的 `BuildInsertRelation` 生成 INSERT SQL。

#### 3.6.2 实现

```go
func (s *UpdateRunner) insertRelationLinks(vModel models.Model, vField models.Field, toInsert []any) (err *cd.Error) {
    if len(toInsert) == 0 {
        return
    }

    elemType := vField.GetType().Elem()
    for _, rightID := range toInsert {
        rModel, rErr := s.modelProvider.GetTypeModel(elemType)
        if rErr != nil {
            err = rErr
            slog.Error("insertRelationLinks GetTypeModel failed",
                "field", vField.GetName(), "error", err.Error())
            return
        }

        // 从 DB 原始值解码为模型可接受的主键值
        rVal, rErr := s.modelCodec.ExtractBasicFieldValue(rModel.GetPrimaryField(), rightID)
        if rErr != nil {
            err = rErr
            slog.Error("insertRelationLinks ExtractBasicFieldValue failed",
                "field", vField.GetName(), "error", err.Error())
            return
        }
        rModel.SetPrimaryFieldValue(rVal)

        relationResult, relationErr := s.sqlBuilder.BuildInsertRelation(vModel, vField, rModel)
        if relationErr != nil {
            err = relationErr
            slog.Error("insertRelationLinks BuildInsertRelation failed",
                "field", vField.GetName(), "error", err.Error())
            return
        }

        var idVal any
        err = s.executor.ExecuteInsert(relationResult.SQL(), &idVal, relationResult.Args()...)
        if err != nil {
            slog.Error("insertRelationLinks ExecuteInsert failed",
                "field", vField.GetName(), "error", err.Error())
            return
        }
    }
    return
}
```

#### 3.6.3 关于 toInsert 中 rightID 的来源

`toInsert` 来自 `diffRelationIDs`，其中的值有两种来源：
- 来自 `newSet`（由 `prepareNewRightIDs` 产生，经 `PackedBasicFieldValue` 编码）
- 来自 `existingSet`（由 `queryExistingRightIDs` 产生，即 DB 驱动原始值）

由于 `toInsert = new - existing`，toInsert 中的值来自 `newSet`（即经 PackedBasicFieldValue 编码后的值）。但 `BuildInsertRelation` 内部直接取 `rModel.GetPrimaryField().GetValue().Get()` 作为参数，所以我们需要通过 `ExtractBasicFieldValue` 将编码后的值解码回模型可接受的值，再通过 `SetPrimaryFieldValue` 设置到 rModel。

**但更简洁的做法**：直接用 `toInsert` 中的值作为 SQL 参数，而不经过 rModel 中转。这需要一个新的 Builder 方法 `BuildInsertRelationByRight`。

**权衡后决定**：保持通过 rModel 中转的方式，复用已有的 `BuildInsertRelation`，避免引入更多新接口。虽然多一次解码/编码，但逻辑清晰、改动最小。

### 3.7 单值关系与切片关系统一

- 上述 1～5 对「单值」与「切片」的区别仅体现在：  
  - 步骤 1 中，单值得到 0 或 1 个 newRightID，切片得到 N 个；  
  - 步骤 3 的 `diffRelationIDs` 对 0/1/N 均适用。  
- 单值可视为「长度为 0 或 1 的 newRightIDs」，与现有 `existingRightIDs`（0 或 1）做差集后，toDelete/toInsert 最多各 1 个，逻辑一致。

### 3.8 引用关系下 toDelete 与关联实体

- **本阶段**：引用关系的差异更新只做「链接的增删」；从集合中移除的 right（toDelete）**仅删关系表行**，不删关联实体（不调用当前 delHost），关联实体仍保留在库中。
- **可选后续**：若业务需要「解除引用时顺带删除被引用的实体」，可在本方案落地后增加配置或单独分支，对 toDelete 再执行实体删除；默认不做。

---

## 4. 接口与层职责

### 4.1 database.Builder 新增

#### 4.1.1 接口定义

在 `database/builder.go` 的 `Builder` 接口中新增：

```go
BuildDeleteRelationByRights(vModel models.Model, vField models.Field, rightIDs []any) (Result, *cd.Error)
```

#### 4.1.2 行为约定

- 生成 `DELETE FROM <relation_table> WHERE left=? AND right IN (...)`，参数为 host 主键 + rightIDs。
- 若 `rightIDs` 为空或 nil，返回 `(nil, nil)`，调用方不执行。
- 表名通过现有 `ConstructRelationTableName(vModel, vField)` 获取。

#### 4.1.3 PostgreSQL 实现

```go
func (s *Builder) BuildDeleteRelationByRights(vModel models.Model, vField models.Field, rightIDs []any) (ret database.Result, err *cd.Error) {
    if len(rightIDs) == 0 {
        return
    }

    hostVal := vModel.GetPrimaryField().GetValue().Get()
    relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
    if relationErr != nil {
        err = relationErr
        slog.Error("BuildDeleteRelationByRights ConstructRelationTableName failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }

    resultStackPtr := &ResultStack{}
    resultStackPtr.PushArgs(hostVal)

    // 构造 IN 列表的占位符
    inPlaceholders := ""
    for _, rightID := range rightIDs {
        resultStackPtr.PushArgs(rightID)
        if inPlaceholders == "" {
            inPlaceholders = fmt.Sprintf("$%d", len(resultStackPtr.argsVal))
        } else {
            inPlaceholders = fmt.Sprintf("%s,$%d", inPlaceholders, len(resultStackPtr.argsVal))
        }
    }

    deleteSQL := fmt.Sprintf("DELETE FROM \"%s\" WHERE \"left\"=$1 AND \"right\" IN (%s)",
        relationTableName, inPlaceholders)

    if traceSQL() {
        slog.Info("[SQL] delete relation by rights", "sql", deleteSQL)
    }

    resultStackPtr.SetSQL(deleteSQL)
    ret = resultStackPtr
    return
}
```

#### 4.1.4 MySQL 实现

```go
func (s *Builder) BuildDeleteRelationByRights(vModel models.Model, vField models.Field, rightIDs []any) (ret database.Result, err *cd.Error) {
    if len(rightIDs) == 0 {
        return
    }

    hostVal := vModel.GetPrimaryField().GetValue().Get()
    relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
    if relationErr != nil {
        err = relationErr
        slog.Error("BuildDeleteRelationByRights ConstructRelationTableName failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }

    resultStackPtr := &ResultStack{}
    resultStackPtr.PushArgs(hostVal)

    // 构造 IN 列表的占位符
    inPlaceholders := ""
    for _, rightID := range rightIDs {
        resultStackPtr.PushArgs(rightID)
        if inPlaceholders == "" {
            inPlaceholders = "?"
        } else {
            inPlaceholders = fmt.Sprintf("%s,?", inPlaceholders)
        }
    }

    deleteSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `left`=? AND `right` IN (%s)",
        relationTableName, inPlaceholders)

    if traceSQL() {
        slog.Info("[SQL] delete relation by rights", "sql", deleteSQL)
    }

    resultStackPtr.SetSQL(deleteSQL)
    ret = resultStackPtr
    return
}
```

### 4.2 orm 层（UpdateRunner）改造

#### 4.2.1 updateRelation 分支入口

```go
func (s *UpdateRunner) updateRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
    elemType := vField.GetType().Elem()

    if elemType.IsPtrType() {
        // 引用关系：走差异增量更新
        err = s.updateReferenceRelation(vModel, vField)
        if err != nil {
            slog.Error("UpdateRunner updateRelation updateReferenceRelation failed",
                "field", vField.GetName(), "error", err.Error())
        }
        return
    }

    // 包含关系：保持先删后插（以新换旧）
    err = s.updateContainRelation(vModel, vField)
    if err != nil {
        slog.Error("UpdateRunner updateRelation updateContainRelation failed",
            "field", vField.GetName(), "error", err.Error())
    }
    return
}
```

#### 4.2.2 包含关系的 updateContainRelation（保持现有逻辑）

```go
func (s *UpdateRunner) updateContainRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
    newVal := vField.GetValue().Get()
    err = s.deleteRelation(vModel, vField, 0)
    if err != nil {
        slog.Error("UpdateRunner updateContainRelation deleteRelation failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }

    vField.SetValue(newVal)
    err = s.insertRelation(vModel, vField)
    if err != nil {
        slog.Error("UpdateRunner updateContainRelation insertRelation failed",
            "field", vField.GetName(), "error", err.Error())
    }
    return
}
```

#### 4.2.3 引用关系的 updateReferenceRelation（差异增量）

```go
func (s *UpdateRunner) updateReferenceRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
    // 步骤 1：准备新 right ID 集合
    newRightIDs, prepErr := s.prepareNewRightIDs(vModel, vField)
    if prepErr != nil {
        err = prepErr
        slog.Error("UpdateRunner updateReferenceRelation prepareNewRightIDs failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }

    // 步骤 2：查询当前 right ID 集合
    existingRightIDs, queryErr := s.queryExistingRightIDs(vModel, vField)
    if queryErr != nil {
        err = queryErr
        slog.Error("UpdateRunner updateReferenceRelation queryExistingRightIDs failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }

    // 步骤 3：求差集
    toDelete, toInsert := diffRelationIDs(existingRightIDs, newRightIDs)

    // 步骤 4：删除需移除的链接
    err = s.deleteRelationLinks(vModel, vField, toDelete)
    if err != nil {
        slog.Error("UpdateRunner updateReferenceRelation deleteRelationLinks failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }

    // 步骤 5：插入需新增的链接
    err = s.insertRelationLinks(vModel, vField, toInsert)
    if err != nil {
        slog.Error("UpdateRunner updateReferenceRelation insertRelationLinks failed",
            "field", vField.GetName(), "error", err.Error())
    }
    return
}
```

### 4.3 主键判等与 Codec

- **判等**：diff 时需对两个 `any`（主键值）判等。使用 `fmt.Sprintf("%v", id)` 统一为字符串后比较。
- **文档约定**：差异比较使用字符串化的可比较表示，适用于常见主键类型（int/int64/string/uuid）；若后续引入非常规主键类型，需在此处扩展 `normalizeID` 函数。

### 4.4 新增函数文件组织

建议在 `orm` 包中新增 `update_diff.go` 文件，包含以下函数：
- `normalizeID(id any) string`
- `diffRelationIDs(existing, new []any) (toDelete, toInsert []any)`
- `(s *UpdateRunner) prepareNewRightIDs(vModel, vField) ([]any, *cd.Error)`
- `(s *UpdateRunner) queryExistingRightIDs(vModel, vField) ([]any, *cd.Error)`
- `(s *UpdateRunner) deleteRelationLinks(vModel, vField, toDelete []any) *cd.Error`
- `(s *UpdateRunner) insertRelationLinks(vModel, vField, toInsert []any) *cd.Error`
- `(s *UpdateRunner) updateReferenceRelation(vModel, vField) *cd.Error`
- `(s *UpdateRunner) updateContainRelation(vModel, vField) *cd.Error`

原 `update.go` 中的 `updateRelation` 改为分支调度（见 4.2.1），其余保持不变。

---

## 5. 边界与兼容

### 5.1 边界情况

| 场景 | existing | new | toDelete | toInsert | 行为 |
|------|----------|-----|----------|----------|------|
| 关系表无行，新增全部 | `[]` | `[A, B]` | `[]` | `[A, B]` | 仅执行两次 InsertRelation |
| 清空全部引用 | `[A, B]` | `[]` | `[A, B]` | `[]` | 执行 BuildDeleteRelationByRights 删 A, B |
| 完全一致 | `[A, B]` | `[A, B]` | `[]` | `[]` | 不执行任何 SQL |
| 部分替换 | `[A, B]` | `[B, C]` | `[A]` | `[C]` | 删 A 的链接，插 C 的链接 |
| 单值引用替换 | `[A]` | `[B]` | `[A]` | `[B]` | 删 A 的链接，插 B 的链接 |
| 单值引用不变 | `[A]` | `[A]` | `[]` | `[]` | 不执行任何 SQL |
| 新集合有重复 | `[A]` | `[B, B]` | `[A]` | `[B]` | map 去重，只插一次 B |

### 5.2 字段未赋值时的行为

- `Update()` 在遍历字段时已通过 `!models.IsAssignedField(field)` 过滤掉未赋值的字段。
- 未赋值的关系字段**不会进入** `updateRelation`，关系表中该字段的链接保持不变。
- 这与当前行为一致。

### 5.3 兼容与回退

- **行为兼容**：对外 API 仍为 `Update(vModel)`，仅内部 `updateRelation` 分支处理。对「结果状态」等价（最终关系表与 host 的链接集合 = newRightIDs）。
- **回退**：若需回退，将 `updateRelation` 恢复为当前统一的「先删后插」实现，保留 `BuildDeleteRelationByRights` 与 diff 工具函数供后续使用。

### 5.4 事务与错误

- Update 仍在现有事务中执行（impl 层 BeginTransaction + defer Commit/Rollback）；上述所有步骤在同一事务内，失败即回滚，与现有一致。
- 步骤 1～5 中任何一步失败，立即返回错误，事务回滚，不会出现「部分链接已删但新链接未插」的不一致状态。

### 5.5 并发安全

- 多个并发 Update 同一 host 的同一字段时，可能出现「查 existing 后、执行删/插前」被另一事务修改的情况。
- 由于 Update 在事务中执行，数据库层面的行锁/事务隔离级别（READ COMMITTED 或更高）保证最终一致性。
- 极端并发下可能出现「关系表中出现重复行」（两个事务同时插入相同的 `(left, right)`），可通过在关系表上添加 `UNIQUE(left, right)` 约束来防止（此为可选增强，不在本方案范围内）。

---

## 6. 测试策略

### 6.1 单元测试

#### 6.1.1 diffRelationIDs 测试

```go
func TestDiffRelationIDs(t *testing.T) {
    tests := []struct {
        name       string
        existing   []any
        new        []any
        wantDelete []any
        wantInsert []any
    }{
        {"both empty", nil, nil, nil, nil},
        {"existing empty, new has items", nil, []any{1, 2}, nil, []any{1, 2}},
        {"new empty, existing has items", []any{1, 2}, nil, []any{1, 2}, nil},
        {"identical sets", []any{1, 2}, []any{1, 2}, nil, nil},
        {"partial overlap", []any{1, 2}, []any{2, 3}, []any{1}, []any{3}},
        {"complete replacement", []any{1, 2}, []any{3, 4}, []any{1, 2}, []any{3, 4}},
        {"new has duplicates", []any{1}, []any{2, 2}, []any{1}, []any{2}},
        {"string IDs", []any{"a", "b"}, []any{"b", "c"}, []any{"a"}, []any{"c"}},
        {"mixed int types", []any{int64(1), int64(2)}, []any{int64(2), int64(3)}, []any{int64(1)}, []any{int64(3)}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            gotDelete, gotInsert := diffRelationIDs(tt.existing, tt.new)
            // 比较时需考虑顺序无关
            assertSameSet(t, "toDelete", tt.wantDelete, gotDelete)
            assertSameSet(t, "toInsert", tt.wantInsert, gotInsert)
        })
    }
}
```

#### 6.1.2 normalizeID 测试

```go
func TestNormalizeID(t *testing.T) {
    tests := []struct {
        input    any
        expected string
    }{
        {int64(42), "42"},
        {int(42), "42"},
        {"uuid-string-123", "uuid-string-123"},
        {int64(0), "0"},
    }

    for _, tt := range tests {
        got := normalizeID(tt.input)
        if got != tt.expected {
            t.Errorf("normalizeID(%v) = %q, want %q", tt.input, got, tt.expected)
        }
    }
}
```

### 6.2 集成测试

需要覆盖以下场景（以 PostgreSQL 和 MySQL 双数据库测试）：

#### 6.2.1 引用关系测试模型

```go
type Author struct {
    ID   int    `orm:"id key auto"`
    Name string `orm:"name"`
}

type Tag struct {
    ID   int    `orm:"id key auto"`
    Name string `orm:"name"`
}

type Article struct {
    ID     int     `orm:"id key auto"`
    Title  string  `orm:"title"`
    Author *Author `orm:"author"`     // 引用关系 (单值)
    Tags   []*Tag  `orm:"tags"`       // 引用关系 (切片)
}
```

#### 6.2.2 测试用例

| 测试 | 操作 | 验证 |
|------|------|------|
| **引用单值：新增** | Insert Article(Author=nil) → Update Article(Author=Author1) | 关系表新增 1 行；Author1 实体不变 |
| **引用单值：替换** | Insert Article(Author=Author1) → Update Article(Author=Author2) | 关系表：删 Author1 行，插 Author2 行；Author1 实体仍在库中 |
| **引用单值：清空** | Insert Article(Author=Author1) → Update Article(Author=nil) | 关系表删 Author1 行；Author1 实体仍在库中 |
| **引用单值：不变** | Insert Article(Author=Author1) → Update Article(Author=Author1) | 关系表不变 |
| **引用切片：新增** | Insert Article(Tags=[]) → Update Article(Tags=[T1,T2]) | 关系表新增 2 行 |
| **引用切片：部分替换** | Insert Article(Tags=[T1,T2]) → Update Article(Tags=[T2,T3]) | 关系表删 T1 行，插 T3 行；T1 实体仍在库中 |
| **引用切片：清空** | Insert Article(Tags=[T1,T2]) → Update Article(Tags=[]) | 关系表删 2 行；T1, T2 实体仍在库中 |
| **引用切片：完全相同** | Insert Article(Tags=[T1,T2]) → Update Article(Tags=[T1,T2]) | 不执行任何 SQL |
| **包含关系：仍为先删后插** | Insert Article(Desc=Desc1) → Update Article(Desc=Desc2) | 旧 Desc1 实体被删除，新 Desc2 实体被创建 |

#### 6.2.3 验证方法

每个测试用例验证三个层面：
1. **返回值**：Update 返回的 Model 数据正确。
2. **关系表**：通过 Query 验证关系表中的链接数据与预期一致。
3. **关联实体表**：通过 Query 验证引用关系下被解除引用的实体仍存在于库中（不被删除）。

### 6.3 性能测试（可选）

针对大切片引用关系（如 100+ Tags），对比改造前后的性能：
- **改造前**：先 DELETE N 行 + INSERT N 行 = 2N 次写操作
- **改造后**：若只变更 1 个 Tag，仅 1 次 DELETE + 1 次 INSERT = 2 次写操作

预期在「少量变更」场景下有显著性能提升。

---

## 7. 实现清单（已实施）

以下清单已全部落地，与当前代码一致：

| 序号 | 项 | 文件 | 说明 | 状态 |
|-----|----|------|------|------|
| 1 | **Builder 接口** | `database/builder.go` | 在 `Builder` 增加 `BuildDeleteRelationByRights(vModel, vField, rightIDs []any) (Result, *cd.Error)` | **已实施** |
| 2 | **Postgres Builder** | `database/postgres/builder_delete.go` | 实现 `BuildDeleteRelationByRights`：`DELETE FROM "relation" WHERE "left"=$1 AND "right" IN ($2,$3,...)`，rightIDs 为空时返回 nil | **已实施** |
| 3 | **MySQL Builder** | `database/mysql/builder_delete.go` | 同上，占位符为 `?` | **已实施** |
| 4 | **新增 update_diff.go** | `orm/update_diff.go` | 包含 `normalizeID`、`diffRelationIDs`、`prepareNewRightIDs`、`queryExistingRightIDs`、`deleteRelationLinks`、`insertRelationLinks`、`updateReferenceRelation`、`updateContainRelation` | **已实施** |
| 5 | **改造 updateRelation** | `orm/update.go` | `updateRelation` 按 `Elem().IsPtrType()` 分支：引用关系走 `updateReferenceRelation`，包含关系走 `updateContainRelation` | **已实施** |
| 6 | **单元测试** | `orm/update_diff_test.go` | `diffRelationIDs`、`normalizeID` 的单元测试 | **已实施** |
| 7 | **集成测试** | `test/update_relation_diff_local_test.go` | 引用关系/包含关系 Update 集成测试（R1～R8、C1），含 R3/R7 清空关系断言 | **已实施** |
| 8 | **文档** | `DESIGN-DATABASE-ORM.md` | 2.6「Update 关系更新策略」已补充引用关系按差异增量、包含关系按以新换旧，并引用本文 | **已实施** |
| 9 | **slice 语义与 Query** | `provider/local/value.go`、`field.go`；`database/*/builder_query.go`、`orm/query.go` | slice：nil=未赋值、[]=已赋值；MetaView 下 slice 重置为 nil；Query 选列仅「已赋值」或「值类型 slice」参与 SELECT/赋值，见 docs/QUERY-SLICE-SEMANTICS-FIX.md | **已实施** |

---

## 8. 与现有代码的详细变更对照

### 8.1 `database/builder.go` 变更

```diff
 type Builder interface {
     // ... 现有方法 ...
     BuildDeleteRelation(vModel models.Model, vField models.Field) (Result, Result, *cd.Error)
     BuildQueryRelation(vModel models.Model, vField models.Field) (Result, *cd.Error)
+    BuildDeleteRelationByRights(vModel models.Model, vField models.Field, rightIDs []any) (Result, *cd.Error)
 
     BuildModuleValueHolder(vModel models.Model) ([]any, *cd.Error)
 }
```

### 8.2 `orm/update.go` 变更

```diff
 func (s *UpdateRunner) updateRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
-    newVal := vField.GetValue().Get()
-    err = s.deleteRelation(vModel, vField, 0)
-    if err != nil {
-        slog.Error("UpdateRunner updateRelation deleteRelation failed", "field", vField.GetName(), "error", err.Error())
-        return
-    }
-    // TODO 这里最合理的逻辑应该是先查询出当前值，与新值进行差异比较
-    // 再根据比较后的结果进行处理
-    // 目前先粗暴点，直接删除再插入
-    vField.SetValue(newVal)
-    err = s.insertRelation(vModel, vField)
-    if err != nil {
-        slog.Error("UpdateRunner updateRelation insertRelation failed", "field", vField.GetName(), "error", err.Error())
-    }
-    return
+    elemType := vField.GetType().Elem()
+    if elemType.IsPtrType() {
+        err = s.updateReferenceRelation(vModel, vField)
+        if err != nil {
+            slog.Error("UpdateRunner updateRelation updateReferenceRelation failed",
+                "field", vField.GetName(), "error", err.Error())
+        }
+        return
+    }
+    err = s.updateContainRelation(vModel, vField)
+    if err != nil {
+        slog.Error("UpdateRunner updateRelation updateContainRelation failed",
+            "field", vField.GetName(), "error", err.Error())
+    }
+    return
 }
```

### 8.3 新增 `orm/update_diff.go`（完整伪代码骨架）

```go
package orm

import (
    "fmt"
    "log/slog"

    cd "github.com/muidea/magicCommon/def"
    "github.com/muidea/magicOrm/models"
)

func normalizeID(id any) string {
    return fmt.Sprintf("%v", id)
}

func diffRelationIDs(existing, new []any) (toDelete, toInsert []any) {
    existingSet := make(map[string]any, len(existing))
    for _, id := range existing {
        existingSet[normalizeID(id)] = id
    }

    newSet := make(map[string]any, len(new))
    for _, id := range new {
        newSet[normalizeID(id)] = id
    }

    for key, id := range existingSet {
        if _, found := newSet[key]; !found {
            toDelete = append(toDelete, id)
        }
    }

    for key, id := range newSet {
        if _, found := existingSet[key]; !found {
            toInsert = append(toInsert, id)
        }
    }

    return
}

func (s *UpdateRunner) prepareNewRightIDs(vModel models.Model, vField models.Field) (ret []any, err *cd.Error) {
    elemType := vField.GetType().Elem()

    if models.IsSliceField(vField) {
        fSliceValue := vField.GetSliceValue()
        if len(fSliceValue) == 0 {
            ret = []any{}
            return
        }

        newRightIDs := make([]any, 0, len(fSliceValue))
        for _, fVal := range fSliceValue {
            rModel, rErr := s.modelProvider.GetTypeModel(elemType)
            if rErr != nil {
                err = rErr
                slog.Error("prepareNewRightIDs GetTypeModel failed",
                    "field", vField.GetName(), "error", err.Error())
                return
            }
            rModel, rErr = s.modelProvider.SetModelValue(rModel, fVal)
            if rErr != nil {
                err = rErr
                slog.Error("prepareNewRightIDs SetModelValue failed",
                    "field", vField.GetName(), "error", err.Error())
                return
            }

            pkField := rModel.GetPrimaryField()
            pkValue := pkField.GetValue()
            if pkValue.IsZero() {
                err = cd.NewError(cd.IllegalParam,
                    fmt.Sprintf("reference relation field %s has entity without primary key", vField.GetName()))
                slog.Error("prepareNewRightIDs: reference entity missing primary key",
                    "field", vField.GetName())
                return
            }

            encodedPK, encErr := s.modelCodec.PackedBasicFieldValue(pkField, pkValue)
            if encErr != nil {
                err = encErr
                slog.Error("prepareNewRightIDs PackedBasicFieldValue failed",
                    "field", vField.GetName(), "error", err.Error())
                return
            }
            newRightIDs = append(newRightIDs, encodedPK)
        }
        ret = newRightIDs
        return
    }

    // 单值关系
    rModel, rErr := s.modelProvider.GetTypeModel(elemType)
    if rErr != nil {
        err = rErr
        slog.Error("prepareNewRightIDs GetTypeModel failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }
    rModel, rErr = s.modelProvider.SetModelValue(rModel, vField.GetValue())
    if rErr != nil {
        err = rErr
        slog.Error("prepareNewRightIDs SetModelValue failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }

    pkField := rModel.GetPrimaryField()
    pkValue := pkField.GetValue()
    if pkValue.IsZero() {
        err = cd.NewError(cd.IllegalParam,
            fmt.Sprintf("reference relation field %s has entity without primary key", vField.GetName()))
        slog.Error("prepareNewRightIDs: reference entity missing primary key",
            "field", vField.GetName())
        return
    }

    encodedPK, encErr := s.modelCodec.PackedBasicFieldValue(pkField, pkValue)
    if encErr != nil {
        err = encErr
        slog.Error("prepareNewRightIDs PackedBasicFieldValue failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }
    ret = []any{encodedPK}
    return
}

func (s *UpdateRunner) queryExistingRightIDs(vModel models.Model, vField models.Field) (ret []any, err *cd.Error) {
    existingIDs, queryErr := s.innerQueryRelationKeys(vModel, vField)
    if queryErr != nil {
        err = queryErr
        slog.Error("queryExistingRightIDs innerQueryRelationKeys failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }
    ret = existingIDs
    return
}

func (s *UpdateRunner) deleteRelationLinks(vModel models.Model, vField models.Field, toDelete []any) (err *cd.Error) {
    if len(toDelete) == 0 {
        return
    }

    result, buildErr := s.sqlBuilder.BuildDeleteRelationByRights(vModel, vField, toDelete)
    if buildErr != nil {
        err = buildErr
        slog.Error("deleteRelationLinks BuildDeleteRelationByRights failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }
    if result == nil {
        return
    }

    _, err = s.executor.Execute(result.SQL(), result.Args()...)
    if err != nil {
        slog.Error("deleteRelationLinks Execute failed",
            "field", vField.GetName(), "error", err.Error())
    }
    return
}

func (s *UpdateRunner) insertRelationLinks(vModel models.Model, vField models.Field, toInsert []any) (err *cd.Error) {
    if len(toInsert) == 0 {
        return
    }

    elemType := vField.GetType().Elem()
    for _, rightID := range toInsert {
        rModel, rErr := s.modelProvider.GetTypeModel(elemType)
        if rErr != nil {
            err = rErr
            slog.Error("insertRelationLinks GetTypeModel failed",
                "field", vField.GetName(), "error", err.Error())
            return
        }

        rVal, rErr := s.modelCodec.ExtractBasicFieldValue(rModel.GetPrimaryField(), rightID)
        if rErr != nil {
            err = rErr
            slog.Error("insertRelationLinks ExtractBasicFieldValue failed",
                "field", vField.GetName(), "error", err.Error())
            return
        }
        rModel.SetPrimaryFieldValue(rVal)

        relationResult, relationErr := s.sqlBuilder.BuildInsertRelation(vModel, vField, rModel)
        if relationErr != nil {
            err = relationErr
            slog.Error("insertRelationLinks BuildInsertRelation failed",
                "field", vField.GetName(), "error", err.Error())
            return
        }

        var idVal any
        err = s.executor.ExecuteInsert(relationResult.SQL(), &idVal, relationResult.Args()...)
        if err != nil {
            slog.Error("insertRelationLinks ExecuteInsert failed",
                "field", vField.GetName(), "error", err.Error())
            return
        }
    }
    return
}

func (s *UpdateRunner) updateReferenceRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
    newRightIDs, prepErr := s.prepareNewRightIDs(vModel, vField)
    if prepErr != nil {
        err = prepErr
        return
    }

    existingRightIDs, queryErr := s.queryExistingRightIDs(vModel, vField)
    if queryErr != nil {
        err = queryErr
        return
    }

    toDelete, toInsert := diffRelationIDs(existingRightIDs, newRightIDs)

    err = s.deleteRelationLinks(vModel, vField, toDelete)
    if err != nil {
        return
    }

    err = s.insertRelationLinks(vModel, vField, toInsert)
    return
}

func (s *UpdateRunner) updateContainRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
    newVal := vField.GetValue().Get()
    err = s.deleteRelation(vModel, vField, 0)
    if err != nil {
        slog.Error("updateContainRelation deleteRelation failed",
            "field", vField.GetName(), "error", err.Error())
        return
    }

    vField.SetValue(newVal)
    err = s.insertRelation(vModel, vField)
    if err != nil {
        slog.Error("updateContainRelation insertRelation failed",
            "field", vField.GetName(), "error", err.Error())
    }
    return
}
```

---

## 9. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 主键类型规范化不一致导致误删/误插 | 关系数据不正确 | 统一使用 `fmt.Sprintf("%v")` 规范化；在集成测试中覆盖 int/string/uuid 主键类型 |
| DB 驱动返回值类型与 PackedBasicFieldValue 不一致 | diffRelationIDs 无法正确匹配 | 集成测试验证；必要时增加类型断言转换 |
| 并发 Update 同一 host 的同一引用字段 | 可能出现重复关系行 | 依赖数据库事务隔离；可选增加 UNIQUE(left, right) 约束 |
| 引用关系下传入无主键实体 | prepareNewRightIDs 返回错误 | 明确约定引用关系实体必须有主键；错误信息清晰 |

---

## 10. 小结

- **关系类型**：**引用关系**（Elem 为指针）只刷新链接差异；**包含关系**（Elem 非指针）用新关联实例替换旧关联实例，仍为先删后插。  
- **引用关系**：Update 时由「先删光再全量插」改为「查当前 right IDs → 与 new right IDs 差集 → 仅删 toDelete 的链接、仅插 toInsert 的链接」。  
- **前提**：引用关系下，关联实体必须已存在于库中且有有效主键（与现有 Insert 的引用关系语义一致）。  
- **新增**：Builder 增加 `BuildDeleteRelationByRights`；orm 在引用关系分支增加差集与准备 newRightIDs 逻辑；包含关系分支保持不变。  
- **不改**：对外 Update 接口、事务边界、关系表结构；引用关系下 toDelete 默认不级联删除关联实体（可选后续）。
- **文件变更**：`database/builder.go`（接口）、`database/postgres/builder_delete.go`（实现）、`database/mysql/builder_delete.go`（实现）、`orm/update.go`（分支调度）、新增 `orm/update_diff.go`（差异更新逻辑）；slice 语义与 Query 选列见 `provider/local`、`database/*/builder_query.go`、`orm/query.go` 及 **docs/QUERY-SLICE-SEMANTICS-FIX.md**。

**实现状态**：第 7 节清单已全部实施；测试覆盖见 **test/update_relation_diff_local_test.go**；场景符合性见 **docs/UPDATE-TEST-COVERAGE.md**、**docs/UPDATE-TEST-DESIGN-COMPLIANCE.md**。本文档与当前代码一致，已归档。
