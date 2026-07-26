package postgres

import "testing"

func TestSplitDatabaseAndSchema(t *testing.T) {
	for _, test := range []struct {
		value, database, schema string
	}{
		{"magicplatform_db/core", "magicplatform_db", "core"},
		{"magicplatform_db", "magicplatform_db", "public"},
		{" database / file ", "database", "file"},
	} {
		database, schema := splitDatabaseAndSchema(test.value)
		if database != test.database || schema != test.schema {
			t.Fatalf("splitDatabaseAndSchema(%q) = %q, %q", test.value, database, schema)
		}
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
