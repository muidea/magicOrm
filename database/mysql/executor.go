package mysql

import (
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql" //引入Mysql驱动

	log "github.com/cihub/seelog"
)

const defaultCharSet = "utf8mb4"

type Config struct {
	dbServer string
	dbName   string
	username string
	password string
	charSet  string
}

func (s *Config) Server() string {
	return s.dbServer
}

func (s *Config) Database() string {
	return s.dbName
}

func (s *Config) Username() string {
	return s.username
}

func (s *Config) Password() string {
	return s.password
}

func (s *Config) CharSet() string {
	if s.charSet == "" {
		return defaultCharSet
	}

	return s.charSet
}

func (s *Config) Same(cfg *Config) bool {
	return s.dbServer == cfg.dbServer &&
		s.dbName == cfg.dbName &&
		s.username == cfg.username &&
		s.password == cfg.password
}

func NewConfig(dbServer, dbName, username, password, charSet string) *Config {
	return &Config{dbServer: dbServer, dbName: dbName, username: username, password: password, charSet: charSet}
}

// Executor Executor
type Executor struct {
	connectStr string
	dbHandle   *sql.DB
	dbTxCount  int32
	dbTx       *sql.Tx
	rowsHandle *sql.Rows
	dbName     string

	startTime  time.Time
	finishTime time.Time
	pool       *Pool
}

// NewExecutor 新建一个数据访问对象
func NewExecutor(config *Config) (ret *Executor, err error) {
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s", config.Username(), config.Password(), config.Server(), config.Database(), config.CharSet())

	executorPtr := &Executor{connectStr: connectStr, dbHandle: nil, dbTx: nil, rowsHandle: nil, dbName: config.Database()}
	err = executorPtr.Connect()
	if err != nil {
		return
	}

	ret = executorPtr
	return
}

func (s *Executor) Connect() (err error) {
	db, err := sql.Open("mysql", s.connectStr)
	if err != nil {
		log.Errorf("open database exception, err:%s", err.Error())
		return err
	}

	//log.Print("open database connection...")
	s.dbHandle = db

	err = db.Ping()
	if err != nil {
		log.Errorf("ping database failed, err:%s", err.Error())
		return err
	}

	s.dbHandle = db
	return
}

func (s *Executor) Ping() (err error) {
	if s.dbHandle == nil {
		err = fmt.Errorf("must connect to database first")
		return
	}

	err = s.dbHandle.Ping()
	if err != nil {
		log.Errorf("ping database failed, err:%s", err.Error())
	}

	return
}

func (s *Executor) Release() {
	if s.dbTx != nil {
		panic("dbTx isn't nil")
	}

	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.pool == nil {
		if s.dbHandle != nil {
			//log.Print("close database connection...")

			s.dbHandle.Close()
		}
		s.dbHandle = nil
		return
	}

	s.pool.PutIn(s)
}

func (s *Executor) destroy() {
	if s.dbHandle != nil {
		s.dbHandle.Close()
	}
}

func (s *Executor) idle() bool {
	return time.Now().Sub(s.finishTime) > 10*time.Minute
}

func (s *Executor) BeginTransaction() (err error) {
	atomic.AddInt32(&s.dbTxCount, 1)
	if s.dbTx == nil && s.dbTxCount == 1 {
		if s.rowsHandle != nil {
			s.rowsHandle.Close()
		}
		s.rowsHandle = nil

		tx, txErr := s.dbHandle.Begin()
		if txErr != nil {
			err = txErr
			log.Errorf("begin transaction failed, err:%s", err.Error())
			return
		}

		s.dbTx = tx
		//log.Print("BeginTransaction")
	}

	return
}

func (s *Executor) CommitTransaction() (err error) {
	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		err = s.dbTx.Commit()
		if err != nil {
			s.dbTx = nil

			log.Errorf("commit transaction failed, err:%s", err.Error())
			return
		}

		s.dbTx = nil
		//log.Print("Commit")
	}

	return
}

