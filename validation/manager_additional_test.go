package validation

import (
	"errors"
	"reflect"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	verrors "github.com/muidea/magicOrm/validation/errors"
)

type testDirective struct {
	key  models.Key
	args []string
}

func (d testDirective) Key() models.Key { return d.key }
func (d testDirective) Args() []string  { return d.args }
func (d testDirective) HasArgs() bool   { return len(d.args) > 0 }

type testConstraints struct {
	directives []models.Directive
}

func (c testConstraints) Has(key models.Key) bool {
	_, ok := c.Get(key)
	return ok
}

func (c testConstraints) Get(key models.Key) (models.Directive, bool) {
	for _, directive := range c.directives {
		if directive.Key() == key {
			return directive, true
		}
	}
	return nil, false
}

func (c testConstraints) Directives() []models.Directive { return c.directives }

type testType struct {
	name  string
	value models.TypeDeclare
	ptr   bool
	elem  models.Type
}

func (t *testType) GetName() string        { return t.name }
func (t *testType) GetPkgPath() string     { return "validation.test" }
func (t *testType) GetPkgKey() string      { return t.GetPkgPath() + "/" + t.name }
func (t *testType) GetDescription() string { return t.name }
func (t *testType) GetValue() models.TypeDeclare {
	return t.value
}
func (t *testType) IsPtrType() bool { return t.ptr }
func (t *testType) Interface(initVal any) (models.Value, *cd.Error) {
	return &testValue{value: initVal, valid: true}, nil
}
func (t *testType) Elem() models.Type {
	if t.elem != nil {
		return t.elem
	}
	return t
}

type testValue struct {
	value    any
	valid    bool
	assigned bool
}

func (v *testValue) IsValid() bool    { return v.valid }
func (v *testValue) IsZero() bool     { return v.value == nil || reflect.ValueOf(v.value).IsZero() }
func (v *testValue) Get() any         { return v.value }
func (v *testValue) IsAssigned() bool { return v.assigned }
func (v *testValue) Set(val any) *cd.Error {
	v.value = val
	v.valid = true
	v.assigned = true
	return nil
}
func (v *testValue) UnpackValue() []models.Value {
	return nil
}

type testSpec struct {
	primary     bool
	constraints models.Constraints
}

func (s *testSpec) IsPrimaryKey() bool                   { return s.primary }
func (s *testSpec) GetValueDeclare() models.ValueDeclare { return models.Customer }
func (s *testSpec) GetConstraints() models.Constraints   { return s.constraints }
func (s *testSpec) EnableView(models.ViewDeclare) bool   { return true }
func (s *testSpec) GetDefaultValue() any                 { return nil }

type testField struct {
	name  string
	typ   models.Type
	spec  models.Spec
	value models.Value
}

func (f *testField) GetName() string        { return f.name }
func (f *testField) GetShowName() string    { return f.name }
func (f *testField) GetDescription() string { return f.name }
func (f *testField) GetType() models.Type   { return f.typ }
func (f *testField) GetSpec() models.Spec   { return f.spec }
func (f *testField) GetValue() models.Value { return f.value }
func (f *testField) SetValue(val any) *cd.Error {
	f.value = &testValue{value: val, valid: true, assigned: true}
	return nil
}
func (f *testField) GetSliceValue() []models.Value { return nil }
func (f *testField) AppendSliceValue(any) *cd.Error {
	return nil
}
func (f *testField) Reset() {}

type testModel struct {
	name   string
	fields models.Fields
}

