package executor

import "github.com/muidea/magicOrm/database/mysql"

// Executor 数据库访问对象
type Executor interface {
	Release()
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
	Query(sql string) error
	Next() bool
	Finish()
	GetField(value ...interface{}) error
	// return auto increment id
	Insert(sql string) (int64, error)
	Delete(sql string) (int64, error)
	Update(sql string) (int64, error)
	Execute(sql string) (int64, error)
	CheckTableExist(tableName string) (bool, error)
}

// NewExecutor NewExecutor
func NewExecutor(user, password, address, dbName string) (Executor, error) {
	return mysql.Fetch(user, password, address, dbName)
}