func (s *Executor) RollbackTransaction() (err error) {
	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		err = s.dbTx.Rollback()
		if err != nil {
			s.dbTx = nil

			log.Errorf("rollback transaction failed, err:%s", err.Error())
			return
		}

		s.dbTx = nil
		//log.Print("Rollback")
	}

	return
}

func (s *Executor) Query(sql string) (err error) {
	//log.Infof("Query, sql:%s", sql)
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		if err != nil {
			log.Errorf("query failed, sql:%s, err:%s", sql, err.Error())
			return
		}

		log.Infof("query ok, elapse:%v", elapse)
	}()

	if s.dbTx == nil {
		if s.dbHandle == nil {
			panic("dbHanlde is nil")
		}
		if s.rowsHandle != nil {
			s.rowsHandle.Close()
			s.rowsHandle = nil
		}

		rows, rowErr := s.dbHandle.Query(sql)
		if rowErr != nil {
			err = rowErr
			return
		}
		s.rowsHandle = rows
	} else {
		if s.rowsHandle != nil {
			s.rowsHandle.Close()
			s.rowsHandle = nil
		}

		rows, rowErr := s.dbTx.Query(sql)
		if rowErr != nil {
			err = rowErr
			return
		}

		s.rowsHandle = rows
	}

	return
}

func (s *Executor) Next() bool {
	if s.rowsHandle == nil {
		panic("rowsHandle is nil")
	}

	ret := s.rowsHandle.Next()
	if !ret {
		//log.Print("Next, close rows")
		s.rowsHandle.Close()
		s.rowsHandle = nil
	}

	return ret
}

func (s *Executor) Finish() {
	if s.rowsHandle != nil {
		s.rowsHandle.Close()
		s.rowsHandle = nil
	}
}

func (s *Executor) GetField(value ...interface{}) (err error) {
	if s.rowsHandle == nil {
		panic("rowsHandle is nil")
	}

	err = s.rowsHandle.Scan(value...)
	if err != nil {
		log.Errorf("scan failed, err:%s", err.Error())
	}

	return
}

func (s *Executor) Execute(sql string) (rowsAffected int64, lastInsertID int64, err error) {
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		if err != nil {
			log.Errorf("execute failed, sql:%s, err:%s", sql, err.Error())
			return
		}

		log.Infof("execute ok, elapse:%v", elapse)
	}()

	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbTx == nil {
		if s.dbHandle == nil {
			panic("dbHandle is nil")
		}

		result, resultErr := s.dbHandle.Exec(sql)
		if resultErr != nil {
			err = resultErr
			return
		}

		rowsAffected, _ = result.RowsAffected()
		lastInsertID, _ = result.LastInsertId()
		return
	}

	result, resultErr := s.dbTx.Exec(sql)
	if resultErr != nil {
		err = resultErr
		return
	}

	rowsAffected, _ = result.RowsAffected()
	lastInsertID, _ = result.LastInsertId()
	return
}

// CheckTableExist Check Table Exist
func (s *Executor) CheckTableExist(tableName string) (ret bool, err error) {
	sql := fmt.Sprintf("SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_NAME ='%s' and TABLE_SCHEMA ='%s'", tableName, s.dbName)

	err = s.Query(sql)
	if err != nil {
		return
	}

	if s.Next() {
		ret = true
	} else {
		ret = false
	}
	s.Finish()

	return
}

const (
	initConnCount     = 16
	defaultMaxConnNum = 1024
)

// Pool executorPool
type Pool struct {
	config        *Config
	maxSize       int
	cacheSize     int
	curSize       int
	executorLock  sync.RWMutex
	cacheExecutor chan *Executor
	idleExecutor  []*Executor
}

// NewPool new pool
func NewPool() *Pool {
	return &Pool{}
}

// Initialize initialize executor pool
func (s *Pool) Initialize(maxConnNum int, configPtr *Config) (err error) {
	initConnNum := 0
	if 0 < maxConnNum {
		if maxConnNum < 16 {
			initConnNum = maxConnNum
		} else {
			initConnNum = maxConnNum / 4
		}
	} else {
		maxConnNum = defaultMaxConnNum
		initConnNum = initConnCount
	}

	s.config = configPtr
	s.maxSize = maxConnNum
	s.cacheSize = initConnNum
	s.curSize = 0

	s.cacheExecutor = make(chan *Executor, s.cacheSize)

	for ; s.curSize < s.cacheSize; s.curSize++ {
		executor, executorErr := NewExecutor(s.config)
		if executorErr == nil {
			executor.pool = s
			s.cacheExecutor <- executor
		} else {
			err = executorErr
			return
		}
	}

	return
}

