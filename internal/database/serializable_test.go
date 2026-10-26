package database_test

import (
	"testing"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/mysql"
	"github.com/muidea/magicOrm/internal/database/postgres"
)

func TestSerializableTransactionRequiresIdleExecutor(t *testing.T) {
	for _, executor := range []database.SerializableTransactionExecutor{
		&postgres.ConnExecutor{}, &postgres.HostExecutor{},
		&mysql.ConnExecutor{}, &mysql.HostExecutor{},
	} {
		if executor.BeginSerializableTransaction() == nil {
			t.Fatal("没有连接的 executor 不能开启事务")
		}
	}
}
