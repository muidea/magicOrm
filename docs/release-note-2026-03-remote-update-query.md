# 2026-03 Remote / Update / Query 变更说明

## 1. 范围

本次变更集中在以下几个方向：

- `provider/remote` 的值语义与 JSON 协议一致性
- `orm.Update` 对引用/包含关系的精细化处理
- `orm.Query` 的过滤条件与返回列语义对齐
- `docs` 知识库与当前实现同步

本说明面向团队成员、测试同学和后续排障人员，重点描述“这次变更后系统行为是什么”。

---

## 2. 结果概览

本轮完成后：

- `go test ./... --count 1` 已通过
- remote 对象的“未赋值 / 显式 nil / 显式零值”语义已经打通到：
  - helper
  - JSON encode/decode
  - provider.SetModelValue
  - ORM update/query
- `orm.Update` 不再对未变化的包含关系做整删整建
- `orm.Query(model)` 不再因为查询输入对象过窄而返回残缺结果

---

## 3. 主要行为变化

### 3.1 Remote 字段赋值语义

当前 remote 字段区分三类状态：

- 未赋值
- 显式赋值为 `nil`
- 显式赋值为零值

关键规则：

- `nil` 指针 / `nil` slice：视为未赋值
- 非 `nil` 指针：即使指向零值，也视为显式赋值
- 非 `nil` slice：即使长度为 `0`，也视为显式赋值
- `FieldValue.Assigned` 会跟随 JSON 一起编码/解码

直接影响：

- Remote update 现在可以同时支持：
  - 不传字段：跳过更新
  - 传 `Assigned:true, Value:nil`：清空单值关系
  - 传空 slice：清空集合关系
  - 传零值：显式更新为零值

### 3.2 Update 关系处理

引用关系：

- 只刷新关系表
- 不删除、不重建对端实体

包含关系：

- 先比较数据库当前值与本次输入
- 未变化：跳过
- 单值包含且主键相同：对子对象走原地 `Update`
- 集合包含：优先按子对象主键做增删改
- 只有无法稳定识别时才回退到整组替换

### 3.3 Query 语义

`Query(model)` 现在采用两层稳定语义：

- 输入模型用于生成过滤条件
- 顶层返回固定按 `DetailView` 控制，不再跟随输入模型自身 view 或字段赋值形状漂移

这修复了以下典型问题：

- 只传 `ID` 查询时，返回对象被错误裁成只有少数字段
- remote `Reference` 一类对象查回后与原对象不一致

---

## 4. 对使用方的影响

### 4.1 对调用方透明的修复

以下行为无需业务调用方改代码：

- `orm.Update` 对关系更新的无谓删除/新增减少
- `Query(model)` 返回对象更完整
- remote helper 对空 slice / 零值指针的处理更稳定

同时保留以下查询约定：

- `BatchQuery(filter)` 顶层返回继续遵循 `ValueMask > view`
- 包含/引用的子对象统一收敛到 `lite`

### 4.2 需要知道的协议约定

如果调用方手工构造 remote `ObjectValue`：

- 想表达“跳过字段”，不要只传零值，应保持未赋值
- 想表达“显式 nil 清空”，应传 `FieldValue{Assigned:true, Value:nil}`
- 想表达“显式更新为 0 / false / \"\" / []”，应传 `Assigned:true`

如果调用方通过 helper 从本地 struct 生成 remote 对象，则 helper 会自动处理这些状态。

---

## 5. 已同步的知识库页面

以下文档已按本次行为刷新：

- [design-orm.md](design-orm.md)
- [design-remote-provider.md](design-remote-provider.md)

如果后续再调整 remote/update/query 语义，应优先更新以上两页和本说明。

---

## 6. 建议后续动作

- 如果继续做覆盖率治理，优先补 `test` 包关键场景的统计口径和报告输出
- 如果后续需要开放远端 patch 接口，建议直接沿用当前 `Assigned` 语义，不再设计另一套 nil/zero 协议
- 如果需要对外发布 SDK 使用说明，可从本说明和 `design-remote-provider.md` 抽出“调用方指南”
