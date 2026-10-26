package postgres

import (
	"context"
	"database/sql"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/database"
)

func (s *ConnExecutor) BeginSerializableTransaction() (err *cd.Error) {
	defer func() {
		database.RecordDatabaseTransaction(database.DatabasePostgreSQL, "begin_serializable", err == nil)
	}()
	if s.released || s.dbRollbackOnly || s.dbTx != nil || s.dbTxCount != 0 || s.dbConnPtr == nil {
		return cd.NewError(cd.InvalidOperation, "serializable transaction requires an idle connection")
	}
	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
		s.rowsHandle = nil
	}
	ctx := s.executeContetxt
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return cd.NewError(cd.InvalidOperation, "serializable transaction was canceled")
	}
	tx, txErr := s.dbConnPtr.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if txErr != nil {
		return databaseError(txErr)
	}
	s.dbTx = tx
	s.dbTxCount = 1
	return nil
}

func (s *HostExecutor) BeginSerializableTransaction() (err *cd.Error) {
	defer func() {
		database.RecordDatabaseTransaction(database.DatabasePostgreSQL, "begin_serializable", err == nil)
	}()
	if s.released || s.dbRollbackOnly || s.dbTx != nil || s.dbTxCount != 0 || s.dbHandle == nil {
		return cd.NewError(cd.InvalidOperation, "serializable transaction requires an idle connection")
	}
	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
		s.rowsHandle = nil
	}
	ctx := s.executeContetxt
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return cd.NewError(cd.InvalidOperation, "serializable transaction was canceled")
	}
	tx, txErr := s.dbHandle.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if txErr != nil {
		return databaseError(txErr)
	}
	s.dbTx = tx
	s.dbTxCount = 1
	return nil
}
