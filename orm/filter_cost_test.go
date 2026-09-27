package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/metrics"
	ormmetrics "github.com/muidea/magicOrm/metrics/orm"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
)

func filterCostFixture(t testing.TB) (*impl, *remote.ObjectFilter, *fakeExecutor) {
	t.Helper()
	obj := &remote.Object{Name: "filterCost", PkgPath: "/cost"}
	row := []any{int64(1)}
	obj.Fields = append(obj.Fields, &remote.Field{Name: "id", Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}, Spec: &remote.SpecImpl{FieldName: "id", PrimaryKey: true, ViewDeclare: []models.ViewDeclare{models.DetailView}}})
	for i := 0; i < 15; i++ {
		name := fmt.Sprintf("field%d", i)
		obj.Fields = append(obj.Fields, &remote.Field{Name: name, Type: &remote.TypeImpl{Name: "string", Value: models.TypeStringValue}, Spec: &remote.SpecImpl{FieldName: name, ViewDeclare: []models.ViewDeclare{models.DetailView}}})
		row = append(row, "value")
	}
	p := provider.NewRemoteProvider("cost", nil)
	if _, err := p.RegisterModel(obj); err != nil {
		t.Fatal(err)
	}
	f, err := p.GetEntityFilter(obj, models.DetailView)
	if err != nil {
		t.Fatal(err)
	}
	exec := &fakeExecutor{responses: []fakeQueryResponse{
		{match: func(q string, _ []any) bool { return strings.Contains(strings.ToLower(q), "count(") }, rows: [][]any{{sql.NullInt64{Int64: 1, Valid: true}}}},
		{match: func(string, []any) bool { return true }, rows: [][]any{row}},
	}}
	return &impl{context: context.Background(), executor: exec, modelProvider: p, modelCodec: codec.New(p, "")}, f.(*remote.ObjectFilter), exec
}

func BenchmarkFilterReadModels(b *testing.B) {
	previous := ormMetricCollector
	ormMetricCollector = ormmetrics.NewORMMetricsCollector()
	b.Cleanup(func() { ormMetricCollector = previous })
	orm, filter, executor := filterCostFixture(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.execCalls = executor.execCalls[:0]
		if _, err := orm.Count(filter); err != nil {
			b.Fatal(err)
		}
		if rows, err := orm.BatchQuery(filter); err != nil || len(rows) != 1 {
			b.Fatalf("rows=%d error=%v", len(rows), err)
		}
	}
}

type trackedMaskFilter struct {
	*remote.ObjectFilter
	maskCalls int
}

func (f *trackedMaskFilter) MaskModel() models.Model {
	f.maskCalls++
	return f.ObjectFilter.MaskModel()
}

func TestFilterMetricsReuseResolvedModel(t *testing.T) {
	previous := ormMetricCollector
	ormMetricCollector = ormmetrics.NewORMMetricsCollector()
	t.Cleanup(func() { ormMetricCollector = previous })
	orm, filter, _ := filterCostFixture(t)
	tracked := &trackedMaskFilter{ObjectFilter: filter}
	if count, err := orm.Count(tracked); err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	if tracked.maskCalls != 1 {
		t.Fatalf("Count metrics rebuilt the model: mask calls=%d", tracked.maskCalls)
	}
	tracked.maskCalls = 0
	if rows, err := orm.BatchQuery(tracked); err != nil || len(rows) != 1 {
		t.Fatalf("rows=%d err=%v", len(rows), err)
	}
	if tracked.maskCalls != 0 {
		t.Fatalf("BatchQuery metrics rebuilt the resolved response: mask calls=%d", tracked.maskCalls)
	}
	counters := ormMetricCollector.GetOperationCounters()
	for _, operation := range []metrics.OperationType{metrics.OperationCount, metrics.OperationBatch} {
		if counters[metrics.BuildKey(string(operation), "/cost/filterCost", "success")] != 1 {
			t.Fatalf("query metric identity/count changed: %v", counters)
		}
	}
}
