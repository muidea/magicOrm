package orm

import (
	"context"
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/provider"
	"strings"
	"testing"
)

type conditionalLockExecutor struct {
	fakeExecutor
	active    bool
	remaining int
	failure   *cd.Error
}

func (s *conditionalLockExecutor) Query(sql string, _ bool, args ...any) ([]string, *cd.Error) {
	s.execCalls = append(s.execCalls, fakeExecCall{kind: "query", sql: sql, args: args})
	return nil, nil
}
func (s *conditionalLockExecutor) TransactionActive() bool { return s.active }
func (s *conditionalLockExecutor) RowsError() *cd.Error    { return s.failure }
func (s *conditionalLockExecutor) Next() bool {
	if s.remaining == 0 {
		return false
	}
	s.remaining--
	return true
}

func TestConditionalLockRequiresTransactionAndPreservesIdentityOnlyDelete(t *testing.T) {
	p := provider.NewLocalProvider("lock", nil)
	model, err := p.RegisterModel(&conditionalUpdateEntity{ID: 7})
	if err != nil {
		t.Fatal(err)
	}
	filter, err := p.GetModelFilter(model)
	if err != nil {
		t.Fatal(err)
	}
	if err := filter.Equal("id", int64(7)); err != nil {
		t.Fatal(err)
	}
	if err := filter.Equal("version", int64(3)); err != nil {
		t.Fatal(err)
	}
	e := &conditionalLockExecutor{remaining: 1}
	h := &impl{context: context.Background(), executor: e, modelProvider: p, modelCodec: codec.New(p, "test")}
	if _, err := LockWithFilter(h, model, filter); err == nil || len(e.execCalls) != 0 {
		t.Fatal("autocommit lock accepted")
	}
	e.active = true
	rows, err := LockWithFilter(h, model, filter)
	if err != nil || rows != 1 {
		t.Fatal(rows, err)
	}
	if len(e.execCalls) != 1 || e.execCalls[0].kind != "query" || !strings.HasSuffix(e.execCalls[0].sql, " FOR UPDATE") {
		t.Fatal("conditional lock changed data or omitted lock", e.execCalls)
	}
	rows, err = LockWithFilter(h, model, filter)
	if err != nil || rows != 0 {
		t.Fatal("stale row was not a conflict", rows, err)
	}
	missing, err := p.GetModelFilter(model)
	if err != nil {
		t.Fatal(err)
	}
	if err := missing.Equal("version", int64(3)); err != nil {
		t.Fatal(err)
	}
	calls := len(e.execCalls)
	if _, err := LockWithFilter(h, model, missing); err == nil || len(e.execCalls) != calls {
		t.Fatal("unscoped lock reached executor")
	}
	e.failure = cd.NewError(cd.Unexpected, "interrupted")
	if _, err := LockWithFilter(h, model, filter); err != e.failure {
		t.Fatal("row-stream error lost", err)
	}
}

type stringKey64 struct {
	ID string `orm:"id key" constraint:"max=64"`
}
type stringKeyDefault struct {
	ID string `orm:"id key"`
}

func TestStringPrimaryCapacityChangeRequiresMigration(t *testing.T) {
	p := provider.NewLocalProvider("key-migration", nil)
	before, err := p.RegisterModel(&stringKeyDefault{})
	if err != nil {
		t.Fatal(err)
	}
	after, err := p.RegisterModel(&stringKey64{})
	if err != nil {
		t.Fatal(err)
	}
	if sameSchemaField(before.GetPrimaryField(), after.GetPrimaryField()) {
		t.Fatal("changed physical key capacity treated as unchanged schema")
	}
}
