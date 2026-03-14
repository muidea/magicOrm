package orm

import (
	"context"
	"testing"

	"github.com/muidea/magicOrm/database/postgres"
)

func TestProviderGuardClauses(t *testing.T) {
	cfg := postgres.NewConfig("127.0.0.1:5432", "demo", "user", "password")

	if _, err := NewOrm(nil, cfg, "tenant"); err == nil {
		t.Fatal("expected NewOrm to reject nil provider")
	}

	if _, err := GetOrm(context.Background(), nil, "tenant"); err == nil {
		t.Fatal("expected GetOrm to reject nil provider")
	}
}
