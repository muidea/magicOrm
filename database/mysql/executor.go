package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"sync/atomic"

	_ "github.com/go-sql-driver/mysql" //引入Mysql驱动
)

// Executor Executor
type Executor struct {
	connectStr string
	dbHandle   *sql.DB
	dbTxCount  int32
	dbTx       *sql.Tx
	rowsHandle *sql.Rows
	dbName     string

	pool *Pool
}

// NewExecutor 新建一个数据访问对象
func NewExecutor(user, password, address, dbName string) (ret *Executor, err error) {
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", user, password, address, dbName)

	ret = &Executor{connectStr: connectStr, dbHandle: nil, dbTx: nil, rowsHandle: nil, dbName: dbName}
	return
}

// FetchExecutor 获取一个数据访问对象
func FetchExecutor(user, password, address, dbName string) (ret *Executor, err error) {
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", user, password, address, dbName)

	i := &Executor{connectStr: connectStr, dbHandle: nil, dbTx: nil, rowsHandle: nil, dbName: dbName}
	err = i.Connect()
	if err != nil {
		return
	}
	ret = i

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

	s.pool.PutIn(s)
}

func (s *Executor) destroy() {
	if s.dbHandle != nil {
		s.dbHandle.Close()
	}
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
	cacheExecutor chan *Executor
	idleExecutor  chan *Executor
}

// NewPool new pool
func NewPool() *Pool {
	return &Pool{}
}

// Initialize initialize executor pool
func (s *Pool) Initialize(maxConnNum int, user, password, address, dbName string) (err error) {
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

	s.cacheExecutor = make(chan *Executor, initConnNum)
	if maxConnNum-initConnNum > 0 {
		s.idleExecutor = make(chan *Executor, maxConnNum-initConnNum)
	}

	for idx := 0; idx < maxConnNum; idx++ {
		if idx < initConnNum {
			executor, executorErr := FetchExecutor(user, password, address, dbName)
			if executorErr == nil {
				executor.pool = s
				s.cacheExecutor <- executor
			} else {
				err = executorErr
				return
			}

			continue
		}

		executor, executorErr := NewExecutor(user, password, address, dbName)
		if executorErr == nil {
			executor.pool = s
			s.idleExecutor <- executor
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

	if s.idleExecutor != nil {
		for {
			var val *Executor
			var ok bool
			select {
			case val, ok = <-s.idleExecutor:
			default:
			}
			if ok && val != nil {
				val.destroy()
				continue
			}

			break
		}
		close(s.idleExecutor)
		s.idleExecutor = nil
	}
}

// FetchOut fetchOut Executor
func (s *Pool) FetchOut() (ret *Executor, err error) {
	executor, executorErr := s.getExecutorFromCache(false)
	if executorErr != nil {
		err = executorErr
		return
	}
	if executor == nil {
		executor, executorErr = s.getExecutorFromIdle()
		if executorErr != nil {
			err = executorErr
			return
		}
	}

	if executor == nil {
		executor, executorErr = s.getExecutorFromCache(true)
		if executorErr != nil {
			err = executorErr
			return
		}
	}

	ret = executor

	return
}

// PutIn putIn Executor
func (s *Pool) PutIn(val *Executor) {
	err := val.Ping()
	if err != nil {
		val.dbHandle.Close()
		val.dbHandle = nil
		//val.dbTx = nil
		//val.rowsHandle = nil
		//val.dbTxCount = 0

		s.idleExecutor <- val
	} else {
		s.cacheExecutor <- val
	}
}

func (s *Pool) getExecutorFromCache(blockFlag bool) (ret *Executor, err error) {
	if !blockFlag {
		var val *Executor
		var ok bool
		select {
		case val, ok = <-s.cacheExecutor:
		default:
		}

		if ok && val != nil {
			err = val.Ping()
			if err == nil {
				ret = val
			}
		}

		return
	}

	val := <-s.cacheExecutor
	err = val.Ping()
	if err == nil {
		ret = val
	}

	return
}

func (s *Pool) getExecutorFromIdle() (ret *Executor, err error) {
	val, ok := <-s.idleExecutor
	if ok {
		err = val.Connect()
		if err == nil {
			ret = val
		}

		return
	}

	return
}
