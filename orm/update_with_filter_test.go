package orm

import (
	"context"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/provider"
)

type conditionalUpdateEntity struct {
	ID      int64  `orm:"id key auto"`
	State   string `orm:"state"`
	Version int64  `orm:"version"`
}

func TestUpdateWithFilterReturnsAffectedRows(t *testing.T) {
	modelProvider := provider.NewLocalProvider("tenant", nil)
	model, err := modelProvider.RegisterModel(&conditionalUpdateEntity{State: "consumed"})
	if err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}
	filter, err := modelProvider.GetModelFilter(model)
	if err != nil {
		t.Fatalf("GetModelFilter failed: %v", err)
	}
	if err := filter.Equal("state", "issued"); err != nil {
		t.Fatalf("filter.Equal(state) failed: %v", err)
	}
	if err := filter.Equal("version", int64(3)); err != nil {
		t.Fatalf("filter.Equal(version) failed: %v", err)
	}

	executor := &fakeExecutor{execRowsAffected: 1}
	instance := &impl{
		context:       context.Background(),
		executor:      executor,
		modelProvider: modelProvider,
		modelCodec:    codec.New(modelProvider, "tenant"),
	}
	rows, updateErr := instance.UpdateWithFilter(model, filter)
	if updateErr != nil {
		t.Fatalf("UpdateWithFilter failed: %v", updateErr)
	}
	if rows != 1 {
		t.Fatalf("unexpected affected rows: got=%d want=1", rows)
	}
	if len(executor.execCalls) != 1 || executor.execCalls[0].kind != "exec" {
		t.Fatalf("expected one execute call, got=%#v", executor.execCalls)
	}
	if got := executor.execCalls[0].sql; !strings.Contains(got, dialectSQLForTest(`WHERE "state" = $2 AND "version" = $3`)) {
		t.Fatalf("conditional predicates missing from sql: %s", got)
	}
}
