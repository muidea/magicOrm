package orm

import (
	"errors"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

type metricModel struct{ name string }

func (m *metricModel) GetName() string                      { return m.name }
func (m *metricModel) GetShowName() string                  { return m.name }
func (m *metricModel) GetPkgPath() string                   { return "metrics.orm" }
func (m *metricModel) GetPkgKey() string                    { return m.GetPkgPath() + "/" + m.name }
func (m *metricModel) GetDescription() string               { return m.name }
func (m *metricModel) GetFields() models.Fields             { return nil }
func (m *metricModel) SetFieldValue(string, any) *cd.Error  { return nil }
func (m *metricModel) SetPrimaryFieldValue(any) *cd.Error   { return nil }
func (m *metricModel) GetPrimaryField() models.Field        { return nil }
func (m *metricModel) GetField(string) models.Field         { return nil }
func (m *metricModel) Interface(bool) any                   { return nil }
func (m *metricModel) Copy(models.ViewDeclare) models.Model { return m }
func (m *metricModel) Reset()                               {}

type panicError struct{}

func (panicError) Error() string { panic("boom") }

func TestORMMetricsCollector(t *testing.T) {
	collector := NewORMMetricsCollector()
	collector.maxDurationKeys = 2

	model := &metricModel{name: "User"}
	collector.RecordOperation("insert", model, time.Second, nil)
	collector.RecordOperation("query", model, time.Millisecond*50, errors.New("database timeout"))
	collector.RecordTransaction("begin", true)
	collector.RecordTransaction("commit", false)
	collector.RecordCacheHit()
	collector.RecordCacheMiss()
	collector.UpdateActiveConnections(3)

	if len(collector.GetOperationCounters()) != 2 {
		t.Fatalf("unexpected operation counters: %+v", collector.GetOperationCounters())
	}
	if len(collector.GetErrorCounters()) != 1 {
		t.Fatalf("unexpected error counters: %+v", collector.GetErrorCounters())
	}
	if hits, misses := collector.GetCacheStats(); hits != 1 || misses != 1 {
		t.Fatalf("unexpected cache stats: hits=%d misses=%d", hits, misses)
	}
	if collector.GetActiveConnections() != 3 {
		t.Fatalf("unexpected active connections: %d", collector.GetActiveConnections())
	}

	collector.recordDurationWithLRU("k1", time.Millisecond)
	collector.recordDurationWithLRU("k2", time.Millisecond)
	collector.recordDurationWithLRU("k3", time.Millisecond)
	if len(collector.GetOperationDurations()) != 2 {
		t.Fatalf("expected LRU eviction of duration keys, got %+v", collector.GetOperationDurations())
	}

	if got := collector.classifyError(errors.New("constraint violation")); got != "constraint" {
		t.Fatalf("unexpected error classification: %s", got)
	}
	if got := collector.classifyError(panicError{}); got != "unknown" {
		t.Fatalf("expected panic-safe unknown classification, got %s", got)
	}

	collector.Clear()
	if len(collector.GetOperationCounters()) != 0 || len(collector.GetTransactionCounters()) != 0 {
		t.Fatal("expected collector state to be cleared")
	}
}

func TestORMMetricsCollectorTreatsNotFoundAsNonError(t *testing.T) {
	collector := NewORMMetricsCollector()
	model := &metricModel{name: "User"}

	collector.RecordOperation("query", model, time.Millisecond*10, cd.NewError(cd.NotFound, "no rows"))

	operationCounters := collector.GetOperationCounters()
	notFoundKey := "query|metrics.orm/User|not_found"
	if got := operationCounters[notFoundKey]; got != 1 {
		t.Fatalf("expected not_found operation counter to be recorded once, got %d in %+v", got, operationCounters)
	}

	if len(collector.GetErrorCounters()) != 0 {
		t.Fatalf("expected not_found to avoid error counters, got %+v", collector.GetErrorCounters())
	}
}

func TestORMMetricsCollectorClassifiesCDErrorCodes(t *testing.T) {
	collector := NewORMMetricsCollector()

	testCases := []struct {
		name string
		err  error
		want string
	}{
		{name: "validation", err: cd.NewError(cd.IllegalParam, "illegal param"), want: "validation"},
		{name: "database", err: cd.NewError(cd.DatabaseError, "db failed"), want: "database"},
		{name: "timeout", err: cd.NewError(cd.Timeout, "timeout"), want: "timeout"},
		{name: "not_found", err: cd.NewError(cd.NotFound, "missing"), want: "not_found"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := collector.classifyError(tc.err); got != tc.want {
				t.Fatalf("classifyError(%v) = %s, want %s", tc.err, got, tc.want)
			}
		})
	}
}

func TestORMMetricsCollectorTreatsTypedNilCDErrorAsSuccess(t *testing.T) {
	collector := NewORMMetricsCollector()
	model := &metricModel{name: "User"}

	var opErr *cd.Error
	collector.RecordOperation("insert", model, time.Millisecond*10, opErr)

	operationCounters := collector.GetOperationCounters()
	successKey := "insert|metrics.orm/User|success"
	if got := operationCounters[successKey]; got != 1 {
		t.Fatalf("expected typed nil error to record success once, got %d in %+v", got, operationCounters)
	}

	if len(collector.GetErrorCounters()) != 0 {
		t.Fatalf("expected typed nil error to avoid error counters, got %+v", collector.GetErrorCounters())
	}
}

func TestORMMetricsCollectorGettersReturnCopies(t *testing.T) {
	collector := NewORMMetricsCollector()
	model := &metricModel{name: "User"}

	collector.RecordOperation("insert", model, time.Millisecond*10, nil)
	collector.RecordTransaction("begin", true)

	counters := collector.GetOperationCounters()
	counters["insert|metrics.orm/User|success"] = 99

	txCounters := collector.GetTransactionCounters()
	txCounters["begin|success"] = 99

	durations := collector.GetOperationDurations()
	durations["insert|metrics.orm/User|success"][0] = time.Second

	freshCounters := collector.GetOperationCounters()
	if got := freshCounters["insert|metrics.orm/User|success"]; got != 1 {
		t.Fatalf("expected operation counter copy isolation, got %d in %+v", got, freshCounters)
	}

	freshTxCounters := collector.GetTransactionCounters()
	if got := freshTxCounters["begin|success"]; got != 1 {
		t.Fatalf("expected transaction counter copy isolation, got %d in %+v", got, freshTxCounters)
	}

	freshDurations := collector.GetOperationDurations()
	if got := freshDurations["insert|metrics.orm/User|success"][0]; got != time.Millisecond*10 {
		t.Fatalf("expected duration copy isolation, got %s in %+v", got, freshDurations)
	}
}
