package mysql

import (
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql" //引入Mysql驱动

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
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
func NewExecutor(config *Config) (ret *Executor, err *cd.Error) {
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s", config.Username(), config.Password(), config.Server(), config.Database(), config.CharSet())

	executorPtr := &Executor{connectStr: connectStr, dbHandle: nil, dbTx: nil, rowsHandle: nil, dbName: config.Database()}
	err = executorPtr.Connect()
	if err != nil {
		log.Errorf("NewExecutor failed, executorPtr.Connect error:%s", err.Error())
		return
	}

	ret = executorPtr
	return
}

func (s *Executor) Connect() (err *cd.Error) {
	dbHandle, dbErr := sql.Open("mysql", s.connectStr)
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		log.Errorf("open database exception, connectStr:%s, err:%s", s.connectStr, err.Error())
		return
	}

	//log.Print("open database connection...")
	s.dbHandle = dbHandle

	dbErr = dbHandle.Ping()
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		log.Errorf("ping database failed, connectStr:%s, err:%s", s.connectStr, err.Error())
		return
	}

	s.dbHandle = dbHandle
	return
}

func (s *Executor) Ping() (err *cd.Error) {
	if s.dbHandle == nil {
		err = cd.NewError(cd.Unexpected, "must connect to database first")
		log.Errorf("Ping failed, error:%s", err.Error())
		return
	}

	dbErr := s.dbHandle.Ping()
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
	}
	return
}

func (s *Executor) Release() {
	if s.dbTx != nil {
		panic("dbTx isn't nil")
	}

	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.pool == nil {
		if s.dbHandle != nil {
			//log.Print("close database connection...")

			_ = s.dbHandle.Close()
		}
		s.dbHandle = nil
		return
	}

	s.pool.PutIn(s)
}

func (s *Executor) destroy() {
	if s.dbHandle != nil {
		_ = s.dbHandle.Close()
	}
}

func (s *Executor) idle() bool {
	return time.Since(s.finishTime) > 10*time.Minute
}

func (s *Executor) BeginTransaction() (err *cd.Error) {
	atomic.AddInt32(&s.dbTxCount, 1)
	if s.dbTx == nil && s.dbTxCount == 1 {
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
		}
		s.rowsHandle = nil

		tx, txErr := s.dbHandle.Begin()
		if txErr != nil {
			err = cd.NewError(cd.Unexpected, txErr.Error())
			log.Errorf("BeginTransaction failed, s.dbHandle.Begin error:%s", err.Error())
			return
		}

		s.dbTx = tx
		//log.Print("BeginTransaction")
	}

	return
}

func (s *Executor) CommitTransaction() (err *cd.Error) {
	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		dbErr := s.dbTx.Commit()
		if dbErr != nil {
			s.dbTx = nil
			err = cd.NewError(cd.Unexpected, dbErr.Error())
			log.Errorf("CommitTransaction failed, s.dbTx.Commit error:%s", err.Error())
			return
		}

		s.dbTx = nil
		//log.Print("Commit")
	}

	return
}

func (s *Executor) RollbackTransaction() (err *cd.Error) {
	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		dbErr := s.dbTx.Rollback()
		if dbErr != nil {
			s.dbTx = nil
			err = cd.NewError(cd.Unexpected, dbErr.Error())
			log.Errorf("RollbackTransaction failed, s.dbTx.Rollback error:%s", err.Error())
			return
		}

		s.dbTx = nil
		//log.Print("Rollback")
	}

	return
}

func (s *Executor) Query(sql string, needCols bool, args ...any) (ret []string, err *cd.Error) {
	//log.Infof("Query, sql:%s", sql)
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		if err != nil {
			log.Errorf("Query failed, execute time:%s, elapse:%v, sql:%s, err:%s", startTime.Local().String(), elapse, sql, err.Error())
			return
		}

		if traceSQL() {
			log.Infof("Query ok, execute time:%s, elapse:%v, sql:%s", startTime.Local().String(), elapse, sql)
		}
	}()

	if s.dbTx == nil {
		if s.dbHandle == nil {
			panic("dbHandle is nil")
		}
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
			s.rowsHandle = nil
		}

		rows, rowErr := s.dbHandle.Query(sql, args...)
		if rowErr != nil {
			err = cd.NewError(cd.Unexpected, rowErr.Error())
			log.Errorf("Query failed, s.dbHandle.Query:%s, args:%+v, error:%s", sql, args, rowErr.Error())
			return
		}
		if needCols {
			cols, colsErr := rows.Columns()
			if colsErr != nil {
				err = cd.NewError(cd.Unexpected, colsErr.Error())
				log.Errorf("Query failed, rows.Columns:%s, error:%s", sql, colsErr.Error())
				return
			}

			ret = cols
		}
		s.rowsHandle = rows
	} else {
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
			s.rowsHandle = nil
		}

		rows, rowErr := s.dbTx.Query(sql, args...)
		if rowErr != nil {
			err = cd.NewError(cd.Unexpected, rowErr.Error())
			log.Errorf("Query failed, s.dbTx.Query:%s, error:%s", sql, rowErr.Error())
			return
		}
		if needCols {
			cols, colsErr := rows.Columns()
			if colsErr != nil {
				err = cd.NewError(cd.Unexpected, colsErr.Error())
				log.Errorf("Query failed, rows.Columns:%s, error:%s", sql, colsErr.Error())
				return
			}

			ret = cols
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
		_ = s.rowsHandle.Close()
		s.rowsHandle = nil
	}

	return ret
}

