package remote

import (
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

type fakeModel struct{}

func (fakeModel) GetName() string                      { return "fake" }
func (fakeModel) GetShowName() string                  { return "fake" }
func (fakeModel) GetPkgPath() string                   { return "/fake" }
func (fakeModel) GetPkgKey() string                    { return "/fake/fake" }
func (fakeModel) GetDescription() string               { return "fake" }
func (fakeModel) GetFields() models.Fields             { return nil }
func (fakeModel) SetFieldValue(string, any) *cd.Error  { return nil }
func (fakeModel) SetPrimaryFieldValue(any) *cd.Error   { return nil }
func (fakeModel) GetPrimaryField() models.Field        { return nil }
func (fakeModel) GetField(string) models.Field         { return nil }
func (fakeModel) Interface(bool) any                   { return nil }
func (fakeModel) Copy(models.ViewDeclare) models.Model { return fakeModel{} }
func (fakeModel) Reset()                               {}

func TestRemoteEntityHelpersRoundTripVMIProduct(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/product/product.json")

	entityType, err := GetEntityType(object)
	if err != nil {
		t.Fatalf("GetEntityType failed: %v", err)
	}
	if entityType.GetName() != object.GetName() || entityType.GetPkgPath() != object.GetPkgPath() {
		t.Fatalf("GetEntityType mismatch, got %s/%s", entityType.GetPkgPath(), entityType.GetName())
	}
	if entityType.GetValue() != models.TypeStructValue || !entityType.IsPtrType() {
		t.Fatalf("GetEntityType should return pointer struct, got value=%v isPtr=%v", entityType.GetValue(), entityType.IsPtrType())
	}

	model, err := GetEntityModel(object, nil)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}

	objectValue := &ObjectValue{Name: object.GetName(), PkgPath: object.GetPkgPath()}
	objectValue.SetFieldValue("id", int64(1001))
	objectValue.SetFieldValue("name", "apple")

	modelValue, err := GetEntityValue(objectValue)
	if err != nil {
		t.Fatalf("GetEntityValue failed: %v", err)
	}

	assignedModel, err := SetModelValue(model, modelValue, true)
	if err != nil {
		t.Fatalf("SetModelValue failed: %v", err)
	}

	assignedObject, ok := assignedModel.(*Object)
	if !ok {
		t.Fatalf("SetModelValue should return *Object, got %T", assignedModel)
	}

	result, ok := assignedObject.Interface(false).(*ObjectValue)
	if !ok {
		t.Fatalf("Interface should return *ObjectValue, got %T", assignedObject.Interface(false))
	}
	if result.ID != "1001" {
		t.Fatalf("Interface ID mismatch, got %v", result.ID)
	}
	if got := result.GetFieldValue("name"); got != "apple" {
		t.Fatalf("Interface name mismatch, got %v", got)
	}

	filter, err := GetModelFilter(assignedObject)
	if err != nil {
		t.Fatalf("GetModelFilter failed: %v", err)
	}
	if err := filter.Equal("name", "apple"); err != nil {
		t.Fatalf("filter.Equal failed: %v", err)
	}

	filterItem := filter.GetFilterItem("name")
	if filterItem == nil {
		t.Fatalf("filter item for name should exist")
	}
	if filterItem.OprCode() != models.EqualOpr {
		t.Fatalf("filter item should use EqualOpr, got %v", filterItem.OprCode())
	}
	if filterItem.OprValue().Get() != "apple" {
		t.Fatalf("filter item value mismatch, got %v", filterItem.OprValue().Get())
	}
}

