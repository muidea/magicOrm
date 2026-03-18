# 错误码定义

**依据**：`github.com/muidea/magicCommon/def` 包中 `Code` 与 `Error` 类型；magicOrm 各层使用约定。

[← 返回设计文档索引](README.md)

---

## 1. 错误类型来源

magicOrm 统一使用 **magicCommon/def** 的 `*cd.Error` 与 `cd.Code`，不单独定义错误码枚举；各模块通过 `cd.NewError(code, message)` 返回错误。

---

## 2. 通用错误码（magicCommon/def）

| Code 常量 | 整型值 | 含义 |
|-----------|--------|------|
| Success | 0 | 成功 |
| UnKnownError | 1 | 未知错误 |
| NotFound | 2 | 未找到（如 Query 无匹配记录） |
| InvalidParameter | 3 | 无效参数 |
| IllegalParam | 4 | 非法参数 |
| InvalidAuthority | 5 | 非法授权 |
| Unexpected | 6 | 意外错误 |
| Duplicated | 7 | 重复（如唯一约束冲突） |
| DatabaseError | 8 | 数据库错误 |
| Timeout | 9 | 超时 |
| NetworkError | 10 | 网络错误 |
| Unauthorized | 11 | 未授权 |
| Forbidden | 12 | 禁止访问 |
| ResourceExhausted | 13 | 资源耗尽 |
| TooManyRequests | 14 | 请求过多 |
| ServiceUnavailable | 15 | 服务不可用 |
| NotImplemented | 16 | 未实现 |
| BadGateway | 17 | 网关错误 |
| DataCorrupted | 18 | 数据损坏 |
| VersionConflict | 19 | 版本冲突 |
| ExternalServiceError | 20 | 外部服务错误 |
| InvalidOperation | 21 | 无效操作 |
| PermissionDenied | 22 | 权限不足 |

---

## 3. magicOrm 中的使用约定

### 3.1 常用 Code

| 场景 | 使用的 Code | 说明 |
|------|-------------|------|
| 参数为 nil / 非法 | IllegalParam | 如 entity nil、filter nil、model value 非法 |
| Query 无匹配记录 | NotFound | 单条 Query 未找到时返回 |
| Query 命中多条记录 | Unexpected | 单条 Query 语义被破坏，需改用 BatchQuery 或收紧条件 |
| 字段值非法、类型不支持 | IllegalParam | 验证失败、字段值不合法 |
| 数据库/执行异常 | 由底层返回（如 DatabaseError、Unexpected） | 具体以 message 为准 |
| 关系字段关联实体无主键 | IllegalParam | 引用关系下关联实体必须有主键 |
| 关系字段参与 Query 但关联主键未赋值 | IllegalParam | 避免把未赋值 relation 静默压成主键零值 |

### 3.2 错误信息格式

- 文本形式：`*cd.Error` 的 `Error()` 或 `Message` 字段，内容为人类可读描述。
- 校验/业务错误：多数带字段名或上下文（如 `"illegal field value, field:xxx"`、`"reference relation field xxx has entity without primary key"`）。

当前未约定额外的结构化错误协议；对外稳定接口仍是 `*cd.Error` 的 `Code + Message`。

---

## 4. 验证系统错误

验证层（`validation/`）失败时，通过 `validation/errors` 包装为 `*cd.Error`：
- 类型层、约束层、场景层错误默认映射为 **IllegalParam**
- 数据库层错误映射为 **DatabaseError**
- 多字段错误会聚合为单个 `IllegalParam`，消息形如 `Multiple validation errors: fieldA (1), fieldB (2)`

---

## 5. 参考

- 错误类型定义：项目依赖的 `vendor/github.com/muidea/magicCommon/def/result.go`。
- 各模块用法：`orm/*.go`、`provider/**/*.go`、`validation/**/*.go` 中的 `cd.NewError` 调用。
