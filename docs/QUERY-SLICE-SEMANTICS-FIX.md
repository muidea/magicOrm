# Query 与 slice 语义导致的失败及修复说明（归档）

本文档说明在「slice：nil=未赋值、[]=已赋值」约定下 Query 选列与 TestLocalReference 的修复，与 **DESIGN-UPDATE-RELATION-DIFF.md**、**DESIGN-DATABASE-ORM.md** 2.6 一致，供归档与维护参考。

---

## 1. 失败用例与原因分析

### 1.1 TestLocalReference（base_local_test.go:346）

- **现象**：`query reference failed`，断言 `s4.IArray == nil || s4.FArray == nil || s4.PtrStrArray != nil || s4.PtrArray == nil` 失败。
- **原因**：
  - 约定「slice：nil=未赋值、[]=已赋值」后，Query 用 `Reference{ID: x, PtrArray: &newPtrArray}` 时，IArray/FArray 为 nil（未赋值）。
  - 原逻辑在 **getFieldQueryNames / BuildModuleValueHolder / innerAssign 第一段循环** 中对「基础字段」使用 `!models.IsValidField(field)` 即跳过未赋值字段。
  - 因此 IArray、FArray 既未出现在 SELECT 中，也未在 innerAssign 中被赋值，Query 后仍为 nil，导致断言失败。
  - 若改为「所有基础列都 SELECT 并赋值」，则 PtrStrArray 也会从 DB 加载，测试期望 PtrStrArray 保持 nil 会失败。

### 1.2 其它日志中的失败（ContentConstraint、error_handling Unit、convertSliceValue）

- **ContentConstraintTestModel**：约束校验错误（too small、too large、out of range、must be one of、invalid format）多为**预期行为**，测试即验证非法值被拒绝，日志中的 ERROR 不一定表示用例失败。
- **error_handling_test_Unit / iArray NOT NULL**：插入 `Unit` 时若 IArray 为 nil，而表上 iArray 为 NOT NULL，会报 null 违反非空约束。与「nil=未赋值」一致；若需通过插入测试，应在测试数据中显式赋 `IArray: []int{}` 或在 Insert 层对 nil slice 写空数组（需按项目约定决定）。
- **convertSliceValue: value is not slice**：来自 `provider/remote` 的 In/NotIn，当传入非 slice 时报错，属正常校验，与本次 slice 语义修改无直接关系。

## 2. 采用的修复（仅针对 TestLocalReference）

在不改变「nil=未赋值、[]=已赋值」的前提下，统一 Query 的「选列 + 占位 + 赋值」规则为：

- **基础字段**：仅当 **已赋值（IsValidField）** 或 **值类型 slice（IsSliceField && !IsPtrField）** 时参与 SELECT / 占位 / 赋值。
- 效果：
  - 值类型 slice（如 `[]int`、`[]float32`）即使当前为 nil 也会被 SELECT 并赋值，Query 后 IArray、FArray 能从 DB 正确加载。
  - 指针型字段（如 `*[]string`）在 nil 时仍不参与 SELECT，Query 后 PtrStrArray 保持 nil，符合测试期望。

修改位置：

- **database/postgres/builder_query.go**：`getFieldQueryNames`、`BuildModuleValueHolder` 在「基础字段」分支中增加：  
  `if !models.IsValidField(field) && !(models.IsSliceField(field) && !models.IsPtrField(field)) { continue }`
- **database/mysql/builder_query.go**：同上。
- **orm/query.go**：`innerAssign` 第一段循环中，对基础字段采用与 builder 相同的条件，只对「已赋值或值类型 slice」从 queryVal 赋值。

## 3. 验证

- `go test ./test/... -run 'TestLocalReference|TestUpdateRelation' -count=1` 全部通过。
- `go test ./test/... -count=1` 全量通过。

## 4. 小结

- **TestLocalReference** 失败根因是：在「nil=未赋值」下，Query 按「只处理 valid 基础字段」会漏掉 IArray/FArray 的加载，而若改成「所有基础列都加载」又会破坏 PtrStrArray 保持 nil 的预期。
- 修复方式：**仅对「已赋值」或「值类型 slice」的基础字段做 SELECT 与赋值**，既保证值类型 slice 能从 DB 完整加载，又保持指针型未赋值字段不被加载。
- 其余日志中的错误（约束校验、Unit iArray NOT NULL、convertSliceValue）与本次 Query/slice 语义修复无必然联系；若后续有稳定复现的失败用例，可再单独针对约束/Insert/remote filter 做排查与处理。
