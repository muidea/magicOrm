package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql" //引入Mysql驱动

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"log/slog"
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
func (s *Config) GetDsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s", s.Username(), s.Password(), s.Server(), s.Database(), s.CharSet())
}

func NewConfig(dbServer, dbName, username, password, charSet string) *Config {
	return &Config{dbServer: dbServer, dbName: dbName, username: username, password: password, charSet: charSet}
}

// NewExecutor 新建一个数据访问对象
func NewExecutor(configPtr database.Config) (ret *HostExecutor, err *cd.Error) {
	dsn := configPtr.GetDsn()
	dbHandle, dbErr := sql.Open("mysql", dsn)
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		slog.Error("open database exception", "dsn", dsn, "error", err.Error())
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
		if err := s.rowsHandle.Close(); err != nil {
			slog.Warn("Failed to close rows handle", "error", err.Error())
		}
	}
	if s.dbTx != nil {
		if err := s.dbTx.Rollback(); err != nil && err != sql.ErrTxDone {
			slog.Warn("Failed to rollback transaction", "error", err.Error())
		}
	}
	if s.dbConnPtr != nil {
		if err := s.dbConnPtr.Close(); err != nil {
			slog.Warn("Failed to close database connection", "error", err.Error())
		}
	}
}

func (s *ConnExecutor) BeginTransaction() (err *cd.Error) {
	defer func() {
		database.RecordDatabaseTransaction(database.DatabaseMySQL, "begin", err == nil)
	}()

	atomic.AddInt32(&s.dbTxCount, 1)
	if s.dbTx == nil && s.dbTxCount == 1 {
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
		}
		s.rowsHandle = nil

		tx, txErr := s.dbConnPtr.BeginTx(s.executeContetxt, nil)
		if txErr != nil {
			err = cd.NewError(cd.Unexpected, txErr.Error())
			slog.Error("BeginTransaction failed", "value", "s.dbHandle.Begin", "error", err.Error())
			return
		}

		s.dbTx = tx
		//log.Print("BeginTransaction")
	}

	return
}

func (s *ConnExecutor) CommitTransaction() (err *cd.Error) {
	defer func() {
		database.RecordDatabaseTransaction(database.DatabaseMySQL, "commit", err == nil)
	}()

	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		dbErr := s.dbTx.Commit()
		if dbErr != nil {
			s.dbTx = nil
			err = cd.NewError(cd.Unexpected, dbErr.Error())
			slog.Error("CommitTransaction failed", "value", "s.dbTx.Commit", "error", err.Error())
			return
		}

		s.dbTx = nil
		//log.Print("Commit")
	}

	return
}

func (s *ConnExecutor) RollbackTransaction() (err *cd.Error) {
	defer func() {
		database.RecordDatabaseTransaction(database.DatabaseMySQL, "rollback", err == nil)
	}()

	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		dbErr := s.dbTx.Rollback()
		if dbErr != nil {
			s.dbTx = nil
			err = cd.NewError(cd.Unexpected, dbErr.Error())
			slog.Error("RollbackTransaction failed", "value", "s.dbTx.Rollback", "error", err.Error())
			return
		}

		s.dbTx = nil
		//log.Print("Rollback")
	}

	return
}

