package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq" //引入PostgreSQL驱动

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
)

const defaultSSLMode = "disable"

type Config struct {
	dbServer string
	dbName   string
	schema   string
	username string
	password string
	sslMode  string
}

func (s *Config) Server() string {
	return s.dbServer
}

func (s *Config) Database() string {
	return s.dbName
}

func (s *Config) Schema() string {
	return s.schema
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

func (s *Config) Same(cfg *Config) bool {
	return s.dbServer == cfg.dbServer &&
		s.dbName == cfg.dbName &&
		s.schema == cfg.schema &&
		s.username == cfg.username &&
		s.password == cfg.password
}

func (s *Config) GetDsn() string {
	return postgresDSN(s, s.Database(), s.Schema())
}

func postgresDSN(config database.Config, databaseName, schemaName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s&options=-c%%20search_path=%s",
		config.Username(), config.Password(), config.Server(), databaseName, defaultSSLMode, schemaName)
}

func validSchemaName(value string) bool {
	if value == "" {
		return false
	}
	for _, item := range value {
		if (item < 'a' || item > 'z') && (item < 'A' || item > 'Z') && (item < '0' || item > '9') && item != '_' {
			return false
		}
	}
	return true
}

// validateSchema verifies the explicitly provisioned PostgreSQL schema. Schema
// lifecycle belongs to the platform storage owner; the ORM must never create
// or repair storage as a side effect of opening an application connection.
func validateSchema(config database.Config) *cd.Error {
	databaseName := strings.TrimSpace(config.Database())
	schemaName := strings.TrimSpace(config.Schema())
	if databaseName == "" || !validSchemaName(schemaName) {
		return cd.NewError(cd.IllegalParam, "postgres database/schema configuration is invalid")
	}
	dbHandle, dbErr := sql.Open("postgres", postgresDSN(config, databaseName, "public"))
	if dbErr != nil {
		return cd.NewError(cd.Unexpected, dbErr.Error())
	}
	defer dbHandle.Close()
	if dbErr = dbHandle.Ping(); dbErr != nil {
		return cd.NewError(cd.Unexpected, dbErr.Error())
	}
	var exists bool
	if dbErr = dbHandle.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = $1)", schemaName).Scan(&exists); dbErr != nil {
		return cd.NewError(cd.Unexpected, dbErr.Error())
	}
	if !exists {
		return cd.NewError(cd.NotFound, "postgres schema was not found")
	}
	return nil
}

// EnsureSchema is an explicit deployment/storage-owner operation. It is not
// called by Pool.Initialize: normal ORM connections still require the schema
// to have been provisioned before they are opened.
func EnsureSchema(config *Config) *cd.Error {
	if config == nil {
		return cd.NewError(cd.IllegalParam, "postgres schema configuration is unavailable")
	}
	databaseName := strings.TrimSpace(config.Database())
	schemaName := strings.TrimSpace(config.Schema())
	if databaseName == "" || !validSchemaName(schemaName) {
		return cd.NewError(cd.IllegalParam, "postgres database/schema configuration is invalid")
	}
	dbHandle, dbErr := sql.Open("postgres", postgresDSN(config, databaseName, "public"))
	if dbErr != nil {
		return cd.NewError(cd.Unexpected, dbErr.Error())
	}
	defer dbHandle.Close()
	if dbErr = dbHandle.Ping(); dbErr != nil {
		return cd.NewError(cd.Unexpected, dbErr.Error())
	}
	// validSchemaName excludes quotes and backslashes, so this identifier is
	// safe to quote here. The value cannot be passed as a bind parameter in DDL.
	quoted := `"` + strings.ReplaceAll(schemaName, `"`, `""`) + `"`
	if _, dbErr = dbHandle.Exec("CREATE SCHEMA IF NOT EXISTS " + quoted); dbErr != nil {
		return cd.NewError(cd.Unexpected, dbErr.Error())
	}
	var exists bool
	if dbErr = dbHandle.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = $1)", schemaName).Scan(&exists); dbErr != nil {
		return cd.NewError(cd.Unexpected, dbErr.Error())
	}
	if !exists {
		return cd.NewError(cd.DataCorrupted, "postgres schema was not observed after provisioning")
	}
	return nil
}

func NewConfig(dbServer, dbName, schemaName, username, password string) *Config {
	return &Config{dbServer: dbServer, dbName: dbName, schema: schemaName, username: username, password: password, sslMode: "disable"}
}

// NewExecutor 新建一个数据访问对象
func NewExecutor(configPtr database.Config) (ret *HostExecutor, err *cd.Error) {
	dsn := configPtr.GetDsn()
	dbHandle, dbErr := sql.Open("postgres", dsn)
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		slog.Error("open database exception", "server", configPtr.Server(), "database", configPtr.Database(), "schema", configPtr.Schema(), "error", err.Error())
		return
	}

	ret = &HostExecutor{
		executeContetxt: context.Background(),
		dbHandle:        dbHandle,
		schemaName:      configPtr.Schema(),
		ownDBHandle:     true,
	}

	return
}

