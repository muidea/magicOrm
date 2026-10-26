package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/database"
)

// BeginSerializableTransaction 必须先于事务内所有读取调用；不降级、不重试。
// 使用独立能力接口，使第三方 executor 未配套时明确失败而非悄悄降低隔离级别。
func BeginSerializableTransaction(handler Orm) *cd.Error {
	capability, ok := handler.(interface{ BeginSerializableTransaction() *cd.Error })
	if !ok || capability == nil {
		return cd.NewError(cd.InvalidOperation, "ORM does not support serializable transactions")
	}
	return capability.BeginSerializableTransaction()
}

func (s *impl) BeginSerializableTransaction() *cd.Error {
	if s == nil || s.executor == nil {
		return cd.NewError(cd.InvalidOperation, "serializable transaction requires an executor")
	}
	executor, ok := s.executor.(database.SerializableTransactionExecutor)
	if !ok {
		return cd.NewError(cd.InvalidOperation, "executor does not support serializable transactions")
	}
	return executor.BeginSerializableTransaction()
}
