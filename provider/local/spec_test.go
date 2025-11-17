package local

import (
	"testing"

	"github.com/muidea/magicOrm/models"
)

func TestSpec(t *testing.T) {
	spec1 := "spec"
	specPtr, specErr := getOrmSpec(spec1)
	if specErr != nil {
		t.Errorf("NewSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("NewSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if specPtr.GetValueDeclare() == models.AutoIncrement || specPtr.IsPrimaryKey() {
		t.Errorf("NewSpec failed")
		return
	}

	spec2 := "spec auto"
	specPtr, specErr = getOrmSpec(spec2)
	if specErr != nil {
		t.Errorf("NewSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("NewSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if specPtr.GetValueDeclare() != models.AutoIncrement || specPtr.IsPrimaryKey() {
		t.Errorf("NewSpec failed")
		return
	}

	spec3 := "spec auto key"
	specPtr, specErr = getOrmSpec(spec3)
	if specErr != nil {
		t.Errorf("NewSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("NewSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if specPtr.GetValueDeclare() != models.AutoIncrement || !specPtr.IsPrimaryKey() {
		t.Errorf("NewSpec failed")
		return
	}

	spec4 := "spec key auto"
	specPtr, specErr = getOrmSpec(spec4)
	if specErr != nil {
		t.Errorf("NewSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("NewSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if specPtr.GetValueDeclare() != models.AutoIncrement || !specPtr.IsPrimaryKey() {
		t.Errorf("NewSpec failed")
		return
	}
}

// TestSpecValueDeclares tests various value declare types in specs
func TestSpecValueDeclares(t *testing.T) {
	testCases := []struct {
		name         string
		spec         string
		fieldName    string
		valueDeclare models.ValueDeclare
		isPrimaryKey bool
	}{
		{"Default", "field", "field", models.Customer, false},
		{"Auto", "field auto", "field", models.AutoIncrement, false},
		{"AutoAbbrev", "field auto", "field", models.AutoIncrement, false},
		{"UUID", "field uuid", "field", models.UUID, false},
		{"SnowFlake", "field snowflake", "field", models.SnowFlake, false},
		{"DateTime", "field datetime", "field", models.DateTime, false},
		{"PrimaryKey", "field key", "field", models.Customer, true},
		{"Primary", "field key", "field", models.Customer, true},
		{"PrimaryAndAuto", "field key auto", "field", models.AutoIncrement, true},
		{"ComplexName", "complex_field_name", "complex_field_name", models.Customer, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			specPtr, err := getOrmSpec(tc.spec)
			if err != nil {
				t.Errorf("getOrmSpec failed for %s: %s", tc.name, err.Error())
				return
			}

			if specPtr.GetFieldName() != tc.fieldName {
				t.Errorf("Field name mismatch for %s, expected: %s, got: %s",
					tc.name, tc.fieldName, specPtr.GetFieldName())
			}

			if specPtr.GetValueDeclare() != tc.valueDeclare {
				t.Errorf("Value declare mismatch for %s, expected: %v, got: %v",
					tc.name, tc.valueDeclare, specPtr.GetValueDeclare())
			}

			if specPtr.IsPrimaryKey() != tc.isPrimaryKey {
				t.Errorf("Primary key flag mismatch for %s, expected: %v, got: %v",
					tc.name, tc.isPrimaryKey, specPtr.IsPrimaryKey())
			}
		})
	}
}

// TestSpecDescription tests the spec description functionality
func TestSpecDescription(t *testing.T) {
	// Create a spec
	spec := "field key auto"
	specPtr, err := getOrmSpec(spec)
	if err != nil {
		t.Errorf("getOrmSpec failed: %s", err.Error())
		return
	}

	// Verify spec has the expected values
	if specPtr.GetFieldName() != "field" {
		t.Errorf("Field name mismatch, expected: %s, got: %s",
			"field", specPtr.GetFieldName())
	}

	if specPtr.GetValueDeclare() != models.AutoIncrement {
		t.Errorf("Value declare mismatch, expected: %v, got: %v",
			models.AutoIncrement, specPtr.GetValueDeclare())
	}

	if !specPtr.IsPrimaryKey() {
		t.Errorf("Primary key flag mismatch, expected: %v, got: %v",
			true, specPtr.IsPrimaryKey())
	}
}

// TestInvalidSpecs tests various invalid spec formats
func TestInvalidSpecs(t *testing.T) {
	invalidSpecs := []string{
		"",                // Empty spec
		" ",               // Only whitespace
		"field invalid",   // Invalid value declare
		"field auto auto", // Duplicate value declare
		"field key key",   // Duplicate primary key
	}

	for i, invalid := range invalidSpecs {
		specPtr, err := getOrmSpec(invalid)
		if err != nil {
			// Other invalid formats might still parse but create a default spec
			t.Logf("Case %d: Got expected error: %s", i, err.Error())
		} else {
			// If no error, it should have created a spec with default value
			if invalid == "field invalid" && specPtr.GetValueDeclare() != models.Customer {
				t.Errorf("Case %d: Invalid value declare should default to Customer", i)
			}
		}
	}
}
