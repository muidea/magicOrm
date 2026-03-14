# `test` 目录说明

本目录保存 `magicOrm` 的数据库集成测试和一致性测试样例。

如果你只是想快速验证不依赖数据库的逻辑，优先使用项目根目录的：

```bash
./unit_test.sh
```

如果你要跑本目录下依赖真实数据库的集成测试，优先看 [../docs/testing-guide.md](../docs/testing-guide.md)。

---

## 1. 目录结构

| 路径 | 作用 |
|------|------|
| `test/*.go` | 数据库集成测试，真实走 `Orm + Provider + Executor + 数据库` 链路 |
| `test/consistency/` | 无数据库一致性测试，主要覆盖 Local/Remote/helper/JSON/ObjectValue 往返 |
| `test/vmi/` | Remote/VMI 运行时定义样例，用于 remote provider、builder、query/write runner 回归 |

---

## 2. 文件分组

### 2.1 Local 集成测试

主要文件：

- `optional_local_test.go`
- `reference_crud_local_test.go`
- `query_local_test.go`
- `compose_local_test.go`
- `reference_local_test.go`
- `simple_local_test.go`
- `constraint_local_test.go`
- `define_local_test.go`
- `group_local_test.go`
- `index_local_test.go`
- `user_local_test.go`
- `system_local_test.go`
- `batch_query_local_test.go`
- `model_special_local_test.go`
- `nested_local_test.go`
- `store_local_test.go`
- `transaction_local_test.go`
- `unit_local_test.go`
- `update_relation_diff_local_test.go`
- `batch_operation_local_test.go`
- `edge_case_local_test.go`
- `performance_local_test.go`

特点：

- 使用 `provider.NewLocalProvider(...)`
- 依赖本机 PostgreSQL 或 MySQL
- 主要覆盖本地模型、关系、事务、索引、边界场景

### 2.2 Remote 集成测试

主要文件：

- `reference_crud_remote_test.go`
- `query_remote_test.go`
- `compose_remote_test.go`
- `reference_remote_test.go`
- `simple_remote_test.go`
- `constraint_remote_test.go`
- `group_remote_model_test.go`
- `user_remote_model_test.go`
- `system_remote_model_test.go`
- `batch_query_remote_model_test.go`
- `policy_remote_test.go`
- `partner_remote_test.go`
- `store_remote_test.go`
- `unit_remote_test.go`

特点：

- 使用 `provider.NewRemoteProvider(...)`
- 仍然依赖真实数据库
- 主要覆盖 remote object/objectValue、关系、query/update、复杂 VMI 运行链路

### 2.3 公共模型与辅助代码

主要文件：

- `base.go`
- `compose_helpers_test.go`
- `constraint.go`
- `define.go`
- `model.go`
- `owners_test.go`
- `store.go`
- `unit.go`
- `global_postgres.go`
- `global_mysql.go`

说明：

- `global_postgres.go` 是默认配置，`//go:build !mysql`
- `global_mysql.go` 在 `-tags mysql` 时生效

`compose_helpers_test.go` 只保存 Local/Remote 共享的准备逻辑；真正的测试入口已经拆到 `compose_local_test.go` 和 `compose_remote_test.go`。

命名约定：

- Local 集成测试优先使用 `TestLocal...`
- Remote 集成测试优先使用 `TestRemote...`
- 少量保留的 `TestReferenceLocal` / `TestReferenceRemote` / `TestSimpleLocal` / `TestSimpleRemote` / `TestComposeLocal` / `TestComposeRemote` / `TestConstraintLocal` / `TestConstraintRemote` 属于历史入口，仍由脚本纳入执行

---

## 3. 推荐执行方式

### 3.1 无数据库

```bash
./unit_test.sh
go test ./test/consistency --count 1
```

### 3.2 全量集成测试

```bash
./integration_test.sh
```

### 3.3 Local 集成测试

```bash
./local_test.sh
```

### 3.4 Remote 集成测试

```bash
./remote_test.sh
```

### 3.5 MySQL 集成测试

```bash
./integration_test.sh -tags mysql
./local_test.sh -tags mysql
./remote_test.sh -tags mysql
```

---

## 4. 常见问题

### 4.1 `dial tcp 127.0.0.1:5432 ...`

说明默认 PostgreSQL 集成测试正在执行，但当前环境没有可访问的本机 PostgreSQL。

### 4.2 `pq:` 或 MySQL SQL 错误

说明已经连上数据库，这类错误通常是 schema / SQL / runner / codec 的真实回归，不是测试环境缺库。

### 4.3 为什么 `go test ./...` 会失败，但很多包单测都通过

通常是因为：

- 包内单元测试和 `test/consistency` 都不依赖数据库
- `test` 包依赖真实数据库
- 因此在没有数据库环境时，通常只有 `github.com/muidea/magicOrm/test` 会失败
