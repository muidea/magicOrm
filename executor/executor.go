package executor

import (
	"fmt"

	"github.com/muidea/magicOrm/database/mysql"
)

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

var executorPool *mysql.Pool

// InitializePool initialize pool
func InitializePool(maxConnNum int, user, password, address, dbName string) (err error) {
	if executorPool == nil {
		executorPool = mysql.NewPool()

		err = executorPool.Initialize(10, user, password, address, dbName)
	}

	return
}

// UninitializePool uninitialize pool
func UninitializePool() {
	if executorPool == nil {
		return
	}

	executorPool.Uninitialize()
	executorPool = nil
}

// GetExecutor Get executor
func GetExecutor() (ret Executor, err error) {
	if executorPool == nil {
		err = fmt.Errorf("must initialze executor poll first")
		return
	}

	ret, err = executorPool.FetchOut()
	return
}

// NewExecutor NewExecutor
func NewExecutor(user, password, address, dbName string) (Executor, error) {
	return mysql.FetchExecutor(user, password, address, dbName)
}
