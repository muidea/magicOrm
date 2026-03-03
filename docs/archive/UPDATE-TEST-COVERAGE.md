# Update 测试用例与场景覆盖分析

本文档汇总当前代码库中**针对 ORM Update（`o1.Update` / `orm.Update`）**的测试用例，并对照 **DESIGN-UPDATE-RELATION-DIFF.md** 第 6.2.2 节的场景表分析是否全部覆盖。

---

## 1. 设计文档要求的 Update 关系场景（6.2.2）

| 场景编号 | 场景名称 | 操作 | 验证要点 |
|----------|----------|------|----------|
| R1 | 引用单值：新增 | Insert Article(Author=nil) → Update Article(Author=Author1) | 关系表新增 1 行；Author1 实体不变 |
| R2 | 引用单值：替换 | Insert Article(Author=Author1) → Update Article(Author=Author2) | 关系表删 Author1 行、插 Author2 行；Author1 仍在库中 |
| R3 | 引用单值：清空 | Insert Article(Author=Author1) → Update Article(Author=nil) | 关系表删 Author1 行；Author1 仍在库中 |
| R4 | 引用单值：不变 | Insert Article(Author=Author1) → Update Article(Author=Author1) | 关系表不变（无多余 SQL） |
| R5 | 引用切片：新增 | Insert Article(Tags=[]) → Update Article(Tags=[T1,T2]) | 关系表新增 2 行 |
| R6 | 引用切片：部分替换 | Insert Article(Tags=[T1,T2]) → Update Article(Tags=[T2,T3]) | 删 T1 链接、插 T3 链接；T1 仍在库中 |
| R7 | 引用切片：清空 | Insert Article(Tags=[T1,T2]) → Update Article(Tags=[]) | 关系表删 2 行；T1、T2 仍在库中 |
| R8 | 引用切片：完全相同 | Insert Article(Tags=[T1,T2]) → Update Article(Tags=[T1,T2]) | 不执行关系表 SQL |
| C1 | 包含关系：以新换旧 | Insert Article(Desc=Desc1) → Update Article(Desc=Desc2) | 旧 Desc1 实体被删，新 Desc2 被创建 |

---

## 2. 当前包含 Update 的测试用例一览

以下仅列**调用 ORM `o1.Update(...)` 或 `orm.Update(...)`** 的用例；`helper.UpdateEntity` / `UpdateSliceEntity` 属于 Provider 层，不在此列。

### 2.1 仅更新基础字段（无关系字段）

| 测试文件 | 测试函数 | 模型 | 更新内容 | 对应设计场景 |
|----------|----------|------|----------|--------------|
| base_local_test.go | TestOptional | Optional | name、optional（基础/指针基础） | — |
| base_local_test.go | TestLocalSimple | Simple | Name 改为 "hello" | — |
| base_remote_test.go | TestRemoteOptional | (Optional) | 同 Optional | — |
| base_remote_test.go | TestRemoteSimple | Simple | 同上 | — |
| simple_test.go | TestSimpleLocal | Simple | Name 改为 "hi" | — |
| simple_test.go | TestSimpleRemote | Simple | 同上 | — |
| unit_local_test.go | (单元测试) | 某 obj | 基础字段 | — |
| unit_remote_test.go | 同上 | 同上 | 同上 | — |
| reference_local_test.go | TestReferenceLocal | Simple | Name 等 | — |
| reference_remote_test.go | TestReferenceRemote | Simple | 同上 | — |
| transaction_local_test.go | TestTransactionRollback | Unit | Name 修改后回滚 | — |
| constraint_local_test.go | (约束测试) | ConstraintTestModel | 可更新字段 | — |
| constraint_remote_test.go | 同上 | 同上 | 同上 | — |
| model_special_local_test.go | (KPI 等) | 某 model | 基础字段 | — |

### 2.2 更新含关系字段的模型

| 测试文件 | 测试函数 | 模型 | 更新内容 | 对应设计场景 |
|----------|----------|------|----------|--------------|
| compose_test.go | TestComposeLocal | Compose | **仅改 Name**（基础字段），关系未改 | — |
| compose_test.go | TestComposeRemote | Compose | 同上 | — |
| model_local_test.go | (User 相关) | User | **Group 引用切片**：user1.Group 增加 group3（[]*Group） | **R5 或 R6**（引用切片新增/部分替换） |
| model_local_test.go | (System 相关) | System | **Users 指针切片**：sys1.Users 变更 | 含关系更新，非 6.2.2 直接对应 |
| model_remote_test.go | 同上 | User / System | 同上（Remote） | 同上 |
| store_local_test.go | (Store 测试) | Product | 基础字段 + 可能含嵌套（如 SKUInfo） | 非纯关系表差异场景 |
| batch_operation_local_test.go | testBatchUpdate | 批量 Item | 仅 Name 追加 "_Updated" | — |

