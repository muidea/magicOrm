package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql" //引入Mysql驱动

	"github.com/muidea/magicOrm/executor"
)

type Config struct {
	dbAddress string
	dbName    string
	username  string
	password  string
}

func (s *Config) HostAddress() string {
	return s.dbAddress
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

func (s *Config) Same(cfg executor.Config) bool {
	return s.dbAddress == cfg.HostAddress() &&
		s.dbName == cfg.Database() &&
		s.username == cfg.Username() &&
		s.password == cfg.Password()
}

func NewConfig(dbAddress, dbName, username, password string) *Config {
	return &Config{dbAddress: dbAddress, dbName: dbName, username: username, password: password}
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
func NewExecutor(config executor.Config) (ret *Executor, err error) {
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4", config.Username(), config.Password(), config.HostAddress(), config.Database())

	executorPtr := &Executor{connectStr: connectStr, dbHandle: nil, dbTx: nil, rowsHandle: nil, dbName: config.Database()}
	err = executorPtr.Connect()
	if err != nil {
		return
	}

	ret = executorPtr
	return
}

// Connect connect database
func (s *Executor) Connect() (err error) {
	db, err := sql.Open("mysql", s.connectStr)
	if err != nil {
		log.Printf("open database exception, err:%s", err.Error())
		return err
	}

	//log.Print("open database connection...")
	s.dbHandle = db

	err = db.Ping()
	if err != nil {
		log.Printf("ping database failed, err:%s", err.Error())
		return err
	}

	s.dbHandle = db
	return
}

// Ping ping connection
func (s *Executor) Ping() (err error) {
	if s.dbHandle == nil {
		err = fmt.Errorf("must connect to database first")
		return
	}

	err = s.dbHandle.Ping()
	if err != nil {
		log.Printf("ping database failed, err:%s", err.Error())
	}

	return
}

// Release Release
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

	s.pool.putIn(s)
}

func (s *Executor) destroy() {
	if s.dbHandle != nil {
		s.dbHandle.Close()
	}
}

func (s *Executor) idle() bool {
	return time.Now().Sub(s.finishTime) > 10*time.Minute
}

// BeginTransaction Begin Transaction
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
			log.Printf("begin transaction failed, err:%s", err.Error())
			return
		}

		s.dbTx = tx
		//log.Print("BeginTransaction")
	}

	return
}

// CommitTransaction Commit Transaction
func (s *Executor) CommitTransaction() (err error) {
	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		err = s.dbTx.Commit()
		if err != nil {
			s.dbTx = nil

			log.Printf("commit transaction failed, err:%s", err.Error())
			return
		}

		s.dbTx = nil
		//log.Print("Commit")
	}

	return
}

// RollbackTransaction Rollback Transaction
func (s *Executor) RollbackTransaction() (err error) {
	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		err = s.dbTx.Rollback()
		if err != nil {
			s.dbTx = nil

			log.Printf("rollback transaction failed, err:%s", err.Error())

			return
		}

		s.dbTx = nil
		//log.Print("Rollback")
	}

	return
}

// Query Query
func (s *Executor) Query(sql string) (err error) {
	//log.Printf("Query, sql:%s", sql)
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
			log.Printf("query failed, sql:%s, err:%s", sql, err.Error())
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
			log.Printf("query failed, sql:%s, err:%s", sql, err.Error())
			return
		}

		s.rowsHandle = rows
	}

	return
}

// Next Next
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

// Finish Finish
func (s *Executor) Finish() {
	if s.rowsHandle != nil {
		s.rowsHandle.Close()
		s.rowsHandle = nil
	}
}

// GetField GetField
func (s *Executor) GetField(value ...interface{}) (err error) {
	if s.rowsHandle == nil {
		panic("rowsHandle is nil")
	}

	err = s.rowsHandle.Scan(value...)
	if err != nil {
		log.Printf("scan failed, err:%s", err.Error())
	}

	return
}

