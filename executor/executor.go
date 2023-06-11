package executor

import "github.com/muidea/magicOrm/database/mysql"

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
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
	Query(sql string) error
	Next() bool
	Finish()
	GetField(value ...interface{}) error
	Execute(sql string) (rowsAffected int64, lastInsertID int64, err error)
	CheckTableExist(tableName string) (bool, error)
}

type Pool interface {
	Initialize(maxConnNum int, config Config) error
	Uninitialized()
	GetExecutor() (Executor, error)
	CheckConfig(config Config) error
}

func NewConfig(dbServer, dbName, username, password, charSet string) Config {
	return mysql.NewConfig(dbServer, dbName, username, password, charSet)
}

func NewExecutor(config Config) (Executor, error) {
	return mysql.NewExecutor(mysql.NewConfig(config.Server(), config.Database(), config.Username(), config.Password(), config.CharSet()))
}

func NewPool() Pool {
	return &poolImpl{}
}

type poolImpl struct {
	mysql.Pool
}

func (s *poolImpl) Initialize(maxConnNum int, config Config) error {
	return s.Pool.Initialize(maxConnNum,
		mysql.NewConfig(config.Server(), config.Database(), config.Username(), config.Password(), config.CharSet()))
}

func (s *poolImpl) Uninitialized() {
	s.Pool.Uninitialized()
}

func (s *poolImpl) GetExecutor() (Executor, error) {
	return s.Pool.GetExecutor()
}

func (s *poolImpl) CheckConfig(config Config) error {
	return s.Pool.CheckConfig(mysql.NewConfig(config.Server(), config.Database(), config.Username(), config.Password(), config.CharSet()))
}
