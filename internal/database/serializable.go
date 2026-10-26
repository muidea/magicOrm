package database

import cd "github.com/muidea/magicCommon/def"

// SerializableTransactionExecutor 在首个读取之前建立可序列化事务。
// 已有事务不能升级；调用方负责处理冲突和未知提交，不自动重放写入。
type SerializableTransactionExecutor interface {
	BeginSerializableTransaction() *cd.Error
}