// ConnExecutor ConnExecutor
type ConnExecutor struct {
	executeContetxt context.Context
	dbConnPtr       *sql.Conn
	schemaName      string
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
		database.RecordDatabaseTransaction(database.DatabasePostgreSQL, "begin", err == nil)
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
		database.RecordDatabaseTransaction(database.DatabasePostgreSQL, "commit", err == nil)
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
		database.RecordDatabaseTransaction(database.DatabasePostgreSQL, "rollback", err == nil)
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
		database.RecordDatabaseQuery(database.DatabasePostgreSQL, sql, elapse, err)
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
				slog.Error("Query failed", "sql", sql, "operation", "rows.Columns", "error", colsErr.Error())
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
			slog.Error("Query failed", "sql", sql, "operation", "s.dbTx.Query", "error", rowErr.Error())
			return
		}
		if needCols {
			cols, colsErr := rows.Columns()
			if colsErr != nil {
				err = cd.NewError(cd.Unexpected, colsErr.Error())
				slog.Error("Query failed", "sql", sql, "operation", "rows.Columns", "error", colsErr.Error())
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
		database.RecordDatabaseExecution(database.DatabasePostgreSQL, sql, err == nil)
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
		database.RecordDatabaseExecution(database.DatabasePostgreSQL, sql, err == nil)
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

		rowPtr := s.dbConnPtr.QueryRowContext(s.executeContetxt, sql, args...)
		if qErr := rowPtr.Err(); qErr != nil {
			err = cd.NewError(cd.Unexpected, qErr.Error())
			slog.Error("ExecuteInsert failed", "value", "rowPtr.Err", "error", qErr.Error())
			return
		}

		if rErr := rowPtr.Scan(pkValOut); rErr != nil {
			err = cd.NewError(cd.Unexpected, rErr.Error())
			slog.Error("ExecuteInsert failed", "value", "rowPtr.Scan", "error", rErr.Error())
			return
		}

		return
	}

	rowPtr := s.dbTx.QueryRowContext(s.executeContetxt, sql, args...)
	if qErr := rowPtr.Err(); qErr != nil {
		err = cd.NewError(cd.Unexpected, qErr.Error())
		slog.Error("ExecuteInsert failed", "value", "rowPtr.Err", "error", qErr.Error())
		return
	}

	if rErr := rowPtr.Scan(pkValOut); rErr != nil {
		err = cd.NewError(cd.Unexpected, rErr.Error())
		slog.Error("ExecuteInsert failed", "value", "rowPtr.Scan", "error", rErr.Error())
		return
	}

	return
}

// CheckTableExist Check Table Exist
func (s *ConnExecutor) CheckTableExist(tableName string) (ret bool, err *cd.Error) {
	strSQL := "SELECT tablename FROM pg_tables WHERE tablename = $1 AND schemaname = $2"
	_, err = s.Query(strSQL, false, tableName, s.schemaName)
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
	executeContetxt context.Context
	dbHandle        *sql.DB
	schemaName      string
	dbTxCount       int32
	dbTx            *sql.Tx
	rowsHandle      *sql.Rows
	ownDBHandle     bool
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
	if s.dbHandle != nil && s.ownDBHandle {
		if err := s.dbHandle.Close(); err != nil {
			slog.Warn("Failed to close database handle", "error", err.Error())
		}
	}
}