// Insert Insert
func (s *Executor) Insert(sql string) (ret int64, err error) {
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
			log.Printf("exec failed, sql:%s, err:%s", sql, err.Error())
			return
		}

		idNum, idErr := result.LastInsertId()
		if idErr != nil {
			err = idErr
			log.Printf("get lastInsertId failed, sql:%s, err:%s", sql, err.Error())
			return
		}
		ret = idNum

		return
	}

	result, resultErr := s.dbTx.Exec(sql)
	if resultErr != nil {
		err = resultErr
		log.Printf("exec failed, sql:%s, err:%s", sql, err.Error())
		return
	}

	idNum, idErr := result.LastInsertId()
	if idErr != nil {
		err = idErr
		log.Printf("get lastInsertId failed, sql:%s, err:%s", sql, err.Error())
		return
	}

	ret = idNum

	return
}

// Update Update
func (s *Executor) Update(sql string) (ret int64, err error) {
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
			log.Printf("exec failed, sql:%s, err:%s", sql, err.Error())
			return
		}

		num, numErr := result.RowsAffected()
		if numErr != nil {
			err = numErr
			log.Printf("get affected rows number failed, sql:%s, err:%s", sql, err.Error())
		}
		ret = num

		return
	}

	result, resultErr := s.dbTx.Exec(sql)
	if resultErr != nil {
		err = resultErr
		log.Printf("exec failed, sql:%s, err:%s", sql, err.Error())
		return
	}

	num, numErr := result.RowsAffected()
	if numErr != nil {
		err = numErr
		log.Printf("get affected rows number failed, sql:%s, err:%s", sql, err.Error())
		return
	}
	ret = num

	return
}

// Delete Delete
func (s *Executor) Delete(sql string) (ret int64, err error) {
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
			log.Printf("exec failed, sql:%s, err:%s", sql, err.Error())
			return
		}

		num, numErr := result.RowsAffected()
		if numErr != nil {
			err = numErr
			log.Printf("get affected rows number failed, sql:%s, err:%s", sql, err.Error())
			return
		}
		ret = num

		return
	}

	result, resultErr := s.dbTx.Exec(sql)
	if resultErr != nil {
		err = resultErr
		log.Printf("exec failed, sql:%s, err:%s", sql, err.Error())
		return
	}

	num, numErr := result.RowsAffected()
	if numErr != nil {
		err = numErr
		log.Printf("get affected rows number failed, sql:%s, err:%s", sql, err.Error())
		return
	}
	ret = num

	return
}

// Execute Execute
func (s *Executor) Execute(sql string) (ret int64, err error) {
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
			log.Printf("exec failed, sql:%s, err:%s", sql, err.Error())
			return
		}

		num, numErr := result.RowsAffected()
		if numErr != nil {
			err = numErr
			log.Printf("get affected rows number failed, sql:%s, err:%s", sql, err.Error())
			return
		}
		ret = num

		return
	}

	result, resultErr := s.dbTx.Exec(sql)
	if resultErr != nil {
		err = resultErr
		log.Printf("exec failed, sql:%s, err:%s", sql, err.Error())
		return
	}

	num, numErr := result.RowsAffected()
	if numErr != nil {
		err = numErr
		log.Printf("get affected rows number failed, sql:%s, err:%s", sql, err.Error())
		return
	}
	ret = num

	return
}

// CheckTableExist CheckTableExist
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
	config        executor.Config
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
func (s *Pool) Initialize(maxConnNum int, cfgPtr executor.Config) (err error) {
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

	s.config = cfgPtr
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

// Uninitialize uninitialize executor pool
func (s *Pool) Uninitialize() {
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

func (s *Pool) GetExecutor() (ret executor.Executor, err error) {
	executorPtr, executorErr := s.fetchOut()
	if executorErr != nil {
		err = executorErr
		return
	}

	ret = executorPtr
	return
}

func (s *Pool) CheckConfig(cfgPtr executor.Config) error {
	if s.config.Same(cfgPtr) {
		return nil
	}

	return fmt.Errorf("mismatch database config")
}

// fetchOut fetchOut Executor
func (s *Pool) fetchOut() (ret *Executor, err error) {
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

// putIn putIn Executor
func (s *Pool) putIn(val *Executor) {
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