func (s *ConnExecutor) Query(sql string, needCols bool, args ...any) (ret []string, err *cd.Error) {
	//slog.Info("message")
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		database.RecordDatabaseQuery(database.DatabaseMySQL, sql, elapse, err)
		if err != nil {
			slog.Error("Query failed", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql, "error", err.Error())
			return
		}

		if traceSQL() {
			slog.Info("Query ok", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql)
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
			slog.Error("Query failed", "sql", sql, "args", args, "error", rowErr.Error())
			return
		}
		if needCols {
			cols, colsErr := rows.Columns()
			if colsErr != nil {
				err = cd.NewError(cd.Unexpected, colsErr.Error())
				slog.Error("Query failed", "operation", "rows.Columns", "sql", sql, "error", colsErr.Error())
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
			slog.Error("Query failed", "operation", "s.dbTx.Query", "sql", sql, "error", rowErr.Error())
			return
		}
		if needCols {
			cols, colsErr := rows.Columns()
			if colsErr != nil {
				err = cd.NewError(cd.Unexpected, colsErr.Error())
				slog.Error("Query failed", "operation", "rows.Columns", "sql", sql, "error", colsErr.Error())
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

func (s *ConnExecutor) GetField(value ...any) (err *cd.Error) {
	if s.rowsHandle == nil {
		panic("rowsHandle is nil")
	}

	dbErr := s.rowsHandle.Scan(value...)
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		slog.Error("GetField failed", "value", "s.rowsHandle.Scan", "error", err.Error())
	}

	return
}

func (s *ConnExecutor) Execute(sql string, args ...any) (rowsAffected int64, err *cd.Error) {
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		database.RecordDatabaseExecution(database.DatabaseMySQL, sql, err == nil)
		if err != nil {
			slog.Error("Execute failed", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql, "error", err.Error())
			return
		}

		if traceSQL() {
			slog.Info("Execute ok", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql)
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
			slog.Error("Execute failed", "value", "s.dbHandle.Exec", "error", resultErr.Error())
			return
		}

		rowsAffected, _ = result.RowsAffected()
		return
	}

	result, resultErr := s.dbTx.Exec(sql, args...)
	if resultErr != nil {
		err = cd.NewError(cd.Unexpected, resultErr.Error())
		slog.Error("Execute failed", "value", "s.dbTx.Exec", "error", resultErr.Error())
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
		database.RecordDatabaseExecution(database.DatabaseMySQL, sql, err == nil)
		if err != nil {
			slog.Error("ExecuteInsert failed", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql, "error", err.Error())
			return
		}

		if traceSQL() {
			slog.Info("ExecuteInsert ok", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql)
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

		execResult, execErr := s.dbConnPtr.ExecContext(s.executeContetxt, sql, args...)
		if execErr != nil {
			err = cd.NewError(cd.Unexpected, execErr.Error())
			slog.Error("ExecuteInsert failed", "value", "s.dbHandle.Exec", "error", execErr.Error())
			return
		}
		idVal, idErr := execResult.LastInsertId()
		if idErr != nil {
			err = cd.NewError(cd.Unexpected, idErr.Error())
			slog.Error("ExecuteInsert failed", "value", "s.dbHandle.Exec", "error", idErr.Error())
			return
		}
		if pkValOut != nil {
			switch raw := pkValOut.(type) {
			case *any:
				*raw = idVal
			default:
				err = cd.NewError(cd.Unexpected, "pkValOut type error, must be *any")
			}
		}

		return
	}

	execResult, execErr := s.dbTx.ExecContext(s.executeContetxt, sql, args...)
	if execErr != nil {
		err = cd.NewError(cd.Unexpected, execErr.Error())
		slog.Error("ExecuteInsert failed", "value", "s.dbHandle.Exec", "error", execErr.Error())
		return
	}
	idVal, idErr := execResult.LastInsertId()
	if idErr != nil {
		err = cd.NewError(cd.Unexpected, idErr.Error())
		slog.Error("ExecuteInsert failed", "value", "s.dbHandle.Exec", "error", idErr.Error())
		return
	}
	if pkValOut != nil {
		switch raw := pkValOut.(type) {
		case *any:
			*raw = idVal
		default:
			err = cd.NewError(cd.Unexpected, "pkValOut type error, must be *any")
		}
	}

	return
}

// CheckTableExist Check Table Exist
func (s *ConnExecutor) CheckTableExist(tableName string) (ret bool, err *cd.Error) {
	strSQL := "SELECT tablename FROM pg_tables WHERE tablename = $1 AND schemaname = 'public'"
	_, err = s.Query(strSQL, false, tableName)
	if err != nil {
		slog.Error("CheckTableExist failed", "value", "s.Query", "error", err.Error())
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
		if err := s.rowsHandle.Close(); err != nil {
			slog.Warn("Failed to close rows handle", "error", err.Error())
		}
	}
	if s.dbTx != nil {
		if err := s.dbTx.Rollback(); err != nil && err != sql.ErrTxDone {
			slog.Warn("Failed to rollback transaction", "error", err.Error())
		}
	}
	if s.dbHandle != nil {
		if err := s.dbHandle.Close(); err != nil {
			slog.Warn("Failed to close database handle", "error", err.Error())
		}
	}
}

func (s *HostExecutor) BeginTransaction() (err *cd.Error) {
	defer func() {
		database.RecordDatabaseTransaction(database.DatabaseMySQL, "begin", err == nil)
	}()

	atomic.AddInt32(&s.dbTxCount, 1)
	if s.dbTx == nil && s.dbTxCount == 1 {
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
		}
		s.rowsHandle = nil

		tx, txErr := s.dbHandle.Begin()
		if txErr != nil {
			err = cd.NewError(cd.Unexpected, txErr.Error())
			slog.Error("BeginTransaction failed", "value", "s.dbHandle.Begin", "error", err.Error())
			return
		}

		s.dbTx = tx
		//log.Print("BeginTransaction")
	}

	return
}

func (s *HostExecutor) CommitTransaction() (err *cd.Error) {
	defer func() {
		database.RecordDatabaseTransaction(database.DatabaseMySQL, "commit", err == nil)
	}()

	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		dbErr := s.dbTx.Commit()
		if dbErr != nil {
			s.dbTx = nil
			err = cd.NewError(cd.Unexpected, dbErr.Error())
			slog.Error("CommitTransaction failed", "value", "s.dbTx.Commit", "error", err.Error())
			return
		}

		s.dbTx = nil
		//log.Print("Commit")
	}

	return
}

func (s *HostExecutor) RollbackTransaction() (err *cd.Error) {
	defer func() {
		database.RecordDatabaseTransaction(database.DatabaseMySQL, "rollback", err == nil)
	}()

	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		dbErr := s.dbTx.Rollback()
		if dbErr != nil {
			s.dbTx = nil
			err = cd.NewError(cd.Unexpected, dbErr.Error())
			slog.Error("RollbackTransaction failed", "value", "s.dbTx.Rollback", "error", err.Error())
			return
		}

		s.dbTx = nil
		//log.Print("Rollback")
	}

	return
}

