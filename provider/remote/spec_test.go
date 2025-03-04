package remote

import (
	"testing"

	"github.com/muidea/magicOrm/model"
)

func TestSpecImplementation(t *testing.T) {
	// Test basic spec
	spec := &SpecImpl{
		FieldName: "testField",
	}

	// Test GetFieldName
	if spec.GetFieldName() != "testField" {
		t.Errorf("GetFieldName failed, expected 'testField', got '%s'", spec.GetFieldName())
	}

	// Test IsPrimaryKey (default)
	if spec.IsPrimaryKey() {
		t.Errorf("IsPrimaryKey failed, expected false by default, got true")
	}

	// Test GetValueDeclare (default)
	if spec.GetValueDeclare() != model.Customer {
		t.Errorf("GetValueDeclare failed, expected Customer by default, got %v", spec.GetValueDeclare())
	}

	// Test primary key spec
	pkSpec := &SpecImpl{
		FieldName:  "idField",
		PrimaryKey: true,
	}

	if !pkSpec.IsPrimaryKey() {
		t.Errorf("IsPrimaryKey failed, expected true, got false")
	}

	// Test value declare spec
	autoSpec := &SpecImpl{
		FieldName:    "autoField",
		ValueDeclare: model.AutoIncrement,
	}

	if autoSpec.GetValueDeclare() != model.AutoIncrement {
		t.Errorf("GetValueDeclare failed, expected AutoIncrement, got %v", autoSpec.GetValueDeclare())
	}

	// Test view declare spec
	viewSpec := &SpecImpl{
		FieldName:   "viewField",
		ViewDeclare: []model.ViewDeclare{model.DetailView, model.LiteView},
	}

	// Test EnableView for DetailView
	if !viewSpec.EnableView(model.DetailView) {
		t.Errorf("EnableView for DetailView failed, expected true, got false")
	}

	// Test EnableView for LiteView
	if !viewSpec.EnableView(model.LiteView) {
		t.Errorf("EnableView for LiteView failed, expected true, got false")
	}

	// Test negative view case
	nonDetailSpec := &SpecImpl{
		FieldName:   "nonDetailField",
		ViewDeclare: []model.ViewDeclare{model.LiteView},
	}

	if nonDetailSpec.EnableView(model.DetailView) {
		t.Errorf("EnableView failed, expected false for non-detail field, got true")
	}

	// Test copy
	originalSpec := &SpecImpl{
		FieldName:    "originalField",
		PrimaryKey:   true,
		ValueDeclare: model.AutoIncrement,
		ViewDeclare:  []model.ViewDeclare{model.DetailView, model.LiteView},
	}

	copiedSpec := originalSpec.copy()

	if copiedSpec.GetFieldName() != originalSpec.GetFieldName() ||
		copiedSpec.IsPrimaryKey() != originalSpec.IsPrimaryKey() ||
		copiedSpec.GetValueDeclare() != originalSpec.GetValueDeclare() {
		t.Errorf("Copy failed, basic properties don't match")
	}

	// Check view declarations
	if !copiedSpec.EnableView(model.DetailView) || !copiedSpec.EnableView(model.LiteView) {
		t.Errorf("Copy failed, view declarations don't match")
	}

	// Make sure it's a deep copy
	originalSpec.ViewDeclare = []model.ViewDeclare{model.DetailView} // Remove LiteView
	if !copiedSpec.EnableView(model.LiteView) {
		t.Errorf("Copy failed, should be a deep copy unaffected by changes to original")
	}
}

func TestOrmSpecParsing(t *testing.T) {
	tests := []struct {
		name           string
		specStr        string
		expectedName   string
		expectedPK     bool
		expectedValue  model.ValueDeclare
		expectedDetail bool
		expectedLite   bool
		expectError    bool
	}{
		{
			name:           "Empty spec",
			specStr:        "",
			expectedName:   "",
			expectedPK:     false,
			expectedValue:  model.Customer,
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "Basic field name",
			specStr:        "username",
			expectedName:   "username",
			expectedPK:     false,
			expectedValue:  model.Customer,
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "Primary key",
			specStr:        "id key",
			expectedName:   "id",
			expectedPK:     true,
			expectedValue:  model.Customer,
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "Auto-increment primary key",
			specStr:        "id auto key",
			expectedName:   "id",
			expectedPK:     true,
			expectedValue:  model.AutoIncrement,
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "UUID field",
			specStr:        "id uuid",
			expectedName:   "id",
			expectedPK:     false,
			expectedValue:  model.UUID,
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "UUID primary key",
			specStr:        "id uuid key",
			expectedName:   "id",
			expectedPK:     true,
			expectedValue:  model.UUID,
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "Snowflake field",
			specStr:        "id snowflake",
			expectedName:   "id",
			expectedPK:     false,
			expectedValue:  model.SnowFlake,
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "DateTime field",
			specStr:        "createTime datetime",
			expectedName:   "createTime",
			expectedPK:     false,
			expectedValue:  model.DateTime,
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "Multiple options",
			specStr:        "id auto key",
			expectedName:   "id",
			expectedPK:     true,
			expectedValue:  model.AutoIncrement,
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "Invalid value declare",
			specStr:        "id invalid",
			expectedName:   "id",
			expectedPK:     false,
			expectedValue:  model.Customer, // Should default to Customer
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false, // Not expecting error, just ignores invalid option
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec, err := getOrmSpec(test.specStr)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error for spec '%s', but got none", test.specStr)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for spec '%s': %v", test.specStr, err)
				return
			}

			// Check field name
			if spec.GetFieldName() != test.expectedName {
				t.Errorf("Field name mismatch, expected '%s', got '%s'",
					test.expectedName, spec.GetFieldName())
			}

			// Check primary key
			if spec.IsPrimaryKey() != test.expectedPK {
				t.Errorf("Primary key mismatch, expected %v, got %v",
					test.expectedPK, spec.IsPrimaryKey())
			}

			// Check value declare
			if spec.GetValueDeclare() != test.expectedValue {
				t.Errorf("Value declare mismatch, expected %v, got %v",
					test.expectedValue, spec.GetValueDeclare())
			}

			// Check view declarations
			if test.expectedDetail && !spec.EnableView(model.DetailView) {
				t.Errorf("Detail view mismatch, expected %v, got %v",
					test.expectedDetail, spec.EnableView(model.DetailView))
			}

			if test.expectedLite && !spec.EnableView(model.LiteView) {
				t.Errorf("Lite view mismatch, expected %v, got %v",
					test.expectedLite, spec.EnableView(model.LiteView))
			}
		})
	}
}

