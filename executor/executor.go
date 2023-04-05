package executor

type Config interface {
	HostAddress() string
	Username() string
	Password() string
	Database() string
	Same(config Config) bool
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
	Initialize(maxConnNum int, cfgPtr Config) error
	Uninitialized()
	GetExecutor() (Executor, error)
	CheckConfig(cfgPtr Config) error
}
