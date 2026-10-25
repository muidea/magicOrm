package test

import (
	"os"
	"strings"
	"testing"
)

// This package contains destructive database integration tests. No connection
// is attempted without explicit authorization and a dedicated namespace.
func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("MAGICORM_INTEGRATION_ALLOW") != "1" {
		t.Skip("set MAGICORM_INTEGRATION_ALLOW=1 and configure a dedicated magicorm_test_ database")
	}
	if config.Server() == "" || config.Username() == "" || config.Password() == "" || !strings.HasPrefix(config.Database(), "magicorm_test_") || config.Schema() == "" || config.Schema() == "public" {
		t.Fatal("integration tests require explicit credentials and a dedicated magicorm_test_ database/non-public schema")
	}
}