func TestRemoteEntityHelpersPrimaryValueAndErrors(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/product/product.json")

	model, err := GetEntityModel(object, nil)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}

	assignedModel, err := SetModelValue(model, &ValueImpl{value: int64(2002)}, true)
	if err != nil {
		t.Fatalf("SetModelValue primary failed: %v", err)
	}
	assignedObject := assignedModel.(*Object)
	if primary := assignedObject.GetPrimaryField().GetValue().Get(); primary != int64(2002) {
		t.Fatalf("primary field should be assigned, got %v", primary)
	}

	if _, err := GetEntityType(nil); err == nil {
		t.Fatalf("GetEntityType(nil) should fail")
	}
	if _, err := GetEntityValue(nil); err == nil {
		t.Fatalf("GetEntityValue(nil) should fail")
	}
	if _, err := GetEntityModel(nil, nil); err == nil {
		t.Fatalf("GetEntityModel(nil) should fail")
	}
	if _, err := GetModelFilter(nil); err == nil {
		t.Fatalf("GetModelFilter(nil) should fail")
	}
	if _, err := SetModelValue(model, &ValueImpl{}, true); err == nil {
		t.Fatalf("SetModelValue with invalid value should fail")
	}
}

func TestRemoteSetModelValueSkipsUnassignedZeroFields(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/product/product.json")

	model, err := GetEntityModel(object, nil)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}

	queryValue := &ObjectValue{
		Name:    object.GetName(),
		PkgPath: object.GetPkgPath(),
		Fields: []*FieldValue{
			{Name: "id", Value: int64(0)},
			{Name: "expire", Value: 0},
			{Name: "name", Value: "apple"},
		},
	}

	assignedModel, err := SetModelValue(model, &ValueImpl{value: queryValue}, true)
	if err != nil {
		t.Fatalf("SetModelValue failed: %v", err)
	}

	assignedObject := assignedModel.(*Object)
	if models.IsAssignedField(assignedObject.GetField("id")) {
		t.Fatal("unassigned zero primary field should be skipped")
	}
	if models.IsAssignedField(assignedObject.GetField("expire")) {
		t.Fatal("unassigned zero basic field should be skipped")
	}
	if got := assignedObject.GetField("name").GetValue().Get(); got != "apple" {
		t.Fatalf("non-zero field should still be assigned, got %#v", got)
	}

	queryValue.Fields[0].Assigned = true
	queryValue.Fields[1].Assigned = true

	assignedModel, err = SetModelValue(model.Copy(models.MetaView), &ValueImpl{value: queryValue}, true)
	if err != nil {
		t.Fatalf("SetModelValue(explicit zero) failed: %v", err)
	}

	assignedObject = assignedModel.(*Object)
	if !models.IsAssignedField(assignedObject.GetField("id")) || assignedObject.GetField("id").GetValue().Get() != int64(0) {
		t.Fatalf("explicit zero primary should be assigned, got %#v", assignedObject.GetField("id").GetValue().Get())
	}
	if !models.IsAssignedField(assignedObject.GetField("expire")) || assignedObject.GetField("expire").GetValue().Get() != int(0) {
		t.Fatalf("explicit zero basic field should be assigned, got %#v", assignedObject.GetField("expire").GetValue().Get())
	}
}

