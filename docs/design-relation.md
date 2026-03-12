# 关联关系设计

**功能块**：模型关系（一对多、多对多、单值引用/包含）  
**依据**：`orm/insert.go`、`orm/update.go`、`orm/update_diff.go`、`orm/delete.go`、`orm/create.go`、`orm/drop.go`、[archive/DESIGN-UPDATE-RELATION-DIFF.md](archive/DESIGN-UPDATE-RELATION-DIFF.md)。

[← 返回设计文档索引](README.md)

---

## 1. 关系类型与判定

### 1.1 引用关系 vs 包含关系

按字段类型的 **元素是否为指针** 区分，与 [archive/DESIGN-UPDATE-RELATION-DIFF.md](archive/DESIGN-UPDATE-RELATION-DIFF.md) 一致：

| 关系类型 | 判定规则 | 示例类型 | 语义 |
|----------|----------|----------|------|
| **引用关系** | 单值：`GetType().IsPtrType() == true`<br>切片：`GetType().Elem().IsPtrType() == true` | `*Status`、`[]*Group` | 关联实体独立存在，仅维护「谁引用谁」的链接 |
| **包含关系** | 单值：`GetType().IsPtrType() == false`<br>切片：`GetType().Elem().IsPtrType() == false` | `Status`、`[]Tag` | 关联实体从属于 host，随 host 一起增删改 |

- **引用关系**：关联实体必须先存在（有主键）；Insert/Update 只写关系表，不插入关联实体本身。
- **包含关系**：Insert 时会先对关联实体执行 Insert，再写关系表；Update 时用新关联实例替换旧实例（先删旧关系与旧实体，再插新实体与新关系）。

### 1.2 单值 vs 切片

- **单值**：`*Child`、`Child` — 一对一或多对一（当前模型为「一」的一方）。
- **切片**：`[]*Role`、`[]Tag` — 一对多或多对多（当前模型为「一」的一方，关系表存储多对多链接）。

**与实现一致的判定规则（评审 ARCH-IMPL-001）**：引用关系在代码中判定为——单值字段使用 `models.IsPtrField(field)`（即 `field.GetType().IsPtrType()`），切片字段使用 `field.GetType().Elem().IsPtrType()`。与本文 1.1 表一致。

---

## 2. 关系表结构

- 关系表由 ORM 自动创建，结构固定：
  - `id`：自增主键（BIGINT/BIGSERIAL）。
  - `left`：host 实体主键（类型与 host 主键一致）。
  - `right`：关联实体主键（类型与关联实体主键一致）。
- 表名由「host 表名 + 字段名」等规则生成，具体命名规则见各数据库 Builder（如 `BuildCreateRelationTable`）。当前实现采用**可预测的固定规则**（如 `user_groups`、`group_users` 等，带 schema 时为 `{schema}.{table}`），不支持通过标签或配置自定义关系表名。

**待确认**：
- 关系表表名的完整生成规则（含 schema 前缀、长度限制、冲突处理）是否需要在文档中以「精确规则 + 示例」形式统一约定（目前以各数据库 Builder 实现为准）？
- 是否有计划在未来版本支持自定义关系表名或显式指定中间表（当前版本不支持）？

---

## 3. Create / Drop 与关系

### 3.1 Create

- `Orm.Create(entity)` 会递归创建该实体对应表及其**所有**关系字段的关系表。
- 创建顺序：先创建依赖的「包含关系」关联实体表（递归），再创建 host 表，再创建各关系表。
- **循环引用约束**：当前实现**不推荐**在实体之间构造直接的循环引用（如 `A` 持有 `*B`，`B` 同时持有 `*A` 或 `[]*A`）。此类场景的建表顺序与依赖解析策略在实现中未作为稳定能力对外保证，可能出现建表/删表顺序不符合预期或更新行为不完整。建议将强依赖关系拆分为单向关联或通过单独的「中间实体」建模。

### 3.2 Drop

- `Orm.Drop(entity)` **仅处理当前 Model 对应的数据表**：删除该实体表 + 以该实体为 host 的关系表；不级联删除其它实体表或对端实体表中的数据。若需清理对端数据，由业务层显式删除。

---

## 4. Insert 与关系

- **引用关系**：不对关联实体执行 Insert，仅向关系表插入 `(left, right)`；要求关联实体已存在且主键有效。
- **包含关系**：先对每个关联实体执行 Insert（递归），再向关系表插入 `(left, right)`。
- 单值关系插入一行；切片关系按切片长度插入多行。

---

## 5. Update 与关系（差异增量）

- **引用关系**：仅做「链接级」差异更新：
  1. 查询当前关系表中该 host 的 `right` ID 列表；
  2. 与本次要写入的新 `right` ID 集合做差集；
  3. 仅删除需移除的链接、仅插入需新增的链接；不删除关联实体本身。
- **包含关系**：用新实例替换旧实例：先删旧（该 host 下该字段的关系行 + 旧关联实体），再插新（新关联实体 + 关系行）。不做链接级差集。
- slice 的 **nil** 与 **[]** 语义：nil 视为未赋值（不参与 Update 关系写）；`[]` 视为已赋值、size 0，参与 Update，用于「清空关系」。详见 [archive/QUERY-SLICE-SEMANTICS-FIX.md](archive/QUERY-SLICE-SEMANTICS-FIX.md) 与 [archive/DESIGN-UPDATE-RELATION-DIFF.md](archive/DESIGN-UPDATE-RELATION-DIFF.md)。

**待确认**：
- R3/R7（单值引用清空、切片引用清空）的「清空」在当前实现与测试中的完成度与约定是否需在主文档中固定？

---

## 6. Delete 与关系

- 先删除该实体在各关系表中作为 `left` 的所有关系行。
- **包含关系**：会级联删除对端关联实体（及其实体表数据）。
- **引用关系**：只删关系行，不删对端实体。
- **待确认**：引用关系下，是否计划支持「级联删除被引用实体」的可选策略？

---

## 7. Query 与关系

- Query / BatchQuery 可按视图与深度加载关联字段；关系数据通过 `queryRelation` 等从关系表与对端表加载并回填到 Model。
- **slice 的 nil/[] 语义**：与 Update 一致，Query 后 slice 的「未赋值」与「空切片」语义区分见 [archive/QUERY-SLICE-SEMANTICS-FIX.md](archive/QUERY-SLICE-SEMANTICS-FIX.md)；主设计文档引用该说明（评审 ARCH-QRY-002）。
- **待确认**：Query 后单值引用字段（如 `*Status`）若未加载，是否视为已知问题（见评审 ARCH-IMPL-002），文档是否需明确「单值引用默认加载策略」？见 [待确认项清单.md](待确认项清单.md)。

---

## 8. 与其它文档的交叉引用

- 模型字段定义与标签：[design-models.md](design-models.md)、[tags-reference.md](tags-reference.md)。
- 类型与主键：[type-mapping.md](type-mapping.md)。
- 数据流与事务：[design-data-flow.md](design-data-flow.md)。
- 归档：引用关系判定、清空语义、差异更新 [archive/DESIGN-UPDATE-RELATION-DIFF.md](archive/DESIGN-UPDATE-RELATION-DIFF.md)、[archive/IMPLEMENTATION-UPDATE-RELATION-DIFF-ISSUES.md](archive/IMPLEMENTATION-UPDATE-RELATION-DIFF-ISSUES.md)。
