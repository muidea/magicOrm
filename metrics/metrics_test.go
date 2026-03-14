package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLabels(t *testing.T) {
	labels := DefaultLabels()
	assert.Equal(t, map[string]string{
		"component": "magicorm",
		"version":   "1.0.0",
	}, labels)
}

func TestMergeLabels(t *testing.T) {
	merged := MergeLabels(
		map[string]string{"component": "magicorm", "env": "dev"},
		nil,
		map[string]string{"env": "test", "version": "1.0.0"},
	)

	assert.Equal(t, map[string]string{
		"component": "magicorm",
		"env":       "test",
		"version":   "1.0.0",
	}, merged)
}
