package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildKey(t *testing.T) {
	tests := []struct {
		name       string
		components []string
		expected   string
	}{
		{
			name:       "simple components",
			components: []string{"insert", "user", "success"},
			expected:   "insert|user|success",
		},
		{
			name:       "component with separator",
			components: []string{"validate", "user_model", "insert", "success"},
			expected:   "validate|user_model|insert|success",
		},
		{
			name:       "component with escape character",
			components: []string{"operation", "test\\backslash", "success"},
			expected:   "operation|test\\\\backslash|success",
		},
		{
			name:       "component with both separator and escape",
			components: []string{"query", "user|model", "error"},
			expected:   "query|user\\|model|error",
		},
		{
			name:       "empty component",
			components: []string{"operation", "", "status"},
			expected:   "operation||status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildKey(tt.components...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected []string
	}{
		{
			name:     "simple key",
			key:      "insert|user|success",
			expected: []string{"insert", "user", "success"},
		},
		{
			name:     "key with escaped separator",
			key:      "query|user\\|model|error",
			expected: []string{"query", "user|model", "error"},
		},
		{
			name:     "key with escaped escape character",
			key:      "operation|test\\\\backslash|success",
			expected: []string{"operation", "test\\backslash", "success"},
		},
		{
			name:     "key with trailing separator",
			key:      "operation|model|",
			expected: []string{"operation", "model", ""},
		},
		{
			name:     "empty key",
			key:      "",
			expected: []string{},
		},
		{
			name:     "key with only separators",
			key:      "||",
			expected: []string{"", "", ""},
		},
		{
			name:     "key with invalid escape sequence",
			key:      "test\\x|value",
			expected: []string{"test\\x", "value"},
		},
		{
			name:     "key with trailing backslash",
			key:      "test\\",
			expected: []string{"test\\"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateComponent(t *testing.T) {
	tests := []struct {
		name      string
		component string
		expected  bool
	}{
		{
			name:      "valid component",
			component: "user_model",
			expected:  true,
		},
		{
			name:      "component with separator",
			component: "user|model",
			expected:  false,
		},
		{
			name:      "component with escaped separator",
			component: "user\\|model",
			expected:  true,
		},
		{
			name:      "component with escape character",
			component: "test\\backslash",
			expected:  true,
		},
		{
			name:      "component with invalid escape",
			component: "test\\",
			expected:  false,
		},
		{
			name:      "empty component",
			component: "",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateComponent(tt.component)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		components []string
	}{
		{
			name:       "simple round trip",
			components: []string{"insert", "user", "success"},
		},
		{
			name:       "with special characters",
			components: []string{"query", "user\\|model", "test\\\\backslash", "error"},
		},
		{
			name:       "empty components",
			components: []string{"", "value", ""},
		},
		{
			name:       "multiple separators",
			components: []string{"a\\|b", "c\\|d", "e\\\\f"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build key from components
			key := BuildKey(tt.components...)

			// Parse key back to components
			parsed := ParseKey(key)

			// Verify round trip
			assert.Equal(t, tt.components, parsed)

			// Verify each component is valid
			for _, comp := range tt.components {
				assert.True(t, ValidateComponent(comp), "component should be valid: %s", comp)
			}
		})
	}
}