// Uninitialized uninitialized executor pool
func (s *Pool) Uninitialized() {
	if s.cacheExecutor != nil {
		for {
			var val *Executor
			var ok bool
			select {
			case val, ok = <-s.cacheExecutor:
			default:
			}
			if ok && val != nil {
				val.destroy()
				continue
			}

			break
		}

		close(s.cacheExecutor)
		s.cacheExecutor = nil
	}

	for _, val := range s.idleExecutor {
		val.destroy()
	}

	s.idleExecutor = nil
	s.curSize = 0
	s.cacheSize = 0
	s.maxSize = 0
}

func (s *Pool) GetExecutor() (ret *Executor, err error) {
	executorPtr, executorErr := s.FetchOut()
	if executorErr != nil {
		err = executorErr
		return
	}

	ret = executorPtr
	return
}

func (s *Pool) CheckConfig(cfgPtr *Config) error {
	if s.config.Same(cfgPtr) {
		return nil
	}

	return fmt.Errorf("mismatch database config")
}

// FetchOut FetchOut Executor
func (s *Pool) FetchOut() (ret *Executor, err error) {
	defer func() {
		if ret != nil {
			ret.startTime = time.Now()
		}
	}()

	executorPtr, executorErr := s.getFromCache(false)
	if executorErr != nil {
		err = executorErr
		return
	}
	if executorPtr == nil {
		executorPtr, executorErr = s.getFromIdle()
		if executorErr != nil {
			err = executorErr
			return
		}
	}

	if executorPtr == nil {
		executorPtr, executorErr = s.getFromCache(true)
		if executorErr != nil {
			err = executorErr
			return
		}
	}

	// if ping error,reconnect...
	if executorPtr.Ping() != nil {
		err = executorPtr.Connect()
		if err != nil {
			return
		}
	}

	ret = executorPtr

	return
}

// PutIn PutIn Executor
func (s *Pool) PutIn(val *Executor) {
	if val == nil {
		return
	}

	val.finishTime = time.Now()

	s.executorLock.RLock()
	defer s.executorLock.RUnlock()
	if s.curSize <= s.cacheSize {
		s.cacheExecutor <- val
		return
	}

	go s.putToIdle(val)
	go s.verifyIdle()
}

func (s *Pool) getFromCache(blockFlag bool) (ret *Executor, err error) {
	if !blockFlag {
		var val *Executor
		select {
		case val = <-s.cacheExecutor:
		default:
		}

		ret = val
		return
	}

	ret = <-s.cacheExecutor

	return
}

func (s *Pool) getFromIdle() (ret *Executor, err error) {
	s.executorLock.Lock()
	defer s.executorLock.Unlock()
	if s.curSize >= s.maxSize {
		return
	}

	if len(s.idleExecutor) > 0 {
		ret = s.idleExecutor[0]

		s.idleExecutor = s.idleExecutor[1:]
		return
	}

	executorPtr, executorErr := NewExecutor(s.config)
	if executorErr != nil {
		err = executorErr
		return
	}
	s.curSize++

	ret = executorPtr
	return
}

func (s *Pool) putToIdle(ptr *Executor) {
	if ptr == nil {
		return
	}

	s.executorLock.Lock()
	defer s.executorLock.Unlock()
	s.curSize--

	s.idleExecutor = append(s.idleExecutor, ptr)
}

func (s *Pool) verifyIdle() {
	s.executorLock.Lock()
	defer s.executorLock.Unlock()

	newList := []*Executor{}
	for _, val := range s.idleExecutor {
		if !val.idle() {
			newList = append(newList, val)
			continue
		}

		val.destroy()
	}
	s.idleExecutor = newList
}
