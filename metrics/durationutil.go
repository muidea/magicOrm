package metrics

import "time"

const (
	DefaultMaxDurationKeys    = 1000
	DefaultMaxDurationSamples = 1000
)

// RecordDurationSample stores a duration sample with optional LRU key eviction and per-key sample caps.
func RecordDurationSample(
	store map[string][]time.Duration,
	lru *[]string,
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

		if lru != nil && maxKeys > 0 {
			if len(*lru) >= maxKeys {
				oldestKey := (*lru)[0]
				delete(store, oldestKey)
				*lru = (*lru)[1:]
			}
			*lru = append(*lru, key)
		}
	} else if lru != nil {
		for idx, existingKey := range *lru {
			if existingKey != key {
				continue
			}
			*lru = append((*lru)[:idx], (*lru)[idx+1:]...)
			*lru = append(*lru, key)
			break
		}
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
