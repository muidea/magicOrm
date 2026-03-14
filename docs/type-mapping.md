# 类型映射表

**依据**：`models/const.go`、`database/postgres/util.go`、`database/mysql/util.go`。

[← 返回设计文档索引](README.md)

---

## 1. Go 类型与模型类型（models）

### 1.1 支持的基本类型

| Go 类型 | models 类型名 | TypeDeclare 常量 | 说明 |
|---------|----------------|------------------|------|
| bool | bool | TypeBooleanValue | 布尔 |
| int8 | int8 | TypeByteValue | 字节 |
| int16 | int16 | TypeSmallIntegerValue | 短整型 |
| int32 | int32 | TypeInteger32Value | 32 位整型 |
| int | int | TypeIntegerValue | 整型 |
| int64 | int64 | TypeBigIntegerValue | 长整型 |
| uint8 | uint8 | TypePositiveByteValue | 无符号字节 |
| uint16 | uint16 | TypePositiveSmallIntegerValue | 无符号短整型 |
| uint32 | uint32 | TypePositiveInteger32Value | 无符号 32 位整型 |
| uint | uint | TypePositiveIntegerValue | 无符号整型 |
| uint64 | uint64 | TypePositiveBigIntegerValue | 无符号长整型 |
| float32 | float32 | TypeFloatValue | 单精度浮点 |
| float64 | float64 | TypeDoubleValue | 双精度浮点 |
| string | string | TypeStringValue | 字符串 |
| time.Time | datetime | TypeDateTimeValue | 日期时间 |
| struct（含 orm 标签） | struct | TypeStructValue | 嵌套/关联实体 |
| slice（上述类型的切片或 *struct 切片） | array | TypeSliceValue | 数组/关系 |

以上类型可带指针（如 `*int64`、`*Status`）；切片可为值类型或指针类型（如 `[]int`、`[]*Group`）。参见 [design-relation.md](design-relation.md)。

### 1.2 不支持或未作为稳定能力文档化的 Go 类型

- `map`、`chan`、`func`、未注册的 `struct`、仅支持类型的 `interface{}` 等未在 models 中声明的类型，不应作为 orm 字段使用。
- `[]byte`、自定义类型别名（如 `type MyID int64`）即使个别路径可能通过反射工作，也未作为稳定字段类型对外承诺。

---

## 2. 模型类型与 PostgreSQL 类型

| models 类型 | PostgreSQL 类型 | 备注 |
|-------------|------------------|------|
| string（主键） | VARCHAR(32) | 主键时固定 32 |
| string（非主键） | TEXT | |
| datetime | TIMESTAMP(3) | 毫秒精度 |
| bool | BOOLEAN | |
| int8 | SMALLINT | |
| int16 / uint8 | SMALLINT | 自增主键时为 SMALLSERIAL |
| int32 / int / uint16 | INTEGER | 自增主键时为 SERIAL |
| int64 / uint32 / uint / uint64 | BIGINT | 自增主键时为 BIGSERIAL |
| float32 | REAL | |
| float64 | DOUBLE PRECISION | |
| slice（序列化） | TEXT | 数组在库中按 TEXT 存储（如 JSON 或逗号分隔，以实现为准） |

---

## 3. 模型类型与 MySQL 类型

| models 类型 | MySQL 类型 | 备注 |
|-------------|------------|------|
| string（主键） | VARCHAR(32) | 主键时固定 32 |
| string（非主键） | TEXT | |
| datetime | DATETIME(3) | 毫秒精度 |
| bool / int8 | TINYINT | |
| int16 / uint8 | SMALLINT | |
| int32 / int / uint16 | INT | |
| int64 / uint32 / uint / uint64 | BIGINT | |
| float32 | FLOAT | |
| float64 | DOUBLE | |
| slice（序列化） | TEXT | 同上 |

---

## 4. 主键与自增

- 主键通过 orm 标签 `key` 指定；自增通过 `auto` 指定（见 [tags-reference.md](tags-reference.md)）。
- 自增仅适用于数值类型主键（如 int64）；PostgreSQL 使用 SERIAL/BIGSERIAL，MySQL 使用 AUTO_INCREMENT。
- `uuid`、`snowflake`、`datetime` 也属于当前实现支持的 `ValueDeclare`：
  - `uuid`：零值时在 Insert 前生成字符串 UUID
  - `snowflake`：零值时在 Insert 前生成 int64 雪花 ID
  - `datetime`：零值时在 Insert 前生成当前时间

---

## 5. 日期时间格式

- 库中为 TIMESTAMP(3)/DATETIME(3)；Go 端为 `time.Time`。
- 在 remote/object value 载体中，`TypeDateTimeValue` 当前按**字符串**承载；本地 model/数据库侧仍对应 `time.Time` 与 TIMESTAMP/DATETIME。
- 字符串格式以当前 codec 与 `utils` 转换实现为准，现阶段未单独定义新的协议层格式文档。
