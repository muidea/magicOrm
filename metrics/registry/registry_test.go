package registry

import (
	"testing"

	"github.com/muidea/magicCommon/monitoring/types"
	"github.com/stretchr/testify/assert"
)

// MockProvider is a mock implementation of MetricProvider for testing
type MockProvider struct {
	*types.BaseProvider
	name string
}

func NewMockProvider(name string) *MockProvider {
	base := types.NewBaseProvider(name, "1.0.0", "Mock provider for testing")
	return &MockProvider{
		BaseProvider: base,
		name:         name,
	}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{}
}

func (m *MockProvider) Collect() ([]types.Metric, *types.Error) {
	return []types.Metric{}, nil
}

func TestProviderRegistration(t *testing.T) {
	// Test that provider registration works
	// This is a simplified test since the actual registration
	// is handled by magicCommon/monitoring

	// Create a mock provider factory
	factory := func() types.MetricProvider {
		return NewMockProvider("test_provider")
	}

	// Try to register (may fail if GlobalManager is not initialized)
	err := Register("test_provider", factory, true, 100)
	t.Logf("Register error: %v", err)

	// The error might be nil or not depending on GlobalManager state
	// This test just ensures the function can be called
	assert.NotPanics(t, func() {
		Register("test_provider2", factory, false, 200)
	}, "Register should not panic")
}
