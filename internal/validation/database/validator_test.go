package database

import (
	"testing"

	"github.com/muidea/magicOrm/models"
)

type dbDirective struct {
	key models.Key
}

func (d dbDirective) Key() models.Key { return d.key }
func (d dbDirective) Args() []string  { return nil }
func (d dbDirective) HasArgs() bool   { return false }

type dbConstraints struct {
	directives []models.Directive
}

func (c dbConstraints) Has(key models.Key) bool {
	_, ok := c.Get(key)
	return ok
}
func (c dbConstraints) Get(key models.Key) (models.Directive, bool) {
	for _, directive := range c.directives {
		if directive.Key() == key {
			return directive, true
		}
	}
	return nil, false
}
func (c dbConstraints) Directives() []models.Directive { return c.directives }

func TestDatabaseValidator(t *testing.T) {
	validator := NewDatabaseValidator()
	constraints := dbConstraints{directives: []models.Directive{dbDirective{key: models.KeyRequired}}}

	if err := validator.ValidateDatabaseConstraints(nil, "name", constraints, "postgresql"); err == nil {
		t.Fatal("expected required database constraint to reject nil value")
	}
	if err := validator.ValidateDatabaseConstraints("value", "name", constraints, "postgresql"); err != nil {
		t.Fatalf("expected valid database constraint, got %v", err)
	}
	if err := validator.ValidateDatabaseConstraints(nil, "name", constraints, ""); err != nil {
		t.Fatalf("expected empty db type to skip validation, got %v", err)
	}

	dbRules := validator.GetDatabaseConstraints(constraints)
	if len(dbRules) != 1 || dbRules[0] != "NOT NULL" {
		t.Fatalf("unexpected database constraints: %v", dbRules)
	}
	if len(validator.GetDatabaseConstraints(nil)) != 0 {
		t.Fatal("expected nil constraints to yield empty database rules")
	}

	converted, err := validator.ConvertToDatabaseValue("value")
	if err != nil || converted != "value" {
		t.Fatalf("unexpected database conversion result: %v %v", converted, err)
	}
}
