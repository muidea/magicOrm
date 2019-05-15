package executor

import "github.com/muidea/magicOrm/database/mysql"

// Executor 数据库访问对象
type Executor interface {
	Release()
	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()
	Query(sql string)
	Next() bool
	Finish()
	GetField(value ...interface{})
	// return auto increment id
	Insert(sql string) int64
	Delete(sql string) int64
	Update(sql string) int64
	Execute(sql string) int64
	CheckTableExist(tableName string) bool
}

// NewExecutor NewExecutor
func NewExecutor(user, password, address, dbName string) (Executor, error) {
	return mysql.Fetch(user, password, address, dbName)
}