func (s *HostExecutor) Query(sql string, needCols bool, args ...any) (ret []string, err *cd.Error) {
	//slog.Info("message")
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		database.RecordDatabaseQuery(database.DatabaseMySQL, sql, elapse, err)
		if err != nil {
			slog.Error("Query failed", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql, "error", err.Error())
			return
		}

		if traceSQL() {
			slog.Info("Query ok", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql)
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
			slog.Error("Query failed", "sql", sql, "args", args, "error", rowErr.Error())
			return
		}
		if needCols {
			cols, colsErr := rows.Columns()
			if colsErr != nil {
				err = cd.NewError(cd.Unexpected, colsErr.Error())
				slog.Error("Query failed", "operation", "rows.Columns", "sql", sql, "error", colsErr.Error())
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
			slog.Error("Query failed", "operation", "s.dbTx.Query", "sql", sql, "error", rowErr.Error())
			return
		}
		if needCols {
			cols, colsErr := rows.Columns()
			if colsErr != nil {
				err = cd.NewError(cd.Unexpected, colsErr.Error())
				slog.Error("Query failed", "operation", "rows.Columns", "sql", sql, "error", colsErr.Error())
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

func (s *HostExecutor) GetField(value ...any) (err *cd.Error) {
	if s.rowsHandle == nil {
		panic("rowsHandle is nil")
	}

	dbErr := s.rowsHandle.Scan(value...)
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		slog.Error("GetField failed", "value", "s.rowsHandle.Scan", "error", err.Error())
	}

	return
}

func (s *HostExecutor) Execute(sql string, args ...any) (rowsAffected int64, err *cd.Error) {
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		database.RecordDatabaseExecution(database.DatabaseMySQL, sql, err == nil)
		if err != nil {
			slog.Error("Execute failed", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql, "error", err.Error())
			return
		}

		if traceSQL() {
			slog.Info("Execute ok", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql)
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
			slog.Error("Execute failed", "value", "s.dbHandle.Exec", "error", resultErr.Error())
			return
		}

		rowsAffected, _ = result.RowsAffected()
		return
	}

	result, resultErr := s.dbTx.Exec(sql, args...)
	if resultErr != nil {
		err = cd.NewError(cd.Unexpected, resultErr.Error())
		slog.Error("Execute failed", "value", "s.dbTx.Exec", "error", resultErr.Error())
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
		database.RecordDatabaseExecution(database.DatabaseMySQL, sql, err == nil)
		if err != nil {
			slog.Error("ExecuteInsert failed", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql, "error", err.Error())
			return
		}

		if traceSQL() {
			slog.Info("ExecuteInsert ok", "execute_time", startTime.Local().String(), "elapse", elapse, "sql", sql)
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

		execResult, execErr := s.dbHandle.Exec(sql, args...)
		if execErr != nil {
			err = cd.NewError(cd.Unexpected, execErr.Error())
			slog.Error("ExecuteInsert failed", "value", "s.dbHandle.Exec", "error", execErr.Error())
			return
		}
		idVal, idErr := execResult.LastInsertId()
		if idErr != nil {
			err = cd.NewError(cd.Unexpected, idErr.Error())
			slog.Error("ExecuteInsert failed", "value", "s.dbHandle.Exec", "error", idErr.Error())
			return
		}
		if pkValOut != nil {
			switch raw := pkValOut.(type) {
			case *any:
				*raw = idVal
			default:
				err = cd.NewError(cd.Unexpected, "pkValOut type error, must be *any")
			}
		}

		return
	}

	execResult, execErr := s.dbTx.Exec(sql, args...)
	if execErr != nil {
		err = cd.NewError(cd.Unexpected, execErr.Error())
		slog.Error("ExecuteInsert failed", "value", "s.dbHandle.Exec", "error", execErr.Error())
		return
	}
	idVal, idErr := execResult.LastInsertId()
	if idErr != nil {
		err = cd.NewError(cd.Unexpected, idErr.Error())
		slog.Error("ExecuteInsert failed", "value", "s.dbHandle.Exec", "error", idErr.Error())
		return
	}
	if pkValOut != nil {
		switch raw := pkValOut.(type) {
		case *any:
			*raw = idVal
		default:
			err = cd.NewError(cd.Unexpected, "pkValOut type error, must be *any")
		}
	}
	return
}

// CheckTableExist Check Table Exist
func (s *HostExecutor) CheckTableExist(tableName string) (ret bool, err *cd.Error) {
	strSQL := "SELECT tablename FROM pg_tables WHERE tablename = $1 AND schemaname = 'public'"
	_, err = s.Query(strSQL, false, tableName)
	if err != nil {
		slog.Error("CheckTableExist failed", "value", "s.Query", "error", err.Error())
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
	config         *Config
	dbHandle       *sql.DB
	referenceCount int
}

// NewPool new pool
func NewPool() *Pool {
	return &Pool{}
}

// Initialize initialize executor pool
func (s *Pool) Initialize(maxConnNum int, config database.Config) (err *cd.Error) {
	if err = s.connect(config.GetDsn(), maxConnNum); err != nil {
		return
	}

	return
}

func (s *Pool) connect(dsn string, maxConnNum int) (err *cd.Error) {
	dbHandle, dbErr := sql.Open("mysql", dsn)
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		slog.Error("Pool connect open database exception", "dsn", dsn, "error", err.Error())
		return
	}

	dbHandle.SetMaxOpenConns(maxConnNum)

	//log.Print("open database connection...")
	s.dbHandle = dbHandle

	dbErr = dbHandle.Ping()
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		slog.Error("Pool connect ping database failed", "dsn", dsn, "error", err.Error())
		return
	}

	s.dbHandle = dbHandle
	database.UpdateDatabaseConnectionStats(database.DatabaseMySQL, s.dbHandle)
	return
}

// Uninitialized uninitialized executor pool
func (s *Pool) Uninitialized() {
	if s.dbHandle != nil {
		database.UpdateDatabaseConnectionStats(database.DatabaseMySQL, s.dbHandle)
		_ = s.dbHandle.Close()
		s.dbHandle = nil
	}
}

func (s *Pool) GetExecutor(ctx context.Context) (ret database.Executor, err *cd.Error) {
	connPtr, connErr := s.dbHandle.Conn(ctx)
	if connErr != nil {
		err = cd.NewError(cd.DatabaseError, connErr.Error())
		return
	}

	database.UpdateDatabaseConnectionStats(database.DatabaseMySQL, s.dbHandle)

	ret = &ConnExecutor{
		executeContetxt: ctx,
		dbConnPtr:       connPtr,
	}
	return
}

func (s *Pool) GetStats() database.PoolStats {
	if s == nil || s.dbHandle == nil {
		return database.PoolStats{}
	}

	stats := s.dbHandle.Stats()
	return database.PoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
	}
}

func (s *Pool) CheckConfig(cfgPtr database.Config) *cd.Error {
	newDsn := cfgPtr.GetDsn()
	preDsn := s.config.GetDsn()
	if newDsn == preDsn {
		return nil
	}

	return cd.NewError(cd.Unexpected, "mismatch database config")
}

func (s *Pool) IncReference() int {
	s.referenceCount++
	return s.referenceCount
}

func (s *Pool) DecReference() int {
	s.referenceCount--
	if s.referenceCount < 0 {
		s.referenceCount = 0
	}

	return s.referenceCount
}