func TestViewSpecParsing(t *testing.T) {
	tests := []struct {
		name           string
		viewStr        string
		expectedDetail bool
		expectedLite   bool
		expectError    bool
	}{
		{
			name:           "Empty view",
			viewStr:        "",
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "Detail view only",
			viewStr:        "detail",
			expectedDetail: true,
			expectedLite:   false,
			expectError:    false,
		},
		{
			name:           "Lite view only",
			viewStr:        "lite",
			expectedDetail: false,
			expectedLite:   true,
			expectError:    false,
		},
		{
			name:           "Both views",
			viewStr:        "detail,lite",
			expectedDetail: true,
			expectedLite:   true,
			expectError:    false,
		},
		{
			name:           "Both views different order",
			viewStr:        "lite,detail",
			expectedDetail: true,
			expectedLite:   true,
			expectError:    false,
		},
		{
			name:           "With spaces",
			viewStr:        " detail , lite ",
			expectedDetail: true,
			expectedLite:   true,
			expectError:    false,
		},
		{
			name:           "Invalid view",
			viewStr:        "invalid",
			expectedDetail: false,
			expectedLite:   false,
			expectError:    false, // Should not error, just ignore invalid options
		},
		{
			name:           "Mixed valid and invalid",
			viewStr:        "detail,invalid,lite",
			expectedDetail: true,
			expectedLite:   true,
			expectError:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := &SpecImpl{}
			viewDeclare := getViewItems(test.viewStr)
			spec.ViewDeclare = viewDeclare

			// Check view declarations
			detailResult := spec.EnableView(model.DetailView)
			if detailResult != test.expectedDetail {
				t.Errorf("Detail view mismatch, name:%s, expected %v, got %v", test.name,
					test.expectedDetail, detailResult)
			}

			liteResult := spec.EnableView(model.LiteView)
			if liteResult != test.expectedLite {
				t.Errorf("Lite view mismatch, name:%s, expected %v, got %v", test.name,
					test.expectedLite, liteResult)
			}
		})
	}
}

func TestCombinedSpecAndViewParsing(t *testing.T) {
	// Test parsing both ORM spec and view tags
	tests := []struct {
		name           string
		ormSpec        string
		viewStr        string
		expectedName   string
		expectedPK     bool
		expectedValue  model.ValueDeclare
		expectedDetail bool
		expectedLite   bool
	}{
		{
			name:           "Primary key with views",
			ormSpec:        "id key",
			viewStr:        "detail,lite",
			expectedName:   "id",
			expectedPK:     true,
			expectedValue:  model.Customer,
			expectedDetail: true,
			expectedLite:   true,
		},
		{
			name:           "Auto-increment primary key with detail view",
			ormSpec:        "id auto key",
			viewStr:        "detail",
			expectedName:   "id",
			expectedPK:     true,
			expectedValue:  model.AutoIncrement,
			expectedDetail: true,
			expectedLite:   false,
		},
		{
			name:           "UUID field with lite view",
			ormSpec:        "id uuid",
			viewStr:        "lite",
			expectedName:   "id",
			expectedPK:     false,
			expectedValue:  model.UUID,
			expectedDetail: false,
			expectedLite:   true,
		},
		{
			name:           "Regular field with no views",
			ormSpec:        "description",
			viewStr:        "",
			expectedName:   "description",
			expectedPK:     false,
			expectedValue:  model.Customer,
			expectedDetail: false,
			expectedLite:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec, err := getOrmSpec(test.ormSpec)
			if err != nil {
				t.Errorf("Unexpected error for orm spec '%s': %v", test.ormSpec, err)
				return
			}

			spec.ViewDeclare = getViewItems(test.viewStr)

			// Check field name
			if spec.GetFieldName() != test.expectedName {
				t.Errorf("Field name mismatch, expected '%s', got '%s'",
					test.expectedName, spec.GetFieldName())
			}

			// Check primary key
			if spec.IsPrimaryKey() != test.expectedPK {
				t.Errorf("Primary key mismatch, expected %v, got %v",
					test.expectedPK, spec.IsPrimaryKey())
			}

			// Check value declare
			if spec.GetValueDeclare() != test.expectedValue {
				t.Errorf("Value declare mismatch, expected %v, got %v",
					test.expectedValue, spec.GetValueDeclare())
			}

			// Check view declarations
			detailResult := spec.EnableView(model.DetailView)
			if detailResult != test.expectedDetail {
				t.Errorf("Detail view mismatch, expected %v, got %v",
					test.expectedDetail, detailResult)
			}

			liteResult := spec.EnableView(model.LiteView)
			if liteResult != test.expectedLite {
				t.Errorf("Lite view mismatch, expected %v, got %v",
					test.expectedLite, liteResult)
			}
		})
	}
}
