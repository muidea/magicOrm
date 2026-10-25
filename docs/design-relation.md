# 关联关系设计

**功能块**：模型关系（一对多、多对多、单值引用/包含）
**依据**：`~/codespace/magicOrm/orm/insert.go`、`~/codespace/magicOrm/orm/update.go`、`~/codespace/magicOrm/orm/update_diff.go`、`~/codespace/magicOrm/orm/delete.go`、`~/codespace/magicOrm/orm/create.go`、`~/codespace/magicOrm/orm/drop.go`、`~/codespace/magicOrm/database/codec/codec.go`。

[← 返回设计文档索引](~/codespace/magicOrm/docs/README.md)

---

## 1. 关系类型与判定

### 1.1 引用关系 vs 包含关系

按字段类型的 **元素是否为指针** 区分：

| 关系类型 | 判定规则 | 示例类型 | 语义 |
|----------|----------|----------|------|
| **引用关系** | 单值：`GetType().IsPtrType() == true`<br>切片：`GetType().Elem().IsPtrType() == true` | `*Status`、`[]*Group` | 关联实体独立存在，仅维护「谁引用谁」的链接 |
| **包含关系** | 单值：`GetType().IsPtrType() == false`<br>切片：`GetType().Elem().IsPtrType() == false` | `Status`、`[]Tag` | 关联实体从属于 host，随 host 一起增删改 |

- **引用关系**：写入要求非零目标主键，不查询证明目标存在；普通引用只写链接表，参与组合唯一的单值引用内联主表列，不插入目标。
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
- 表名由 `database/codec.ConstructRelationTableName` 固定生成，规则为：
  `LeftModelName + FieldName + RelationCode + RightModelName`，必要时再加数据库前缀。
- 其中 `RelationCode` 当前取值为：
  - `1`：包含单值
  - `2`：包含切片
  - `3`：引用单值
  - `4`：引用切片
- 例如 VMI 回归中的 `tenant_ProductSkuInfo2SkuInfo`、`tenant_ProductStatus3Status`。
- 当前版本不支持通过标签或配置自定义关系表名。

---

## 3. Create / Drop 与关系

### 3.1 Create

- `Orm.Create(entity)` 会递归创建该实体对应表及其需要链接表的关系字段、递归 owned 表。
- 执行前完成全图校验，实际计划从宿主遍历 owned 与关系表；不依赖物理外键顺序。
- **循环引用约束**：当前实现**不推荐**在实体之间构造直接的循环引用（如 `A` 持有 `*B`，`B` 同时持有 `*A` 或 `[]*A`）。此类场景的建表顺序与依赖解析策略在实现中未作为稳定能力对外保证，可能出现建表/删表顺序不符合预期或更新行为不完整。建议将强依赖关系拆分为单向关联或通过单独的「中间实体」建模。

### 3.2 Drop

- `Orm.Drop(entity)` 删除完整 owned 图：宿主、关系表和递归 owned 表；独立引用目标主表保留。Inspect(false) 与 Repair 使用同一图。同图共享 owned 去重，owned 环、物理表名冲突在执行前拒绝。
- 调用 owner 持有管理写入围栏，核对物理命名空间与完整持久使用者目录。跨根 shared owned、图外独立入口或入向引用存在时 Base 拒绝单根删除；本批不新增多根删除协议。停用定义仍有保留结构/数据，不能忽略。
- 单值引用参与组合唯一时，目标标量主键存入宿主列；不建或双写该字段链接表。已有链接表切换为内联列必须显式迁移。

---

## 4. Insert 与关系

- **引用关系**：不对关联实体执行 Insert，普通引用向关系表插入 `(left, right)`，内联单值引用写宿主列；要求关联主键非零，不查询证明目标存在。
- **包含关系**：先对每个关联实体执行 Insert（递归），再向关系表插入 `(left, right)`。
- 单值关系插入一行；切片关系按切片长度插入多行。

---

## 5. Update 与关系（差异增量）

- **引用关系**：内联单值引用更新宿主列；其余做「链接级」差异更新：
  1. 查询当前关系表中该 host 的 `right` ID 列表；
  2. 与本次要写入的新 `right` ID 集合做差集；
  3. 仅删除需移除的链接、仅插入需新增的链接；不删除关联实体本身。
- **包含关系**：用新实例替换旧实例：先删旧（该 host 下该字段的关系行 + 旧关联实体），再插新（新关联实体 + 关系行）。不做链接级差集。
- slice 的 **nil** 与 **[]** 语义：`nil` 视为未赋值，不参与 Update 的关系写；`[]` 视为显式空值，会参与 Update，用于清空关系。

- 单值/切片引用的“清空”语义已经在 Update 路径与 consistency 回归中固定：未赋值表示不覆盖，显式空值表示清空关系。

---

## 6. Delete 与关系

- 先删除该实体在各关系表中作为 `left` 的所有关系行。
- **包含关系**：会级联删除对端关联实体（及其实体表数据）。
- **引用关系**：只删关系行，不删对端实体。
- 当前不支持“删除 host 时级联删除被引用实体”的可选策略；引用关系始终只删除关系行。

---

## 7. Query 与关系

- Query / BatchQuery 可按视图与深度加载关联字段；关系数据通过 `queryRelation` 等从关系表与对端表加载并回填到 Model。
- **slice 的 nil/[] 语义**：与 Update 一致；Query 回填后，字段最终呈现为未赋值还是空切片，取决于写路径和 provider/helper 的对象值语义。
- 当前单值引用字段在未命中关系记录时保持 `nil`；切片关系在无记录时保持未赋值或空切片语义，具体取决于写入链路产生的是 `nil` 还是显式空值。

---

## 8. 与其它文档的交叉引用

- 模型字段定义与标签：[design-models.md](~/codespace/magicOrm/docs/design-models.md)、[tags-reference.md](~/codespace/magicOrm/docs/tags-reference.md)。
- 类型与主键：[type-mapping.md](~/codespace/magicOrm/docs/type-mapping.md)。
- Orm 入口与事务：[design-orm.md](~/codespace/magicOrm/docs/design-orm.md)。