func (m *testModel) GetName() string        { return m.name }
func (m *testModel) GetShowName() string    { return m.name }
func (m *testModel) GetPkgPath() string     { return "validation.test" }
func (m *testModel) GetPkgKey() string      { return m.GetPkgPath() + "/" + m.name }
func (m *testModel) GetDescription() string { return m.name }
func (m *testModel) GetFields() models.Fields {
	return m.fields
}
func (m *testModel) SetFieldValue(name string, val any) *cd.Error {
	for _, field := range m.fields {
		if field.GetName() == name {
			return field.SetValue(val)
		}
	}
	return cd.NewError(cd.IllegalParam, "field not found")
}
func (m *testModel) SetPrimaryFieldValue(val any) *cd.Error {
	primary := m.GetPrimaryField()
	if primary == nil {
		return cd.NewError(cd.IllegalParam, "primary field not found")
	}
	return primary.SetValue(val)
}
func (m *testModel) GetPrimaryField() models.Field { return m.fields.GetPrimaryField() }
func (m *testModel) GetField(name string) models.Field {
	for _, field := range m.fields {
		if field.GetName() == name {
			return field
		}
	}
	return nil
}
func (m *testModel) Interface(bool) any                   { return nil }
func (m *testModel) Copy(models.ViewDeclare) models.Model { return m }
func (m *testModel) Reset()                               {}

type registeredTypeHandler struct{}

func (registeredTypeHandler) Validate(any) error             { return nil }
func (registeredTypeHandler) Convert(value any) (any, error) { return value, nil }
func (registeredTypeHandler) GetZeroValue() any              { return "" }
func (registeredTypeHandler) GetType() reflect.Type          { return reflect.TypeOf("") }

func TestReflectTypeFromModelType(t *testing.T) {
	intType := &testType{name: "Count", value: models.TypeIntegerValue}
	ptrStringType := &testType{name: "Alias", value: models.TypeStringValue, ptr: true}
	sliceType := &testType{
		name:  "Tags",
		value: models.TypeSliceValue,
		elem:  &testType{name: "Tag", value: models.TypeStringValue},
	}

	if got := ReflectTypeFromModelType(intType, nil); got != reflect.TypeOf(int(0)) {
		t.Fatalf("unexpected int reflect type: %v", got)
	}
	if got := ReflectTypeFromModelType(ptrStringType, nil); got != reflect.TypeOf((*string)(nil)) {
		t.Fatalf("unexpected pointer reflect type: %v", got)
	}
	if got := ReflectTypeFromModelType(sliceType, nil); got != reflect.TypeOf([]string{}) {
		t.Fatalf("unexpected slice reflect type: %v", got)
	}
	if got := ReflectTypeFromModelType(&testType{name: "Unknown", value: models.TypeStructValue}, nil); got != interfaceType {
		t.Fatalf("unexpected fallback reflect type: %v", got)
	}
	if got := ReflectTypeFromModelType(&testType{name: "When", value: models.TypeDateTimeValue}, nil); got != reflect.TypeOf(time.Time{}) {
		t.Fatalf("unexpected datetime reflect type: %v", got)
	}
	if got := ReflectTypeFromModelType(&testType{name: "When", value: models.TypeDateTimeValue}, "2025-01-01 00:00:00"); got != reflect.TypeOf("") {
		t.Fatalf("unexpected datetime fallback reflect type: %v", got)
	}
	if got := ReflectTypeFromModelType(&testType{name: "Flag", value: models.TypeBooleanValue}, nil); got != reflect.TypeOf(false) {
		t.Fatalf("unexpected bool reflect type: %v", got)
	}
	if got := ReflectTypeFromModelType(nil, 1); got != reflect.TypeOf(1) {
		t.Fatalf("unexpected fallback reflect type from value: %v", got)
	}
	if got := ReflectTypeFromModelType(
		&testType{
			name:  "Children",
			value: models.TypeSliceValue,
			elem:  &testType{name: "Child", value: models.TypeStructValue, ptr: true},
		},
		[]*testModel{{}},
	); got != reflect.TypeOf([]*testModel{}) {
		t.Fatalf("unexpected struct slice fallback reflect type: %v", got)
	}
}

