# 归档文档说明

本目录存放**已合并或已过时**的设计与实现文档，仅供历史查阅与深度实现参考。**当前设计入口与功能块文档请使用上级目录 [docs/README.md](../README.md)。**

## 归档清单

| 文档 | 说明 |
|------|------|
| DESIGN-CONSISTENCY.md | 数据一致性与 Provider 模型抽象（与 docs/design-provider.md、design-models.md 主题重叠，细节更多） |
| DESIGN-CONSISTENCY-VERIFICATION.md | DESIGN-CONSISTENCY 的测试验证记录 |
| DESIGN-DATABASE-ORM.md | 数据库 ORM 操作设计（分层、Runner/Builder/Codec、契约；与 docs/design-orm.md、design-database.md 重叠） |
| DESIGN-UPDATE-RELATION-DIFF.md | Update 关系按差异增量更新方案（已实现并归档） |
| IMPLEMENTATION-UPDATE-RELATION-DIFF-ISSUES.md | 上述实现过程中的异常与核对说明 |
| QUERY-SLICE-SEMANTICS-FIX.md | Query 选列与 slice 语义修复说明 |
| UPDATE-TEST-DESIGN-COMPLIANCE.md | Update 测试与设计符合性说明 |
| UPDATE-TEST-COVERAGE.md | 测试覆盖补充说明 |
| VALIDATION_SYSTEM_COMPLETION.md | 验证系统四层架构完成报告 |

归档文档内部相互引用路径为同一目录下文件名（如 `DESIGN-CONSISTENCY.md`）；引用项目根目录文档请使用 `../../README.md` 等。
