package orm

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
	"github.com/muidea/magicOrm/metrics"
)

func TestORMMetricProviderCollect(t *testing.T) {
	collector := NewORMMetricsCollector()
	collector.RecordTransaction("begin", true)
	collector.RecordCacheHit()
	collector.RecordCacheMiss()
	collector.UpdateActiveConnections(5)
	collector.recordDurationWithLRU("insert|metrics.orm/User|success", time.Second)
	collector.operationCounters["insert|metrics.orm/User|success"] = 2
	collector.errorCounters["query|metrics.orm/User|database"] = 1

	provider := NewORMMetricProvider(collector)
	definitions := provider.Metrics()
	if len(definitions) != 6 {
		t.Fatalf("unexpected metric definition count: %d", len(definitions))
	}
	for _, definition := range definitions {
		if err := definition.Validate(); err != nil {
			t.Fatalf("metric definition should be valid: %v", err)
		}
		if definition.ConstLabels["version"] != "1.0.0" {
			t.Fatalf("expected version label, got %+v", definition.ConstLabels)
		}
		if definition.ConstLabels["component"] != "orm" {
			t.Fatalf("expected orm component label, got %+v", definition.ConstLabels)
		}
	}

	metrics, err := provider.Collect()
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}
	if len(metrics) == 0 {
		t.Fatal("expected collected metrics")
	}

	var foundConnections bool
	for _, metric := range metrics {
		if metric.Name == "magicorm_orm_active_connections" && metric.Value == 5 {
			foundConnections = true
		}
	}
	if !foundConnections {
		t.Fatalf("expected active connections metric, got %+v", metrics)
	}

	if provider.Name() != "magicorm_orm" {
		t.Fatalf("unexpected provider name: %s", provider.Name())
	}
	if initErr := provider.Init(nil); initErr != nil {
		t.Fatalf("expected init success, got %v", initErr)
	}
	if shutdownErr := provider.Shutdown(); shutdownErr != nil {
		t.Fatalf("expected shutdown success, got %v", shutdownErr)
	}
}

func TestORMMetricProviderWithoutCollector(t *testing.T) {
	provider := NewORMMetricProvider(nil)

	metrics, err := provider.Collect()
	if err != nil {
		t.Fatalf("expected empty collect success without collector, got %v", err)
	}
	if len(metrics) != 0 {
		t.Fatalf("expected no metrics without collector, got %+v", metrics)
	}

	if got := parseKey("a|b|c"); len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("unexpected parseKey result: %v", got)
	}

	health := provider.BaseProvider.GetMetadata().HealthStatus
	if health != types.ProviderUnknown && health != types.ProviderHealthy {
		t.Fatalf("unexpected provider health status: %s", health)
	}
}

func TestORMMetricProviderCollectSkipsMalformedKeys(t *testing.T) {
	collector := NewORMMetricsCollector()
	collector.operationCounters["invalid"] = 1
	collector.errorCounters["broken"] = 2
	collector.operationDurations["bad"] = []time.Duration{time.Second}
	collector.transactionCounters["oops"] = 3

	validKey := metrics.BuildKey("insert", "metrics.orm/User", "success")
	collector.operationCounters[validKey] = 1
	collector.operationDurations[validKey] = []time.Duration{100 * time.Millisecond, 300 * time.Millisecond}
	collector.UpdateActiveConnections(2)

	metricsList, err := NewORMMetricProvider(collector).Collect()
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}

	foundDuration := false
	for _, metric := range metricsList {
		if metric.Name == "magicorm_orm_operation_duration_seconds" {
			foundDuration = true
			if metric.Labels["operation"] != "insert" {
				t.Fatalf("unexpected operation label: %+v", metric.Labels)
			}
			if metric.Value < 0.19 || metric.Value > 0.21 {
				t.Fatalf("unexpected duration value: %v", metric.Value)
			}
		}
	}
	if !foundDuration {
		t.Fatalf("expected valid duration metric in %+v", metricsList)
	}
}