func TestValidateFieldUsesRealFieldMetadata(t *testing.T) {
	manager := NewValidationManager(DefaultConfig())

	field := &testField{
		name: "name",
		typ:  &testType{name: "Name", value: models.TypeStringValue},
		spec: &testSpec{constraints: testConstraints{directives: []models.Directive{
			testDirective{key: models.KeyRequired},
			testDirective{key: models.KeyMin, args: []string{"3"}},
		}}},
		value: &testValue{value: "", valid: true},
	}

	ctx := NewContext(verrors.ScenarioInsert, OperationCreate, nil, "")
	if err := manager.ValidateField(field, "ab", ctx); err == nil {
		t.Fatal("expected field validation error for short value")
	}

	if err := manager.ValidateField(field, "abcd", ctx); err != nil {
		t.Fatalf("expected valid field value, got %v", err)
	}
}

func TestValidateModelUsesActualModelFields(t *testing.T) {
	manager := NewValidationManager(DefaultConfig())

	model := &testModel{
		name: "User",
		fields: models.Fields{
			&testField{
				name:  "id",
				typ:   &testType{name: "ID", value: models.TypeIntegerValue},
				spec:  &testSpec{primary: true},
				value: &testValue{value: 1, valid: true},
			},
			&testField{
				name: "name",
				typ:  &testType{name: "Name", value: models.TypeStringValue},
				spec: &testSpec{constraints: testConstraints{directives: []models.Directive{
					testDirective{key: models.KeyRequired},
					testDirective{key: models.KeyMin, args: []string{"3"}},
				}}},
				value: &testValue{value: "ab", valid: true},
			},
		},
	}

	ctx := NewContext(verrors.ScenarioInsert, OperationCreate, nil, "postgresql")
	if err := manager.ValidateModel(model, ctx); err == nil {
		t.Fatal("expected model validation error for invalid field")
	}

	if setErr := model.SetFieldValue("name", "valid-name"); setErr != nil {
		t.Fatalf("failed to update field value: %v", setErr)
	}
	if err := manager.ValidateModel(model, ctx); err != nil {
		t.Fatalf("expected valid model, got %v", err)
	}
}

func TestValidateModelUpdateSkipsUnassignedFields(t *testing.T) {
	manager := NewValidationManager(DefaultConfig())

	requiredConstraints := testConstraints{directives: []models.Directive{
		testDirective{key: models.KeyRequired},
		testDirective{key: models.KeyMin, args: []string{"3"}},
	}}

	model := &testModel{
		name: "User",
		fields: models.Fields{
			&testField{
				name:  "id",
				typ:   &testType{name: "ID", value: models.TypeIntegerValue},
				spec:  &testSpec{primary: true},
				value: &testValue{value: 1, valid: true, assigned: true},
			},
			&testField{
				name:  "name",
				typ:   &testType{name: "Name", value: models.TypeStringValue},
				spec:  &testSpec{constraints: requiredConstraints},
				value: &testValue{value: nil, valid: false, assigned: false},
			},
		},
	}

	updateCtx := NewContext(verrors.ScenarioUpdate, OperationUpdate, nil, "")
	if err := manager.ValidateModel(model, updateCtx); err != nil {
		t.Fatalf("expected update validation to skip unassigned required field, got %v", err)
	}

	if setErr := model.SetFieldValue("name", nil); setErr != nil {
		t.Fatalf("failed to update field value: %v", setErr)
	}
	if err := manager.ValidateModel(model, updateCtx); err == nil {
		t.Fatal("expected update validation to reject assigned nil required value")
	}
}

func TestValidateModelHonorsProvidedAdapter(t *testing.T) {
	manager := NewValidationManager(DefaultConfig())

	ctx := NewContext(verrors.ScenarioInsert, OperationCreate, NewModelAdapter([]FieldAdapter{
		NewFieldAdapter(
			"override",
			reflect.TypeOf(""),
			testConstraints{directives: []models.Directive{
				testDirective{key: models.KeyRequired},
				testDirective{key: models.KeyMin, args: []string{"5"}},
			}},
			"bad",
		),
	}), "")

	if err := manager.ValidateModel(&testModel{name: "Ignored"}, ctx); err == nil {
		t.Fatal("expected validation error from provided adapter")
	}
}

