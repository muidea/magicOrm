package executor

import (
	"context"

	cd "github.com/muidea/magicCommon/def"
)

type Config interface {
	Server() string
	Username() string
	Password() string
	Database() string
	GetDsn() string
}

// Executor 数据库访问对象
type Executor interface {
	Release()
	BeginTransaction() *cd.Error
	CommitTransaction() *cd.Error
	RollbackTransaction() *cd.Error
	Query(sql string, needCols bool, args ...any) (ret []string, err *cd.Error)
	Next() bool
	Finish()
	GetField(value ...any) *cd.Error
	Execute(sql string, args ...any) (rowsAffected int64, err *cd.Error)
	ExecuteInsert(sql string, pkValOut any, args ...any) (err *cd.Error)
	CheckTableExist(tableName string) (bool, *cd.Error)
}

type Pool interface {
	Initialize(maxConnNum int, config Config) *cd.Error
	Uninitialized()
	GetExecutor(ctx context.Context) (Executor, *cd.Error)
	CheckConfig(config Config) *cd.Error
	IncReference() int
	DecReference() int
}
