package executor

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/mysql"
)

type Config interface {
	Server() string
	Username() string
	Password() string
	Database() string
	CharSet() string
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
	GetField(value ...interface{}) *cd.Error
	Execute(sql string, args ...any) (rowsAffected int64, lastInsertID int64, err *cd.Error)
	CheckTableExist(tableName string) (bool, *cd.Error)
}

type Pool interface {
	Initialize(maxConnNum int, config Config) *cd.Error
	Uninitialized()
	GetExecutor() (Executor, *cd.Error)
	CheckConfig(config Config) *cd.Error
	IncReference() int
	DecReference() int
}

func NewConfig(dbServer, dbName, username, password, charSet string) Config {
	return mysql.NewConfig(dbServer, dbName, username, password, charSet)
}

func NewExecutor(config Config) (Executor, *cd.Error) {
	return mysql.NewExecutor(mysql.NewConfig(config.Server(), config.Database(), config.Username(), config.Password(), config.CharSet()))
}

func NewPool() Pool {
	return &poolImpl{}
}

type poolImpl struct {
	mysql.Pool
	referenceCount int
}

func (s *poolImpl) Initialize(maxConnNum int, config Config) *cd.Error {
	return s.Pool.Initialize(maxConnNum,
		mysql.NewConfig(config.Server(), config.Database(), config.Username(), config.Password(), config.CharSet()))
}

func (s *poolImpl) Uninitialized() {
	s.Pool.Uninitialized()
}

func (s *poolImpl) GetExecutor() (Executor, *cd.Error) {
	return s.Pool.GetExecutor()
}

func (s *poolImpl) CheckConfig(config Config) *cd.Error {
	return s.Pool.CheckConfig(mysql.NewConfig(config.Server(), config.Database(), config.Username(), config.Password(), config.CharSet()))
}

func (s *poolImpl) IncReference() int {
	s.referenceCount++
	return s.referenceCount
}

func (s *poolImpl) DecReference() int {
	s.referenceCount--
	if s.referenceCount < 0 {
		s.referenceCount = 0
	}

	return s.referenceCount
}