### 2.3 专门针对“关系差异更新”的用例

| 测试文件 | 测试函数 | 模型 | 更新内容 | 对应设计场景 |
|----------|----------|------|----------|--------------|
| update_relation_diff_local_test.go | TestUpdateRelationDiffReference | Compose | **ReferencePtr**: r1→r2；**ReferencePtrArray**: [r1]→[r2] | **R2（引用单值替换）+ 引用切片全量替换** |
| update_relation_diff_local_test.go | TestUpdateRelationDiffReferencePartial | Compose | **ReferencePtrArray**: [r1,r2]→[r2,r3] | **R6（引用切片部分替换）** |

---

## 3. 按设计场景的覆盖情况

| 场景 | 是否覆盖 | 说明 |
|------|----------|------|
| **R1 引用单值：新增** | ✅ 已覆盖 | TestUpdateRelationR1ReferenceSingleAdd。 |
| **R2 引用单值：替换** | ✅ 已覆盖 | TestUpdateRelationDiffReference：ReferencePtr r1→r2，并校验 r1/r2 仍存在。 |
| **R3 引用单值：清空** | ⚠️ 部分 | TestUpdateRelationR3ReferenceSingleClear：仅验证 r1 仍存在；nil 未视为已赋值时不断言关系清空。 |
| **R4 引用单值：不变** | ✅ 已覆盖 | TestUpdateRelationR4ReferenceSingleUnchanged。 |
| **R5 引用切片：新增** | ✅ 已覆盖 | TestUpdateRelationR1ReferenceSingleAdd 含 ReferencePtrArray [r1]。 |
| **R6 引用切片：部分替换** | ✅ 已覆盖 | TestUpdateRelationDiffReferencePartial：[r1,r2]→[r2,r3]，并校验 r1 仍在。 |
| **R7 引用切片：清空** | ⚠️ 部分 | TestUpdateRelationR7ReferenceSliceClear：仅验证 r1/r2 仍存在；[] 未视为已赋值，不断言关系表清空。 |
| **R8 引用切片：完全相同** | ✅ 已覆盖 | TestUpdateRelationR8ReferenceSliceUnchanged。 |
| **C1 包含关系：以新换旧** | ✅ 已覆盖 | TestUpdateRelationC1ContainReplace。 |

---

## 4. 小结与建议

### 4.1 已覆盖

- **基础字段 Update**：多处（Simple、Optional、Unit、Constraint、Batch 等）。
- **引用单值替换（R2）**：update_relation_diff_local_test.TestUpdateRelationDiffReference。
- **引用切片部分替换（R6）**：update_relation_diff_local_test.TestUpdateRelationDiffReferencePartial。
- **引用切片新增（R5）**：model_local/model_remote 中 User.Group 追加元素，为间接覆盖。

### 4.2 已补充用例（update_relation_diff_local_test.go）

| 场景 | 测试函数 | 说明 |
|------|----------|------|
| R1 | TestUpdateRelationR1ReferenceSingleAdd | 引用单值新增（nil → r1），验证 r1 仍在、Compose 关系为 [r1] |
| R2 | TestUpdateRelationDiffReference | 引用单值/切片替换（r1→r2） |
| R3 | TestUpdateRelationR3ReferenceSingleClear | 引用清空：仅验证 r1 仍存在；当前 nil/[] 视为未赋值，不断言关系表已清空 |
| R4 | TestUpdateRelationR4ReferenceSingleUnchanged | 引用单值不变，Query 仍为 [r1] |
| R5/R6 | TestUpdateRelationDiffReferencePartial | 引用切片部分替换 [r1,r2]→[r2,r3] |
| R7 | TestUpdateRelationR7ReferenceSliceClear | 引用切片清空：仅验证 r1/r2 仍存在；空切片 [] 未视为已赋值，不断言关系表清空 |
| R8 | TestUpdateRelationR8ReferenceSliceUnchanged | 引用切片不变，Query 仍为 [r1,r2] |
| C1 | TestUpdateRelationC1ContainReplace | 包含关系以新换旧，旧实体删除、新实体创建 |

### 4.3 未完全覆盖的语义

- **R3/R7 的“清空”**：设计上希望 `ReferencePtr=nil` / `ReferencePtrArray=[]` 时更新关系表并清空链接。当前 Value 层将 nil/空切片视为未赋值，Update 不进入 `updateRelation`，故不会清空。R3/R7 用例仅验证“被引用实体仍存在”，完整清空语义需后续将空切片/ nil 指针在“显式设置”场景下视为已赋值，见 IMPLEMENTATION-UPDATE-RELATION-DIFF-ISSUES.md。
