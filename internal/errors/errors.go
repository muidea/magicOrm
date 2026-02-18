// Package errors provides enhanced error handling utilities for the magicOrm project
package errors

import (
	"fmt"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
)

// LogError logs an error with context information
// This is a generic version that can be used throughout the project
func LogError(component, operation string, err error) {
	if err == nil {
		return
	}

	// Check for typed nil (*cd.Error that is nil)
	if cdErr, ok := err.(*cd.Error); ok && cdErr == nil {
		return
	}

	var errMsg string
	if cdErr, ok := err.(*cd.Error); ok {
		errMsg = cdErr.Error()
	} else {
		errMsg = err.Error()
	}

	if operation != "" {
		log.Errorf("%s failed, %s error:%v", component, operation, errMsg)
	} else {
		log.Errorf("%s failed, error:%v", component, errMsg)
	}
}

// LogErrorf logs an error with formatted context
func LogErrorf(component, format string, args ...any) {
	log.Errorf("%s %s", component, fmt.Sprintf(format, args...))
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with formatted context
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// MultiError represents multiple errors
type MultiError struct {
	Errors []error
}

// Error returns a concatenated error message
func (m *MultiError) Error() string {
	if len(m.Errors) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("multiple errors:")
	for i, err := range m.Errors {
		sb.WriteString(fmt.Sprintf("\n  [%d] %v", i+1, err))
	}
	return sb.String()
}

// Add adds an error to the MultiError
func (m *MultiError) Add(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

// HasErrors returns true if there are any errors
func (m *MultiError) HasErrors() bool {
	return len(m.Errors) > 0
}

// NewMultiError creates a new MultiError
func NewMultiError() *MultiError {
	return &MultiError{
		Errors: make([]error, 0),
	}
}

// CheckAndLog checks if an error exists and logs it
func CheckAndLog(component, operation string, err error) bool {
	if err != nil {
		LogError(component, operation, err)
		return true
	}
	return false
}

// Must logs and returns if error is not nil
// Useful for the common pattern: if err != nil { log.Errorf(...); return }
func Must(component, operation string, err error) error {
	if err != nil {
		LogError(component, operation, err)
	}
	return err
}

// MustReturn logs and returns if error is not nil
// Returns true if error occurred (for use in if statements)
func MustReturn(component, operation string, err error) bool {
	if err != nil {
		// Check for typed nil (*cd.Error that is nil)
		if cdErr, ok := err.(*cd.Error); ok && cdErr == nil {
			return false
		}
		LogError(component, operation, err)
		return true
	}
	return false
}

// HandleError provides a more flexible error handling with custom action
func HandleError(component, operation string, err error, action func()) {
	if err != nil {
		LogError(component, operation, err)
		if action != nil {
			action()
		}
	}
}

// CDError creates a new cd.Error with logging
func CDError(code cd.Code, message, component, operation string) *cd.Error {
	err := cd.NewError(code, message)
	LogError(component, operation, err)
	return err
}