func TestValidateModelAdapterUnwrapsWrappedFieldValue(t *testing.T) {
	manager := NewValidationManager(DefaultConfig())

	ctx := NewContext(verrors.ScenarioInsert, OperationCreate, NewModelAdapter([]FieldAdapter{
		NewFieldAdapter(
			"warehouse",
			reflect.TypeOf((*struct{ ID int64 })(nil)),
			testConstraints{directives: []models.Directive{
				testDirective{key: models.KeyRequired},
			}},
			&testValue{value: nil, valid: false},
		),
	}), "")

	if err := manager.ValidateModel(&testModel{name: "Ignored"}, ctx); err == nil {
		t.Fatal("expected validation error from wrapped nil field value")
	}
}

func TestAdapterHelpersAndScenarios(t *testing.T) {
	field := &testField{
		name: "secret",
		typ:  &testType{name: "Secret", value: models.TypeStringValue},
		spec: &testSpec{constraints: testConstraints{directives: []models.Directive{
			testDirective{key: models.KeyRequired},
			testDirective{key: models.KeyReadOnly},
		}}},
		value: &testValue{value: "value", valid: true},
	}

	adapter := AdaptModel(&testModel{name: "User", fields: models.Fields{field}})
	if _, err := adapter.GetField("missing"); err == nil {
		t.Fatal("expected missing field lookup to fail")
	}

	if GetFieldSpec(field) == nil || !HasFieldConstraint(field, models.KeyRequired) {
		t.Fatal("expected helper functions to expose field constraints")
	}
	if len(GetFieldConstraints(field).Directives()) != 2 {
		t.Fatal("expected field constraints to be returned")
	}
	if GetFieldTypeName(field) != "Secret" {
		t.Fatalf("unexpected field type name: %s", GetFieldTypeName(field))
	}
	if !AdaptField(field, field.GetValue().Get()).HasConstraint(models.KeyReadOnly) {
		t.Fatal("expected adapted field to expose read-only constraint")
	}
	if !IsFieldRequired(field) || !IsFieldReadOnly(field) || IsFieldWriteOnly(field) {
		t.Fatal("unexpected field helper results")
	}
	if GetFieldConstraints(&testField{name: "plain", typ: &testType{name: "plain", value: models.TypeStringValue}}) != nil {
		t.Fatal("expected field without spec to return nil constraints")
	}
	if GetFieldTypeName(&testField{name: "empty"}) != "" {
		t.Fatal("expected field without type to return empty type name")
	}

	scenarioAdapter := NewScenarioAdapter()
	if !scenarioAdapter.ShouldValidateConstraint(models.KeyRequired, verrors.ScenarioInsert) {
		t.Fatal("expected insert scenario to validate required constraint")
	}
	if scenarioAdapter.ShouldValidateConstraint(models.KeyWriteOnly, verrors.ScenarioQuery) {
		t.Fatal("expected query scenario to skip write-only constraint")
	}
	if !scenarioAdapter.GetValidationStrategy(verrors.ScenarioDelete).ShouldSkipReadOnlyFields() {
		t.Fatal("expected delete scenario to skip read-only fields")
	}
	if !scenarioAdapter.GetValidationStrategy(verrors.ScenarioInsert).IsStrictMode() {
		t.Fatal("expected insert scenario strict mode")
	}
	if !scenarioAdapter.GetValidationStrategy(verrors.ScenarioUpdate).ShouldValidate(models.KeyReadOnly) {
		t.Fatal("expected update strategy to validate read-only constraint")
	}
	if scenarioAdapter.GetValidationStrategy(verrors.ScenarioQuery).ShouldValidate(models.KeyRequired) {
		t.Fatal("expected query strategy to skip required constraint")
	}
	if scenarioAdapter.GetValidationStrategy(verrors.ScenarioDelete).ShouldValidate(models.KeyMax) {
		t.Fatal("expected delete strategy to skip non-required constraints")
	}
	if scenarioAdapter.GetValidationStrategy(verrors.ScenarioUpdate).IsStrictMode() {
		t.Fatal("expected update strategy to be non-strict")
	}
}

