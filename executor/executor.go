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
	Insert(sql string) (int64, error)
	Delete(sql string) (int64, error)
	Update(sql string) (int64, error)
	Execute(sql string) (int64, error)
	CheckTableExist(tableName string) (bool, error)
}

type Pool interface {
	Initialize(maxConnNum int, cfgPtr Config) error
	Uninitialize()
	GetExecutor() (Executor, error)
	CheckConfig(cfgPtr Config) error
}
