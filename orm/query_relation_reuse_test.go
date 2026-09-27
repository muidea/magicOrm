package orm

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

func TestBatchRelationsReuseOnlyThisQueryTargetsAndMisses(t *testing.T) {
	p := provider.NewLocalProvider("relation-reuse", nil)
	for _, v := range []any{&queryMaskRelationChild{}, &queryMaskRelationParent{}} {
		if _, err := p.RegisterModel(v); err != nil {
			t.Fatal(err)
		}
	}
	parent, err := p.GetEntityModel(&queryMaskRelationParent{}, true)
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeExecutor{responses: []fakeQueryResponse{{
		match: func(sql string, args []any) bool { return strings.Contains(sql, "_QueryMaskRelationChild") },
		rows:  [][]any{{int64(1), "first", "hidden"}, {int64(2), "second", "hidden"}},
	}}}
	makeRunner := func() *QueryRunner {
		return NewQueryRunner(context.Background(), parent, parent.Copy(models.LiteView), false, executor, p, codec.New(p, "tenant"), true, 0)
	}
	runner := makeRunner()
	field := parent.GetField("child")
	if err := runner.batchQueryRelationModels(field, []any{int64(1), int64(1), int64(2)}, 0); err != nil {
		t.Fatal(err)
	}
	if len(executor.execCalls) != 1 || !reflect.DeepEqual(executor.execCalls[0].args, []any{int64(1), int64(2)}) {
		t.Fatalf("initial queries: %+v", executor.execCalls)
	}
	executor.responses[0].rows = [][]any{{int64(3), "third", "hidden"}}
	if err := runner.batchQueryRelationModels(parent.GetField("childList"), []any{int64(2), int64(3)}, 0); err != nil {
		t.Fatal(err)
	}
	if len(executor.execCalls) != 2 || !reflect.DeepEqual(executor.execCalls[1].args, []any{int64(3)}) {
		t.Fatalf("repeated target fetched: %+v", executor.execCalls)
	}
	child, err := p.GetEntityModel(&queryMaskRelationChild{}, true)
	if err != nil {
		t.Fatal(err)
	}
	cached := runner.getCachedRelationModel(child.GetPkgKey(), int64(1))
	if cached == nil {
		t.Fatal("first target lost")
	}
	value := cached.Interface(true).(*queryMaskRelationChild)
	if value.Name != "first" || value.Secret != "" {
		t.Fatalf("lite projection changed: %+v", value)
	}
	executor.responses[0].rows = nil
	for i := 0; i < 2; i++ {
		if err := runner.batchQueryRelationModels(field, []any{int64(4)}, 0); err != nil {
			t.Fatal(err)
		}
	}
	if len(executor.execCalls) != 3 {
		t.Fatalf("confirmed miss queried again: %+v", executor.execCalls)
	}
	executor.responses[0].rows = [][]any{{int64(4), "new", "hidden"}}
	fresh := makeRunner()
	if err := fresh.batchQueryRelationModels(field, []any{int64(4)}, 0); err != nil {
		t.Fatal(err)
	}
	if len(executor.execCalls) != 4 || fresh.getCachedRelationModel(cached.GetPkgKey(), int64(4)) == nil {
		t.Fatal("cache escaped query lifetime")
	}
}
