// Package metrics provides utility functions for metric key handling.
package metrics

import (
	"strings"
)

const (
	keySeparator     = "|"
	keyEscapeChar    = "\\"
	escapedSeparator = keyEscapeChar + keySeparator
	escapedEscape    = keyEscapeChar + keyEscapeChar
)

// BuildKey builds a metric key from components, escaping any separators.
func BuildKey(components ...string) string {
	var escaped []string
	for _, comp := range components {
		escaped = append(escaped, escapeComponent(comp))
	}
	return strings.Join(escaped, keySeparator)
}

// ParseKey parses a metric key into its original components.
func ParseKey(key string) []string {
	if key == "" {
		return []string{}
	}

	var components []string
	var current strings.Builder
	escaping := false

	for _, ch := range key {
		if escaping {
			if ch == '|' || ch == '\\' {
				current.WriteRune(ch)
			} else {
				// Invalid escape sequence, treat backslash as literal
				current.WriteRune('\\')
				current.WriteRune(ch)
			}
			escaping = false
		} else if ch == '\\' {
			escaping = true
		} else if ch == '|' {
			components = append(components, current.String())
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	if escaping {
		// Trailing backslash, treat as literal
		current.WriteRune('\\')
	}

	if current.Len() > 0 || strings.HasSuffix(key, keySeparator) {
		components = append(components, current.String())
	}

	return components
}

// escapeComponent escapes separator and escape characters in a component.
func escapeComponent(component string) string {
	if !strings.ContainsAny(component, keySeparator+keyEscapeChar) {
		return component
	}

	var result strings.Builder
	for _, ch := range component {
		if ch == '\\' {
			result.WriteString(escapedEscape)
		} else if ch == '|' {
			result.WriteString(escapedSeparator)
		} else {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// ValidateComponent validates that a component doesn't contain unescaped separators.
func ValidateComponent(component string) bool {
	escaping := false
	for _, ch := range component {
		if escaping {
			escaping = false
		} else if ch == '\\' {
			escaping = true
		} else if ch == '|' {
			return false
		}
	}
	return !escaping
}
