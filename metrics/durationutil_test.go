package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecordDurationSample(t *testing.T) {
	store := map[string][]time.Duration{}
	lru := make([]string, 0, 2)

	RecordDurationSample(store, &lru, 2, 2, "a", 10*time.Millisecond)
	RecordDurationSample(store, &lru, 2, 2, "b", 20*time.Millisecond)
	RecordDurationSample(store, &lru, 2, 2, "a", 30*time.Millisecond)
	RecordDurationSample(store, &lru, 2, 2, "c", 40*time.Millisecond)

	assert.NotContains(t, store, "b")
	assert.Equal(t, []string{"a", "c"}, lru)
	assert.Equal(t, []time.Duration{10 * time.Millisecond, 30 * time.Millisecond}, store["a"])
	assert.Equal(t, []time.Duration{40 * time.Millisecond}, store["c"])
}

func TestRecordDurationSampleEdgeCases(t *testing.T) {
	RecordDurationSample(nil, nil, 1, 1, "ignored", time.Second)

	store := map[string][]time.Duration{}

	RecordDurationSample(store, nil, 0, 0, "a", 10*time.Millisecond)
	RecordDurationSample(store, nil, 0, 1, "a", 20*time.Millisecond)
	RecordDurationSample(store, nil, 0, 1, "a", 30*time.Millisecond)

	assert.Len(t, store["a"], 1)
	assert.Equal(t, 30*time.Millisecond, store["a"][0])
}

func TestAverageDurationSeconds(t *testing.T) {
	assert.Equal(t, 0.0, AverageDurationSeconds(nil))
	assert.InDelta(t, 0.2, AverageDurationSeconds([]time.Duration{
		100 * time.Millisecond,
		300 * time.Millisecond,
	}), 0.001)
}
