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
	BeginTransaction() *cd.Result
	CommitTransaction() *cd.Result
	RollbackTransaction() *cd.Result
	Query(sql string) *cd.Result
	Next() bool
	Finish()
	GetField(value ...interface{}) *cd.Result
	Execute(sql string) (rowsAffected int64, lastInsertID int64, err *cd.Result)
	CheckTableExist(tableName string) (bool, *cd.Result)
}

type Pool interface {
	Initialize(maxConnNum int, config Config) *cd.Result
	Uninitialized()
	GetExecutor() (Executor, *cd.Result)
	CheckConfig(config Config) *cd.Result
	IncReference() int
	DecReference() int
}

func NewConfig(dbServer, dbName, username, password, charSet string) Config {
	return mysql.NewConfig(dbServer, dbName, username, password, charSet)
}

func NewExecutor(config Config) (Executor, *cd.Result) {
	return mysql.NewExecutor(mysql.NewConfig(config.Server(), config.Database(), config.Username(), config.Password(), config.CharSet()))
}

func NewPool() Pool {
	return &poolImpl{}
}

type poolImpl struct {
	mysql.Pool
	referenceCount int
}

func (s *poolImpl) Initialize(maxConnNum int, config Config) *cd.Result {
	return s.Pool.Initialize(maxConnNum,
		mysql.NewConfig(config.Server(), config.Database(), config.Username(), config.Password(), config.CharSet()))
}

func (s *poolImpl) Uninitialized() {
	s.Pool.Uninitialized()
}

func (s *poolImpl) GetExecutor() (Executor, *cd.Result) {
	return s.Pool.GetExecutor()
}

func (s *poolImpl) CheckConfig(config Config) *cd.Result {
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
