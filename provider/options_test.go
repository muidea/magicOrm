package provider

import (
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/stretchr/testify/assert"
)

// TestNewLocalProviderWithOptions tests the functional options pattern for local provider
func TestNewLocalProviderWithOptions(t *testing.T) {
	// Test with default options
	provider1 := NewLocalProviderWithOptions("testOwner1")
	assert.NotNil(t, provider1)
	assert.Equal(t, "testOwner1", provider1.Owner())

	// Test with custom validator
	customValidator := &mockValidator{}
	provider2 := NewLocalProviderWithOptions("testOwner2", WithValueValidator(customValidator))
	assert.NotNil(t, provider2)
	assert.Equal(t, "testOwner2", provider2.Owner())

	// Test with custom cache
	customCache := models.NewCache()
	provider3 := NewLocalProviderWithOptions("testOwner3",
		WithValueValidator(customValidator),
		WithModelCache(customCache),
	)
	assert.NotNil(t, provider3)
	assert.Equal(t, "testOwner3", provider3.Owner())
}

// TestNewRemoteProviderWithOptions tests the functional options pattern for remote provider
func TestNewRemoteProviderWithOptions(t *testing.T) {
	// Test with default options
	provider1 := NewRemoteProviderWithOptions("testOwner1")
	assert.NotNil(t, provider1)
	assert.Equal(t, "testOwner1", provider1.Owner())

	// Test with custom validator
	customValidator := &mockValidator{}
	provider2 := NewRemoteProviderWithOptions("testOwner2", WithValueValidator(customValidator))
	assert.NotNil(t, provider2)
	assert.Equal(t, "testOwner2", provider2.Owner())
}

// TestBackwardCompatibility tests that the old constructors still work
func TestBackwardCompatibility(t *testing.T) {
	// Test old local provider constructor
	oldLocalProvider := NewLocalProvider("oldOwner", nil)
	assert.NotNil(t, oldLocalProvider)
	assert.Equal(t, "oldOwner", oldLocalProvider.Owner())

	// Test old remote provider constructor
	oldRemoteProvider := NewRemoteProvider("oldOwner", nil)
	assert.NotNil(t, oldRemoteProvider)
	assert.Equal(t, "oldOwner", oldRemoteProvider.Owner())
}

// mockValidator is a mock implementation of ValueValidator for testing
type mockValidator struct{}

func (m *mockValidator) Register(k models.Key, fn models.ValidatorFunc) {
	// Mock implementation
}

func (m *mockValidator) ValidateValue(val any, directives []models.Directive) error {
	// Mock implementation - always return nil
	return nil
}
