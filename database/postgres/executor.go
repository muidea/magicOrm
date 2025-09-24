package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq" //引入PostgreSQL驱动

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
)

const defaultSSLMode = "disable"
const defaultCharSet = "UTF8"

type Config struct {
	dbServer string
	dbName   string
	username string
	password string
	sslMode  string
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

func (s *Config) SSLMode() string {
	if s.sslMode == "" {
		return defaultSSLMode
	}

	return s.sslMode
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

func (s *Config) GetDsn() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", s.Username(), s.Password(), s.Server(), s.Database(), s.SSLMode())
}

func NewConfig(dbServer, dbName, username, password, charSet string) *Config {
	return &Config{dbServer: dbServer, dbName: dbName, username: username, password: password, sslMode: "disable"}
}

// NewExecutor 新建一个数据访问对象
func NewExecutor(configPtr *Config) (ret *HostExecutor, err *cd.Error) {
	dsn := configPtr.GetDsn()
	dbHandle, dbErr := sql.Open("postgres", dsn)
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		log.Errorf("open database exception, connectStr:%s, err:%s", dsn, err.Error())
		return
	}

	ret = &HostExecutor{
		dbHandle: dbHandle,
	}

	return
}

// ConnExecutor ConnExecutor
type ConnExecutor struct {
	executeContetxt context.Context
	dbConnPtr       *sql.Conn
	dbTxCount       int32
	dbTx            *sql.Tx
	rowsHandle      *sql.Rows
}

func (s *ConnExecutor) Release() {
	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	if s.dbTx != nil {
		s.dbTx.Rollback()
	}
	if s.dbConnPtr != nil {
		s.dbConnPtr.Close()
	}
}

