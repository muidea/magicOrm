# 模型与查询设计

**功能块**：`models/`  
**依据**：`models/model.go`、`models/filter.go`、`models/constraint.go`、`models/field.go`、`models/const.go`。

[← 返回设计文档索引](README.md)

---

## 1. Model 接口

| 方法 | 说明 |
|------|------|
| GetName / GetShowName / GetPkgPath / GetPkgKey / GetDescription | 元信息 |
| GetFields / GetField(name) / GetPrimaryField | 字段访问 |
| SetFieldValue(name, val) / SetPrimaryFieldValue(val) | 设置字段/主键 |
| Interface(ptrValue bool) | 转成原始实体指针/值 |
| Copy(viewSpec ViewDeclare) | 按视图复制 Model |
| Reset() | 重置 |

---

## 2. Filter 接口

| 方法 | 说明 |
|------|------|
| Equal / NotEqual / Below / Above | 比较条件 |
| In / NotIn | 集合条件 |
| Like | 模糊匹配 |
| Pagination(pageNum, pageSize) / Sort(fieldName, ascFlag) | 分页与排序 |
| ValueMask(val) | 用实体值填充条件（如主键） |
| MaskModel() | 过滤对应的 Model |
| Paginationer() / Sorter() / GetFilterItem(key) | 分页/排序/单项访问 |

**操作符常量**（`models` 包）：EqualOpr、NotEqualOpr、BelowOpr、AboveOpr、InOpr、NotInOpr、LikeOpr。

---

## 3. 约束（constraint 标签）

与 README 一致，支持以下 `models.Key`：

- **访问行为**：`req`（必填）、`ro`（只读）、`wo`（只写）。
- **内容值**：`min`、`max`、`range`、`in`、`re`（正则）。

约束在验证系统中使用，见 [design-validation.md](design-validation.md)。

---

## 4. 视图（ViewDeclare）

| 视图 | 说明 |
|------|------|
| OriginView | 按 MaskValue 定义字段；MaskValue 为空时等价默认。 |
| DetailView | 详细视图，需在类型定义中声明。 |
| LiteView | 精简视图。 |

Query 前可对 Model 做 `Copy(DetailView)` / `Copy(LiteView)` 控制加载字段。

---

## 附录：操作符与约束键

### Filter 操作符（models 包常量）

| 常量 | 含义 |
|------|------|
| EqualOpr | 等于 (=) |
| NotEqualOpr | 不等于 (!=) |
| BelowOpr | 小于 (<) |
| AboveOpr | 大于 (>) |
| InOpr | 在集合内 (in) |
| NotInOpr | 不在集合内 (!in) |
| LikeOpr | 模糊匹配 (like) |

### 约束 Key（models.Key）

| Key | 含义 |
|-----|------|
| req | 必填 |
| ro | 只读 |
| wo | 只写 |
| min | 最小值/最小长度 |
| max | 最大值/最大长度 |
| range | 数值闭区间 [min, max] |
| in | 枚举约束 |
| re | 正则约束 |