func (s *HostExecutor) BeginTransaction() (err *cd.Error) {
	defer func() {
		database.RecordDatabaseTransaction(database.DatabasePostgreSQL, "begin", err == nil)
	}()

	atomic.AddInt32(&s.dbTxCount, 1)
	if s.dbTx == nil && s.dbTxCount == 1 {
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
		}
		s.rowsHandle = nil

		var tx *sql.Tx
		var txErr error
		if s.executeContetxt != nil {
			tx, txErr = s.dbHandle.BeginTx(s.executeContetxt, nil)
		} else {
			tx, txErr = s.dbHandle.Begin()
		}
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
		database.RecordDatabaseTransaction(database.DatabasePostgreSQL, "commit", err == nil)
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
		database.RecordDatabaseTransaction(database.DatabasePostgreSQL, "rollback", err == nil)
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
		database.RecordDatabaseQuery(database.DatabasePostgreSQL, sql, elapse, err)
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

		if s.executeContetxt != nil {
			rows, rowErr := s.dbHandle.QueryContext(s.executeContetxt, sql, args...)
			if rowErr != nil {
				err = cd.NewError(cd.Unexpected, rowErr.Error())
				slog.Error("Query failed", "sql", sql, "args", args, "error", rowErr.Error())
				return
			}
			if needCols {
				cols, colsErr := rows.Columns()
				if colsErr != nil {
					err = cd.NewError(cd.Unexpected, colsErr.Error())
					slog.Error("Query failed", "sql", sql, "operation", "rows.Columns", "error", colsErr.Error())
					return
				}

				ret = cols
			}
			s.rowsHandle = rows
			return
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
				slog.Error("Query failed", "sql", sql, "operation", "rows.Columns", "error", colsErr.Error())
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
			slog.Error("Query failed", "sql", sql, "operation", "s.dbTx.Query", "error", rowErr.Error())
			return
		}
		if needCols {
			cols, colsErr := rows.Columns()
			if colsErr != nil {
				err = cd.NewError(cd.Unexpected, colsErr.Error())
				slog.Error("Query failed", "sql", sql, "operation", "rows.Columns", "error", colsErr.Error())
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
		database.RecordDatabaseExecution(database.DatabasePostgreSQL, sql, err == nil)
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

		if s.executeContetxt != nil {
			result, resultErr := s.dbHandle.ExecContext(s.executeContetxt, sql, args...)
			if resultErr != nil {
				err = cd.NewError(cd.Unexpected, resultErr.Error())
				slog.Error("Execute failed", "value", "s.dbHandle.Exec", "error", resultErr.Error())
				return
			}

			rowsAffected, _ = result.RowsAffected()
			return
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
		database.RecordDatabaseExecution(database.DatabasePostgreSQL, sql, err == nil)
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

		if s.executeContetxt != nil {
			rowPtr := s.dbHandle.QueryRowContext(s.executeContetxt, sql, args...)
			if qErr := rowPtr.Err(); qErr != nil {
				err = cd.NewError(cd.Unexpected, qErr.Error())
				slog.Error("ExecuteInsert failed", "value", "rowPtr.Err", "error", qErr.Error())
				return
			}

			if rErr := rowPtr.Scan(pkValOut); rErr != nil {
				err = cd.NewError(cd.Unexpected, rErr.Error())
				slog.Error("ExecuteInsert failed", "value", "rowPtr.Scan", "error", rErr.Error())
				return
			}

			return
		}
		rowPtr := s.dbHandle.QueryRow(sql, args...)
		if qErr := rowPtr.Err(); qErr != nil {
			err = cd.NewError(cd.Unexpected, qErr.Error())
			slog.Error("ExecuteInsert failed", "value", "rowPtr.Err", "error", qErr.Error())
			return
		}
		if rErr := rowPtr.Scan(pkValOut); rErr != nil {
			err = cd.NewError(cd.Unexpected, rErr.Error())
			slog.Error("ExecuteInsert failed", "value", "rowPtr.Scan", "error", rErr.Error())
			return
		}
		return
	}

	rowPtr := s.dbTx.QueryRow(sql, args...)
	if qErr := rowPtr.Err(); qErr != nil {
		err = cd.NewError(cd.Unexpected, qErr.Error())
		slog.Error("ExecuteInsert failed", "value", "rowPtr.Err", "error", qErr.Error())
		return
	}

	if rErr := rowPtr.Scan(pkValOut); rErr != nil {
		err = cd.NewError(cd.Unexpected, rErr.Error())
		slog.Error("ExecuteInsert failed", "value", "rowPtr.Scan", "error", rErr.Error())
		return
	}

	return
}

// CheckTableExist Check Table Exist
func (s *HostExecutor) CheckTableExist(tableName string) (ret bool, err *cd.Error) {
	strSQL := "SELECT tablename FROM pg_tables WHERE tablename = $1 AND schemaname = $2"
	_, err = s.Query(strSQL, false, tableName, s.schemaName)
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
	dbDSN          string
	schemaName     string
	dbHandle       *sql.DB
	referenceCount int
}

// NewPool new pool
func NewPool() *Pool {
	return &Pool{}
}

// Initialize initialize executor pool
func (s *Pool) Initialize(maxConnNum int, config database.Config) (err *cd.Error) {
	if err = validateSchema(config); err != nil {
		return
	}
	if err = s.connect(config.GetDsn(), maxConnNum); err != nil {
		return
	}
	s.schemaName = config.Schema()

	return
}

func (s *Pool) connect(dsn string, maxConnNum int) (err *cd.Error) {
	dbHandle, dbErr := sql.Open("postgres", dsn)
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		slog.Error("Pool connect open database exception", "error", err.Error())
		return
	}

	dbHandle.SetMaxOpenConns(maxConnNum)

	//log.Print("open database connection...")
	s.dbHandle = dbHandle

	dbErr = dbHandle.Ping()
	if dbErr != nil {
		err = cd.NewError(cd.Unexpected, dbErr.Error())
		slog.Error("Pool connect ping database failed", "error", err.Error())
		return
	}

	s.dbDSN = dsn
	s.dbHandle = dbHandle
	database.UpdateDatabaseConnectionStats(database.DatabasePostgreSQL, s.dbHandle)
	return
}

// Uninitialized uninitialized executor pool
func (s *Pool) Uninitialized() {
	if s.dbHandle != nil {
		database.UpdateDatabaseConnectionStats(database.DatabasePostgreSQL, s.dbHandle)
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

	database.UpdateDatabaseConnectionStats(database.DatabasePostgreSQL, s.dbHandle)

	ret = &ConnExecutor{
		executeContetxt: ctx,
		dbConnPtr:       connPtr,
		schemaName:      s.schemaName,
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
	if newDsn == s.dbDSN {
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