func TestRemoteEntityHelpersVariants(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/product/product.json")
	objectValue := ObjectValue{
		ID:      "1001",
		Name:    object.GetName(),
		PkgPath: object.GetPkgPath(),
		Fields: []*FieldValue{
			{Name: "id", Value: int64(1001)},
		},
	}
	sliceValue := SliceObjectValue{
		Name:    object.GetName(),
		PkgPath: object.GetPkgPath(),
		Values: []*ObjectValue{
			{
				ID:      "1001",
				Name:    object.GetName(),
				PkgPath: object.GetPkgPath(),
				Fields: []*FieldValue{
					{Name: "id", Value: int64(1001)},
				},
			},
		},
	}

	cases := []struct {
		name      string
		entity    any
		wantValue models.TypeDeclare
	}{
		{name: "object pointer", entity: object, wantValue: models.TypeStructValue},
		{name: "object value", entity: *object, wantValue: models.TypeStructValue},
		{name: "object value pointer", entity: &objectValue, wantValue: models.TypeStructValue},
		{name: "object value value", entity: objectValue, wantValue: models.TypeStructValue},
		{name: "slice pointer", entity: &sliceValue, wantValue: models.TypeSliceValue},
		{name: "slice value", entity: sliceValue, wantValue: models.TypeSliceValue},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entityType, err := GetEntityType(tc.entity)
			if err != nil {
				t.Fatalf("GetEntityType failed: %v", err)
			}
			if entityType.GetValue() != tc.wantValue || !entityType.IsPtrType() {
				t.Fatalf("GetEntityType mismatch, got value=%v isPtr=%v", entityType.GetValue(), entityType.IsPtrType())
			}
		})
	}

	valueCases := []struct {
		name   string
		entity any
	}{
		{name: "object value pointer", entity: &objectValue},
		{name: "object value value", entity: objectValue},
		{name: "slice pointer", entity: &sliceValue},
		{name: "slice value", entity: sliceValue},
	}
	for _, tc := range valueCases {
		t.Run(tc.name, func(t *testing.T) {
			entityValue, err := GetEntityValue(tc.entity)
			if err != nil {
				t.Fatalf("GetEntityValue failed: %v", err)
			}
			if !entityValue.IsValid() {
				t.Fatalf("GetEntityValue should be valid")
			}
		})
	}

	valueModel, err := GetEntityModel(*object, nil)
	if err != nil {
		t.Fatalf("GetEntityModel(object value) failed: %v", err)
	}
	if valueModel.GetPkgKey() != object.GetPkgKey() {
		t.Fatalf("GetEntityModel(object value) mismatch, got %s", valueModel.GetPkgKey())
	}
}

func TestRemoteEntityHelpersInvalidTypeAndRecovery(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/product/product.json")

	if _, err := GetEntityType(123); err == nil {
		t.Fatalf("GetEntityType(non-entity) should fail")
	}
	if _, err := GetEntityValue(123); err == nil {
		t.Fatalf("GetEntityValue(non-entity) should fail")
	}
	if _, err := GetEntityModel(123, nil); err == nil {
		t.Fatalf("GetEntityModel(non-entity) should fail")
	}
	if _, err := GetModelFilter(fakeModel{}); err == nil {
		t.Fatalf("GetModelFilter(fakeModel) should fail")
	}
	if _, err := SetModelValue(nil, &ValueImpl{value: int64(1)}, true); err == nil {
		t.Fatalf("SetModelValue(nil model) should fail")
	}
	if _, err := SetModelValue(object, nil, true); err == nil {
		t.Fatalf("SetModelValue(nil value) should fail")
	}
	if _, err := SetModelValue(fakeModel{}, &ValueImpl{value: int64(1)}, true); err == nil {
		t.Fatalf("SetModelValue(fakeModel) should recover panic and fail")
	}
}

func TestRemoteSetModelValueAssignObjectValueErrors(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/product/product.json")

	unknownFieldValue := &ObjectValue{
		Name:    object.GetName(),
		PkgPath: object.GetPkgPath(),
		Fields: []*FieldValue{
			{Name: "unknown", Value: "x"},
		},
	}
	assignedModel, err := SetModelValue(object, &ValueImpl{value: unknownFieldValue}, true)
	if err != nil {
		t.Fatalf("SetModelValue with unknown field should keep current ignore behavior, got %v", err)
	}
	if got := assignedModel.(*Object).GetField("unknown"); got != nil {
		t.Fatalf("unknown field should not be materialized, got %#v", got)
	}

	illegalStatusValue := &ObjectValue{
		Name:    object.GetName(),
		PkgPath: object.GetPkgPath(),
		Fields: []*FieldValue{
			{Name: "status", Value: "illegal"},
		},
	}
	if _, err := SetModelValue(object, &ValueImpl{value: illegalStatusValue}, true); err == nil {
		t.Fatalf("SetModelValue with illegal relation value should fail")
	}
}