func (s *Executor) Finish() {
	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
		s.rowsHandle = nil
	}
}

func (s *Executor) GetField(value ...interface{}) (err *cd.Error) {
	if s.rowsHandle == nil {
		panic("rowsHandle is nil")
	}

	dbErr := s.rowsHandle.Scan(value...)
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		log.Errorf("GetField failed, s.rowsHandle.Scan error:%s", err.Error())
	}

	return
}

func (s *Executor) Execute(sql string, args ...any) (rowsAffected int64, lastInsertID int64, err *cd.Error) {
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		if err != nil {
			log.Errorf("Execute failed, execute time:%v, elapse:%v, sql:%s, err:%s", startTime.Local().String(), elapse, sql, err.Error())
			return
		}

		if traceSQL() {
			log.Infof("Execute ok, execute time:%s, elapse:%v, sql:%s", startTime.Local().String(), elapse, sql)
		}
	}()

	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbTx == nil {
		if s.dbHandle == nil {
			panic("dbHandle is nil")
		}

		result, resultErr := s.dbHandle.Exec(sql, args...)
		if resultErr != nil {
			err = cd.NewError(cd.Unexpected, resultErr.Error())
			log.Errorf("Execute failed, s.dbHandle.Exec error:%s", resultErr.Error())
			return
		}

		rowsAffected, _ = result.RowsAffected()
		lastInsertID, _ = result.LastInsertId()
		return
	}

	result, resultErr := s.dbTx.Exec(sql, args...)
	if resultErr != nil {
		err = cd.NewError(cd.Unexpected, resultErr.Error())
		log.Errorf("Execute failed, s.dbTx.Exec error:%s", resultErr.Error())
		return
	}

	rowsAffected, _ = result.RowsAffected()
	lastInsertID, _ = result.LastInsertId()
	return
}

// CheckTableExist Check Table Exist
func (s *Executor) CheckTableExist(tableName string) (ret bool, err *cd.Error) {
	strSQL := "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_NAME =? and TABLE_SCHEMA =?"
	_, err = s.Query(strSQL, false, tableName, s.dbName)
	if err != nil {
		log.Errorf("CheckTableExist failed, s.Query error:%s", err.Error())
		return
	}

	if s.Next() {
		ret = true
	}

	s.Finish()

	return
}

const (
	initConnCount     = 5
	defaultMaxConnNum = 50
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
func (s *Pool) Initialize(maxConnNum int, configPtr *Config) (err *cd.Error) {
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
			log.Errorf("Initialize failed, NewExecutor error:%s", err.Error())
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

func (s *Pool) GetExecutor() (ret *Executor, err *cd.Error) {
	executorPtr, executorErr := s.FetchOut()
	if executorErr != nil {
		err = executorErr
		return
	}

	ret = executorPtr
	return
}

func (s *Pool) CheckConfig(cfgPtr *Config) *cd.Error {
	if s.config.Same(cfgPtr) {
		return nil
	}

	return cd.NewError(cd.Unexpected, "mismatch database config")
}

// FetchOut FetchOut Executor
func (s *Pool) FetchOut() (ret *Executor, err *cd.Error) {
	defer func() {
		if ret != nil {
			ret.startTime = time.Now()
		}
	}()

	executorPtr, executorErr := s.getFromCache(false)
	if executorErr != nil {
		err = executorErr
		log.Errorf("FetchOut failed, s.getFromCache error:%s", err.Error())
		return
	}
	if executorPtr == nil {
		executorPtr, executorErr = s.getFromIdle()
		if executorErr != nil {
			err = executorErr
			log.Errorf("FetchOut failed, s.getFromIdle error:%s", err.Error())
			return
		}
	}

	if executorPtr == nil {
		executorPtr, executorErr = s.getFromCache(true)
		if executorErr != nil {
			err = executorErr
			log.Errorf("FetchOut failed, s.getFromCache error:%s", err.Error())
			return
		}
	}

	// if ping *cd.Error, reconnect...
	if executorPtr.Ping() != nil {
		err = executorPtr.Connect()
		if err != nil {
			log.Errorf("FetchOut failed, executorPtr.Connect error:%s", err.Error())
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

func (s *Pool) getFromCache(blockFlag bool) (ret *Executor, err *cd.Error) {
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

func (s *Pool) getFromIdle() (ret *Executor, err *cd.Error) {
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
		log.Errorf("getFromIdle failed, NewExecutor error:%s", err.Error())
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
