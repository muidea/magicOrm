package metrics

import (
	"container/list"
	"time"
)

const (
	DefaultMaxDurationKeys    = 1000
	DefaultMaxDurationSamples = 1000
)

type DurationKeyTracker struct {
	order *list.List
	nodes map[string]*list.Element
}

func NewDurationKeyTracker() *DurationKeyTracker {
	return &DurationKeyTracker{
		order: list.New(),
		nodes: map[string]*list.Element{},
	}
}

func (t *DurationKeyTracker) Len() int {
	if t == nil || t.order == nil {
		return 0
	}

	return t.order.Len()
}

func (t *DurationKeyTracker) Keys() []string {
	if t == nil || t.order == nil {
		return nil
	}

	keys := make([]string, 0, t.order.Len())
	for elem := t.order.Front(); elem != nil; elem = elem.Next() {
		keys = append(keys, elem.Value.(string))
	}

	return keys
}

func (t *DurationKeyTracker) Track(key string, maxKeys int) (evictedKey string, evicted bool) {
	if t == nil || t.order == nil || maxKeys <= 0 {
		return
	}

	if elem := t.nodes[key]; elem != nil {
		t.order.MoveToBack(elem)
		return
	}

	if t.order.Len() >= maxKeys {
		oldest := t.order.Front()
		if oldest != nil {
			evictedKey = oldest.Value.(string)
			delete(t.nodes, evictedKey)
			t.order.Remove(oldest)
			evicted = true
		}
	}

	t.nodes[key] = t.order.PushBack(key)
	return
}

// RecordDurationSample stores a duration sample with optional bounded key eviction and per-key sample caps.
func RecordDurationSample(
	store map[string][]time.Duration,
	tracker *DurationKeyTracker,
	maxKeys int,
	maxSamples int,
	key string,
	duration time.Duration,
) {
	if store == nil {
		return
	}
	if maxSamples <= 0 {
		maxSamples = DefaultMaxDurationSamples
	}

	if _, found := store[key]; !found {
		store[key] = make([]time.Duration, 0, maxSamples)

		if tracker != nil {
			if evictedKey, evicted := tracker.Track(key, maxKeys); evicted {
				delete(store, evictedKey)
			}
		}
	} else if tracker != nil {
		tracker.Track(key, maxKeys)
	}

	durations := store[key]
	if len(durations) >= maxSamples {
		newDurations := make([]time.Duration, maxSamples-1, maxSamples)
		copy(newDurations, durations[1:])
		durations = newDurations
	}
	store[key] = append(durations, duration)
}

// AverageDurationSeconds returns the average duration in seconds.
func AverageDurationSeconds(durations []time.Duration) float64 {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, duration := range durations {
		total += duration
	}
	return total.Seconds() / float64(len(durations))
}
