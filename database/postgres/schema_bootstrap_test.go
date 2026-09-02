package postgres

import "testing"

func TestConfigKeepsDatabaseAndSchemaSeparate(t *testing.T) {
	config := NewConfig("postgres:5432", "magicplatform_db", "core", "app", "secret")
	if config.Database() != "magicplatform_db" || config.Schema() != "core" {
		t.Fatalf("unexpected database/schema: %q/%q", config.Database(), config.Schema())
	}
	if got := config.GetDsn(); got != "postgres://app:secret@postgres:5432/magicplatform_db?sslmode=disable&options=-c%20search_path=core" {
		t.Fatalf("unexpected postgres DSN: %q", got)
	}
}

func TestValidSchemaName(t *testing.T) {
	if !validSchemaName("core_1") {
		t.Fatal("valid schema name was rejected")
	}
	for _, value := range []string{"", "core-name", "core;drop"} {
		if validSchemaName(value) {
			t.Fatalf("invalid schema name %q was accepted", value)
		}
	}
}