func TestManagerRegistrationAndStats(t *testing.T) {
	manager := NewValidationManager(DefaultConfig())
	factory := NewValidationFactory()

	if err := manager.RegisterCustomConstraint("custom", func(val any, _ []string) error {
		if val == "ok" {
			return nil
		}
		return errors.New("bad")
	}); err != nil {
		t.Fatalf("register custom constraint failed: %v", err)
	}

	if err := manager.RegisterTypeHandler("custom", registeredTypeHandler{}); err != nil {
		t.Fatalf("register type handler failed: %v", err)
	}
	if err := factory.RegisterCustomConstraint("custom", func(val any, _ []string) error { return nil }); err != nil {
		t.Fatalf("factory register custom constraint failed: %v", err)
	}
	if err := factory.RegisterTypeHandler("custom", registeredTypeHandler{}); err != nil {
		t.Fatalf("factory register type handler failed: %v", err)
	}
	if _, ok := factory.CreateTypeValidator().(TypeValidator); !ok {
		t.Fatal("expected factory to create type validator")
	}

	field := NewFieldAdapter("name", reflect.TypeOf(""), testConstraints{directives: []models.Directive{
		testDirective{key: "custom"},
	}}, "bad")

	ctx := NewContext(verrors.ScenarioInsert, OperationCreate, NewModelAdapter([]FieldAdapter{field}), "")
	ctx.Field = field
	if err := manager.Validate("bad", ctx); err == nil {
		t.Fatal("expected custom constraint validation failure")
	}

	stats := manager.GetValidationStats()
	if stats.TotalValidations == 0 || stats.FailedValidations == 0 {
		t.Fatalf("expected validation stats to be updated, got %+v", stats)
	}

	manager.ResetStats()
	stats = manager.GetValidationStats()
	if stats.TotalValidations != 0 || stats.FailedValidations != 0 {
		t.Fatalf("expected stats reset, got %+v", stats)
	}

	manager.SetScenario(verrors.ScenarioUpdate)
}

func TestInternalValidationSteps(t *testing.T) {
	manager := NewValidationManager(DefaultConfig()).(*validationManagerImpl)
	collector := verrors.NewErrorCollector()
	field := NewFieldAdapter(
		"age",
		reflect.TypeOf(int(0)),
		testConstraints{directives: []models.Directive{
			testDirective{key: models.KeyMin, args: []string{"2"}},
			testDirective{key: models.KeyRequired},
		}},
		1,
	)

	ctx := ValidationContext{
		Scenario:     verrors.ScenarioInsert,
		Field:        field,
		DatabaseType: "postgresql",
		Options: ValidationOptions{
			ValidateReadOnlyFields:  true,
			ValidateWriteOnlyFields: true,
		},
		Collector: collector,
	}

	if err := manager.validateType("bad", ctx); err == nil {
		t.Fatal("expected type validation failure")
	}
	collector.Clear()
	if err := manager.validateConstraints(1, ctx); err == nil {
		t.Fatal("expected constraint validation failure")
	}
	collector.Clear()
	if err := manager.validateDatabase(nil, ctx); err == nil {
		t.Fatal("expected database validation failure")
	}

	ctx.Field = NewFieldAdapter(
		"secret",
		reflect.TypeOf(""),
		testConstraints{directives: []models.Directive{
			testDirective{key: models.KeyWriteOnly},
		}},
		"token",
	)
	ctx.Scenario = verrors.ScenarioQuery
	ctx.Options.ValidateWriteOnlyFields = false
	if err := manager.validateConstraints("token", ctx); err != nil {
		t.Fatalf("expected write-only field to be skipped during query validation, got %v", err)
	}
}
