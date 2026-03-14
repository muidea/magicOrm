# 测试执行指南

**功能块**：`test/`、`test/consistency/`、各包内 `*_test.go`  
**依据**：当前仓库测试布局、`test/global_postgres.go`、`test/global_mysql.go`

[← 返回设计文档索引](README.md)

---

## 1. 总览

`magicOrm` 当前测试可以按执行依赖分成三层：

| 层级 | 典型目录/包 | 是否依赖数据库 | 说明 |
|------|-------------|----------------|------|
| 包内单元测试 | `metrics/...`、`provider/...`、`orm`、`validation/...`、`database/...` | 否 | 主要覆盖本地逻辑、转换、SQL builder、provider 语义、validation 与 metrics。 |
| 一致性测试 | `test/consistency` | 否 | 主要覆盖 Local/Remote、JSON、ObjectValue、helper 往返的一致性。 |
| 集成测试 | `test` | 是 | 真实走 `Orm + Provider + Executor + 数据库` 链路，会创建/删除表并写入测试数据。 |

`go test ./...` 会同时包含这三层，因此在没有数据库环境时，通常只有 `test` 包会失败。

---

## 2. 数据库切换规则

### 2.1 默认：PostgreSQL

`test/global_postgres.go` 使用：

- 地址：`localhost:5432`
- 数据库：`testdb`
- 用户：`postgres`
- 密码：`rootkit`

该文件带有：

```go
//go:build !mysql
```

因此默认执行 `go test ./test` 时，会走 PostgreSQL 配置。

### 2.2 MySQL 变体

`test/global_mysql.go` 使用：

- 地址：`localhost:3306`
- 数据库：`testdb`
- 用户：`root`
- 密码：`rootkit`

该文件带有：

```go
//go:build mysql
```

因此执行：

```bash
go test -tags mysql ./test --count 1
```

时，`test` 包会切换为 MySQL 配置。

---

## 3. 推荐执行方式

### 3.1 纯单元/无数据库

适合本地快速验证绝大部分代码语义：

```bash
go test ./database/... ./metrics/... ./models ./orm ./provider/... ./validation/... ./test/consistency --count 1
```

如果只想验证某条主线：

```bash
go test ./provider/remote ./provider/helper ./orm ./test/consistency --count 1
go test ./metrics/... --count 1
go test ./validation/... --count 1
```

仓库根目录也提供了一个便捷脚本：

```bash
./unit_test.sh
```

### 3.2 PostgreSQL 集成测试

在本机 PostgreSQL 已准备好时执行：

```bash
go test ./test --count 1
./integration_test.sh
```

如果只想跑 Local 集成测试，可以使用：

```bash
./local_test.sh
```

或直接全量：

```bash
go test ./... --count 1
```

### 3.3 MySQL 集成测试

在本机 MySQL 已准备好时执行：

```bash
go test -tags mysql ./test --count 1
```

如果只想跑 Remote 集成测试，可以使用：

```bash
./remote_test.sh
```

如果要切换到 MySQL 变体，可以把参数继续传给脚本：

```bash
./local_test.sh -tags mysql
./remote_test.sh -tags mysql
```

---

## 4. `test` 包内容分组

命名约定：

- Local 集成测试优先使用 `TestLocal...`
- Remote 集成测试优先使用 `TestRemote...`
- `local_test.sh` / `remote_test.sh` 会在这两个前缀基础上兼容少量历史入口名

### 4.1 Local 集成测试

主要文件：

- `optional_local_test.go`
- `reference_crud_local_test.go`
- `query_local_test.go`
- `reference_local_test.go`
- `compose_local_test.go`
- `group_local_test.go`
- `user_local_test.go`
- `system_local_test.go`
- `batch_query_local_test.go`
- `store_local_test.go`
- `transaction_local_test.go`
- `unit_local_test.go`
- `update_relation_diff_local_test.go`
- `constraint_local_test.go`
- `edge_case_local_test.go`
- `nested_local_test.go`
- `performance_local_test.go`

特点：

- 使用 `provider.NewLocalProvider(...)`
- 真实创建/删除表
- 覆盖 Local + ORM + DB executor 路径

### 4.2 Remote 集成测试

主要文件：

- `reference_crud_remote_test.go`
- `query_remote_test.go`
- `reference_remote_test.go`
- `compose_remote_test.go`
- `group_remote_model_test.go`
- `user_remote_model_test.go`
- `system_remote_model_test.go`
- `batch_query_remote_model_test.go`
- `policy_remote_test.go`
- `store_remote_test.go`
- `unit_remote_test.go`
- `constraint_remote_test.go`
- `partner_remote_test.go`

特点：

- 使用 `provider.NewRemoteProvider(...)`
- 仍然依赖真实数据库
- 覆盖 Remote Object/ObjectValue -> ORM -> DB 的端到端链路

注意：

- 这里的 “Remote” 不是外部网络依赖，而是 remote provider 运行时模型。
- 因此它们仍然属于数据库集成测试。

### 4.3 `test/consistency`

这部分不依赖数据库，主要覆盖：

- Local/Remote 类型映射
- helper 往返
- JSON 编解码
- ObjectValue/SliceObjectValue 一致性
- 设计文档与实现的一致性验证

适合作为没有数据库环境时的稳定回归入口。

---

## 5. 常见失败含义

### 5.1 `dial tcp 127.0.0.1:5432 ...`

说明：

- 默认 PostgreSQL 集成测试正在运行
- 当前环境无法访问本机 PostgreSQL

处理方式：

- 启动本机 PostgreSQL，并准备 `testdb/postgres/rootkit`
- 或只跑无数据库测试
- 或切换到 MySQL：
  `go test -tags mysql ./test --count 1`

### 5.2 `pq: ...` / MySQL SQL 错误

说明：

- 集成测试已经连上数据库
- 当前是 schema/SQL/数据语义问题，而不是测试环境缺库

优先检查：

- 最近是否调整了 builder、codec、update/query/insert/delete runner
- 当前数据库是否存在上轮残留表或脏数据

### 5.3 `provider/local` 或 `provider/remote` 单测失败

说明：

- 通常是 Model/Value/Object/ObjectValue 语义发生变化
- 这类失败优先看 provider 层，不要先怀疑数据库

---

## 6. 协作建议

- 日常开发优先跑“无数据库”集合，保证反馈快。
- 改动 ORM runner、database builder、integration 路径时，再补跑 `./test`。
- 日常也可以直接使用仓库根目录脚本：
  - `./unit_test.sh`
  - `./integration_test.sh`
  - `./local_test.sh`
  - `./remote_test.sh`
- 提交前如果只改了 provider/helper/validation/metrics，通常不需要先跑完整数据库集成测试。
- 如果要贴测试失败日志，先说明执行的是哪条命令，以及是否启用了 `-tags mysql`。

---

## 7. 当前边界

- `test` 包当前没有再细分成独立的 `db-integration` 包或 build tag。
- `go test ./...` 仍然会把集成测试和单元测试一起跑。
- 当前仓库默认数据库选择是 PostgreSQL；MySQL 只在显式 `-tags mysql` 时启用。