func (s *ConnExecutor) BeginTransaction() (err *cd.Error) {
	atomic.AddInt32(&s.dbTxCount, 1)
	if s.dbTx == nil && s.dbTxCount == 1 {
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
		}
		s.rowsHandle = nil

		tx, txErr := s.dbConnPtr.BeginTx(s.executeContetxt, nil)
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

func (s *ConnExecutor) CommitTransaction() (err *cd.Error) {
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

func (s *ConnExecutor) RollbackTransaction() (err *cd.Error) {
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

func (s *ConnExecutor) Query(sql string, needCols bool, args ...any) (ret []string, err *cd.Error) {
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
		if s.dbConnPtr == nil {
			panic("dbHandle is nil")
		}
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
			s.rowsHandle = nil
		}

		rows, rowErr := s.dbConnPtr.QueryContext(s.executeContetxt, sql, args...)
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

func (s *ConnExecutor) Next() bool {
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

func (s *ConnExecutor) Finish() {
	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
		s.rowsHandle = nil
	}
}

func (s *ConnExecutor) GetField(value ...interface{}) (err *cd.Error) {
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

func (s *ConnExecutor) Execute(sql string, args ...any) (rowsAffected int64, err *cd.Error) {
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
		if s.dbConnPtr == nil {
			panic("dbHandle is nil")
		}

		result, resultErr := s.dbConnPtr.ExecContext(s.executeContetxt, sql, args...)
		if resultErr != nil {
			err = cd.NewError(cd.Unexpected, resultErr.Error())
			log.Errorf("Execute failed, s.dbHandle.Exec error:%s", resultErr.Error())
			return
		}

		rowsAffected, _ = result.RowsAffected()
		return
	}

	result, resultErr := s.dbTx.Exec(sql, args...)
	if resultErr != nil {
		err = cd.NewError(cd.Unexpected, resultErr.Error())
		log.Errorf("Execute failed, s.dbTx.Exec error:%s", resultErr.Error())
		return
	}

	rowsAffected, _ = result.RowsAffected()
	return
}

func (s *ConnExecutor) ExecuteInsert(sql string, pkValOut any, args ...any) (err *cd.Error) {
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		if err != nil {
			log.Errorf("ExecuteInsert failed, execute time:%v, elapse:%v, sql:%s, err:%s", startTime.Local().String(), elapse, sql, err.Error())
			return
		}

		if traceSQL() {
			log.Infof("ExecuteInsert ok, execute time:%s, elapse:%v, sql:%s", startTime.Local().String(), elapse, sql)
		}
	}()

	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbTx == nil {
		if s.dbConnPtr == nil {
			panic("dbHandle is nil")
		}

		rowPtr := s.dbConnPtr.QueryRowContext(s.executeContetxt, sql, args...)
		if qErr := rowPtr.Err(); qErr != nil {
			err = cd.NewError(cd.Unexpected, qErr.Error())
			log.Errorf("ExecuteInsert failed, rowPtr.Err error:%s", qErr.Error())
			return
		}

		if rErr := rowPtr.Scan(pkValOut); rErr != nil {
			err = cd.NewError(cd.Unexpected, rErr.Error())
			log.Errorf("ExecuteInsert failed, rowPtr.Scan error:%s", rErr.Error())
			return
		}

		return
	}

	rowPtr := s.dbTx.QueryRowContext(s.executeContetxt, sql, args...)
	if qErr := rowPtr.Err(); qErr != nil {
		err = cd.NewError(cd.Unexpected, qErr.Error())
		log.Errorf("ExecuteInsert failed, rowPtr.Err error:%s", qErr.Error())
		return
	}

	if rErr := rowPtr.Scan(pkValOut); rErr != nil {
		err = cd.NewError(cd.Unexpected, rErr.Error())
		log.Errorf("ExecuteInsert failed, rowPtr.Scan error:%s", rErr.Error())
		return
	}

	return
}

// CheckTableExist Check Table Exist
func (s *ConnExecutor) CheckTableExist(tableName string) (ret bool, err *cd.Error) {
	strSQL := "SELECT tablename FROM pg_tables WHERE tablename = $1 AND schemaname = 'public'"
	_, err = s.Query(strSQL, false, tableName)
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

// ConnExecutor ConnExecutor
type HostExecutor struct {
	dbHandle   *sql.DB
	dbTxCount  int32
	dbTx       *sql.Tx
	rowsHandle *sql.Rows
}

func (s *HostExecutor) Release() {
	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	if s.dbTx != nil {
		s.dbTx.Rollback()
	}
	if s.dbHandle != nil {
		s.dbHandle.Close()
	}
}

func (s *HostExecutor) BeginTransaction() (err *cd.Error) {
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

func (s *HostExecutor) CommitTransaction() (err *cd.Error) {
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

func (s *HostExecutor) RollbackTransaction() (err *cd.Error) {
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

func (s *HostExecutor) Query(sql string, needCols bool, args ...any) (ret []string, err *cd.Error) {
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

func (s *HostExecutor) Next() bool {
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

func (s *HostExecutor) Finish() {
	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
		s.rowsHandle = nil
	}
}

func (s *HostExecutor) GetField(value ...interface{}) (err *cd.Error) {
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

func (s *HostExecutor) Execute(sql string, args ...any) (rowsAffected int64, err *cd.Error) {
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
		return
	}

	result, resultErr := s.dbTx.Exec(sql, args...)
	if resultErr != nil {
		err = cd.NewError(cd.Unexpected, resultErr.Error())
		log.Errorf("Execute failed, s.dbTx.Exec error:%s", resultErr.Error())
		return
	}

	rowsAffected, _ = result.RowsAffected()
	return
}

func (s *HostExecutor) ExecuteInsert(sql string, pkValOut any, args ...any) (err *cd.Error) {
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		if err != nil {
			log.Errorf("ExecuteInsert failed, execute time:%v, elapse:%v, sql:%s, err:%s", startTime.Local().String(), elapse, sql, err.Error())
			return
		}

		if traceSQL() {
			log.Infof("ExecuteInsert ok, execute time:%s, elapse:%v, sql:%s", startTime.Local().String(), elapse, sql)
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

		rowPtr := s.dbHandle.QueryRow(sql, args...)
		if qErr := rowPtr.Err(); qErr != nil {
			err = cd.NewError(cd.Unexpected, qErr.Error())
			log.Errorf("ExecuteInsert failed, rowPtr.Err error:%s", qErr.Error())
			return
		}

		if rErr := rowPtr.Scan(pkValOut); rErr != nil {
			err = cd.NewError(cd.Unexpected, rErr.Error())
			log.Errorf("ExecuteInsert failed, rowPtr.Scan error:%s", rErr.Error())
			return
		}

		return
	}

	rowPtr := s.dbTx.QueryRow(sql, args...)
	if qErr := rowPtr.Err(); qErr != nil {
		err = cd.NewError(cd.Unexpected, qErr.Error())
		log.Errorf("ExecuteInsert failed, rowPtr.Err error:%s", qErr.Error())
		return
	}

	if rErr := rowPtr.Scan(pkValOut); rErr != nil {
		err = cd.NewError(cd.Unexpected, rErr.Error())
		log.Errorf("ExecuteInsert failed, rowPtr.Scan error:%s", rErr.Error())
		return
	}

	return
}

// CheckTableExist Check Table Exist
func (s *HostExecutor) CheckTableExist(tableName string) (ret bool, err *cd.Error) {
	strSQL := "SELECT tablename FROM pg_tables WHERE tablename = $1 AND schemaname = 'public'"
	_, err = s.Query(strSQL, false, tableName)
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

// Pool executorPool
type Pool struct {
	config   *Config
	dbHandle *sql.DB

	maxSize int
}

// NewPool new pool
func NewPool() *Pool {
	return &Pool{}
}

// Initialize initialize executor pool
func (s *Pool) Initialize(maxConnNum int, config *Config) (err *cd.Error) {
	if err = s.connect(config.GetDsn()); err != nil {
		return
	}

	s.maxSize = maxConnNum
	return
}

func (s *Pool) connect(dsn string) (err *cd.Error) {
	dbHandle, dbErr := sql.Open("postgres", dsn)
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		log.Errorf("open database exception, connectStr:%s, err:%s", dsn, err.Error())
		return
	}

	//log.Print("open database connection...")
	s.dbHandle = dbHandle

	dbErr = dbHandle.Ping()
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		log.Errorf("ping database failed, connectStr:%s, err:%s", dsn, err.Error())
		return
	}

	s.dbHandle = dbHandle
	return
}

// Uninitialized uninitialized executor pool
func (s *Pool) Uninitialized() {
	if s.dbHandle != nil {
		_ = s.dbHandle.Close()
		s.dbHandle = nil
	}
	s.maxSize = 0
}

func (s *Pool) GetExecutor(ctx context.Context) (ret *ConnExecutor, err *cd.Error) {
	connPtr, connErr := s.dbHandle.Conn(ctx)
	if connErr != nil {
		err = cd.NewError(cd.DatabaseError, connErr.Error())
		return
	}

	ret = &ConnExecutor{
		executeContetxt: ctx,
		dbConnPtr:       connPtr,
	}
	return
}

func (s *Pool) CheckConfig(cfgPtr *Config) *cd.Error {
	if s.config.Same(cfgPtr) {
		return nil
	}

	return cd.NewError(cd.Unexpected, "mismatch database config")
}
