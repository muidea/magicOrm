// design_verification_test.go 对照 DESIGN-CONSISTENCY.md 编写，用于验证实现是否符合设计文档。
// 验证结果与异常记录见 DESIGN-CONSISTENCY-VERIFICATION.md。

package consistency

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

// TestModelInterfaceLocal 对应设计 2.1/2.2：Local 通过 Model/Field/Value 暴露数据，禁止直接操作 reflect。
func TestModelInterfaceLocal(t *testing.T) {
	entity := NewBasicTypes()
	m, err := local.GetEntityModel(entity, nil)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}
	if m.GetName() == "" {
		t.Error("Model.GetName() should not be empty")
	}
	if m.GetFields() == nil || len(m.GetFields()) == 0 {
		t.Error("Model.GetFields() should return non-empty")
	}
	f := m.GetField("id")
	if f == nil {
		t.Error("GetField(\"id\") should not be nil")
	}
	if f != nil && f.GetValue() != nil && !f.GetValue().IsValid() {
		t.Error("primary field value should be valid")
	}
	// SetFieldValue 通过 Model 接口写入
	if err := m.SetFieldValue("str", "updated"); err != nil {
		t.Errorf("SetFieldValue failed: %v", err)
	}
	if entity.Str != "updated" {
		t.Errorf("entity.Str expected 'updated', got %q", entity.Str)
	}
}

// TestModelInterfaceRemote 对应设计 2.1/2.2：Remote Object 实现 models.Model，通过 Model/Field/Value 暴露数据。
func TestModelInterfaceRemote(t *testing.T) {
	entity := NewBasicTypes()
	obj, err := helper.GetObject(entity)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	var m models.Model = obj
	if m.GetName() == "" {
		t.Error("Remote Model.GetName() should not be empty")
	}
	if m.GetFields() == nil || len(m.GetFields()) == 0 {
		t.Error("Remote Model.GetFields() should return non-empty")
	}
	f := m.GetField("id")
	if f == nil {
		t.Error("Remote GetField(\"id\") should not be nil")
	}
	// 通过 Model 接口赋值（设计 7.4.5）
	if err := m.SetFieldValue("str", "remote_updated"); err != nil {
		t.Errorf("Remote SetFieldValue failed: %v", err)
	}
	// 读回验证
	f2 := m.GetField("str")
	if f2 == nil || f2.GetValue() == nil {
		t.Error("GetField(\"str\") after SetFieldValue should not be nil")
	}
	if f2 != nil && f2.GetValue() != nil {
		if v, ok := f2.GetValue().Get().(string); !ok || v != "remote_updated" {
			t.Errorf("expected str=remote_updated, got %v", f2.GetValue().Get())
		}
	}
}

// TestDesignRoundTripLocalRemoteJSON 对应设计 8.4.1：Local → Remote → marshal/unmarshal → Remote → Local 完全一致。
func TestDesignRoundTripLocalRemoteJSON(t *testing.T) {
	original := NewBasicTypes()
	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}
	jsonData, err := remote.EncodeObjectValue(objValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}
	decodedValue, err := remote.DecodeObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}
	target := &BasicTypes{}
	if err = helper.UpdateEntity(decodedValue, target); err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}
	if !compareBasicTypes(original, target) {
		t.Error("design 8.4.1: Local→Remote→JSON→Remote→Local 与起点不一致")
	}
}

// TestDesignRoundTripLocalRemoteJSONWithNested 对应设计 8.4.1：Local→Remote→marshal/unmarshal→Remote→Local 完全一致，且包含嵌套定义（单层/多层嵌套对象、对象切片 []T、指针切片 []*T）。
func TestDesignRoundTripLocalRemoteJSONWithNested(t *testing.T) {
	tests := []struct {
		name     string
		original any
		target   func() any
		compare  func(a, b any) bool
	}{
		{
			name:     "NestedParent",
			original: NewNestedParent(),
			target:   func() any { return &NestedParent{} },
			compare:  func(a, b any) bool { return compareNestedParent(a.(*NestedParent), b.(*NestedParent)) },
		},
		{
			name:     "NestedSliceParent",
			original: NewNestedSliceParent(),
			target:   func() any { return &NestedSliceParent{} },
			compare:  func(a, b any) bool { return compareNestedSliceParent(a.(*NestedSliceParent), b.(*NestedSliceParent)) },
		},
		{
			name:     "NestedSlicePtrParent",
			original: NewNestedSlicePtrParent(),
			target:   func() any { return &NestedSlicePtrParent{} },
			compare:  func(a, b any) bool { return compareNestedSlicePtrParent(a.(*NestedSlicePtrParent), b.(*NestedSlicePtrParent)) },
		},
		{
			name:     "DeepLevel3",
			original: NewDeepLevel3(),
			target:   func() any { return &DeepLevel3{} },
			compare:  func(a, b any) bool { return compareDeepLevel3(a.(*DeepLevel3), b.(*DeepLevel3)) },
		},
		{
			name:     "ComplexEntity",
			original: NewComplexEntity(),
			target:   func() any { return &ComplexEntity{} },
			compare:  func(a, b any) bool { return compareComplexEntity(a.(*ComplexEntity), b.(*ComplexEntity)) },
		},
		{
			name:     "AllInOne",
			original: NewAllInOne(),
			target:   func() any { return &AllInOne{} },
			compare:  func(a, b any) bool { return compareAllInOne(a.(*AllInOne), b.(*AllInOne)) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objValue, err := helper.GetObjectValue(tt.original)
			if err != nil {
				t.Fatalf("GetObjectValue failed: %v", err)
			}
			jsonData, err := remote.EncodeObjectValue(objValue)
			if err != nil {
				t.Fatalf("EncodeObjectValue failed: %v", err)
			}
			decodedValue, err := remote.DecodeObjectValue(jsonData)
			if err != nil {
				t.Fatalf("DecodeObjectValue failed: %v", err)
			}
			target := tt.target()
			if err = helper.UpdateEntity(decodedValue, target); err != nil {
				t.Fatalf("UpdateEntity failed: %v", err)
			}
			if !tt.compare(tt.original, target) {
				t.Errorf("design 8.4.1 (nested): %s Local→Remote→marshal/unmarshal→Remote→Local 与起点不一致", tt.name)
			}
		})
	}

	// 实体切片（元素含嵌套）：Local slice → GetSliceObjectValue → marshal/unmarshal → UpdateSliceEntity → Local
	t.Run("SliceOfNestedParent", func(t *testing.T) {
		original := []*NestedParent{
			NewNestedParent(),
			{ID: 2, Name: "p2", Child: &NestedChild{ID: 20, Name: "c2"}},
		}
		sliceVal, err := helper.GetSliceObjectValue(original)
		if err != nil {
			t.Fatalf("GetSliceObjectValue failed: %v", err)
		}
		jsonData, err := remote.EncodeSliceObjectValue(sliceVal)
		if err != nil {
			t.Fatalf("EncodeSliceObjectValue failed: %v", err)
		}
		decodedSlice, err := remote.DecodeSliceObjectValue(jsonData)
		if err != nil {
			t.Fatalf("DecodeSliceObjectValue failed: %v", err)
		}
		var target []*NestedParent
		if err = helper.UpdateSliceEntity(decodedSlice, &target); err != nil {
			t.Fatalf("UpdateSliceEntity failed: %v", err)
		}
		if len(target) != len(original) {
			t.Fatalf("slice length: expected %d, got %d", len(original), len(target))
		}
		for i := range original {
			if !compareNestedParent(original[i], target[i]) {
				t.Errorf("design 8.4.1 (nested slice)[%d]: Local→Remote→marshal/unmarshal→Remote→Local 与起点不一致", i)
			}
		}
	})
}

// TestDesignRoundTripRemoteObjectValue 对应设计 8.4.2：ObjectValue → marshal → unmarshal → ObjectValue 完全一致。
func TestDesignRoundTripRemoteObjectValue(t *testing.T) {
	original := NewBasicTypes()
	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}
	jsonData, err := remote.EncodeObjectValue(objValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}
	decoded, err := remote.DecodeObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}
	if !remote.CompareObjectValue(objValue, decoded) {
		t.Error("design 8.4.2: ObjectValue marshal/unmarshal 后与起点不一致")
	}
}

// TestDesignRoundTripRemoteSliceObjectValue 对应设计 8.4.2：SliceObjectValue → marshal → unmarshal → SliceObjectValue 完全一致。
func TestDesignRoundTripRemoteSliceObjectValue(t *testing.T) {
	original := []*BasicTypes{NewBasicTypes(), {ID: 2, Str: "second"}}
	sliceValue, err := helper.GetSliceObjectValue(original)
	if err != nil {
		t.Fatalf("GetSliceObjectValue failed: %v", err)
	}
	jsonData, err := remote.EncodeSliceObjectValue(sliceValue)
	if err != nil {
		t.Fatalf("EncodeSliceObjectValue failed: %v", err)
	}
	decoded, err := remote.DecodeSliceObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeSliceObjectValue failed: %v", err)
	}
	if !remote.CompareSliceObjectValue(sliceValue, decoded) {
		t.Error("design 8.4.2: SliceObjectValue marshal/unmarshal 后与起点不一致")
	}
}

// TestDesignRoundTripRemoteObject 对应设计 8.4.2：Object（结构定义）→ marshal → unmarshal → Object 完全一致。
func TestDesignRoundTripRemoteObject(t *testing.T) {
	entity := NewBasicTypes()
	obj, err := helper.GetObject(entity)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	jsonData, err := helper.EncodeObject(obj)
	if err != nil {
		t.Fatalf("EncodeObject failed: %v", err)
	}
	decoded, err := helper.DecodeObject(jsonData)
	if err != nil {
		t.Fatalf("DecodeObject failed: %v", err)
	}
	if !remote.CompareObject(obj, decoded) {
		t.Error("design 8.4.2: Object marshal/unmarshal 后与起点不一致")
	}
}

// TestDesignRoundTripNestedAndSliceObject 对应设计 8.4：属性为对象、属性为对象切片的实体，以及整体为实体切片的往返一致性。
// 覆盖：ComplexEntity（Child + Items）、[]*NestedParent（元素含 Child）的 Local→Remote→JSON→Remote→Local。
func TestDesignRoundTripNestedAndSliceObject(t *testing.T) {
	// 1) 实体含嵌套对象 + 对象切片：ComplexEntity
	origEntity := NewComplexEntity()
	objVal, err := helper.GetObjectValue(origEntity)
	if err != nil {
		t.Fatalf("GetObjectValue(ComplexEntity) failed: %v", err)
	}
	jsonData, err := remote.EncodeObjectValue(objVal)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}
	decodedVal, err := remote.DecodeObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}
	targetEntity := &ComplexEntity{}
	if err = helper.UpdateEntity(decodedVal, targetEntity); err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}
	if !compareComplexEntity(origEntity, targetEntity) {
		t.Error("design 8.4: ComplexEntity (nested object + slice of objects) roundtrip: result not equal to original")
	}

	// 2) 整体为实体切片且元素含嵌套对象：[]*NestedParent
	origSlice := []*NestedParent{
		NewNestedParent(),
		{ID: 2, Name: "p2", Child: &NestedChild{ID: 20, Name: "c2"}},
	}
	sliceVal, err := helper.GetSliceObjectValue(origSlice)
	if err != nil {
		t.Fatalf("GetSliceObjectValue([]*NestedParent) failed: %v", err)
	}
	sliceJSON, err := remote.EncodeSliceObjectValue(sliceVal)
	if err != nil {
		t.Fatalf("EncodeSliceObjectValue failed: %v", err)
	}
	decodedSlice, err := remote.DecodeSliceObjectValue(sliceJSON)
	if err != nil {
		t.Fatalf("DecodeSliceObjectValue failed: %v", err)
	}
	var targetSlice []*NestedParent
	if err = helper.UpdateSliceEntity(decodedSlice, &targetSlice); err != nil {
		t.Fatalf("UpdateSliceEntity failed: %v", err)
	}
	if len(targetSlice) != len(origSlice) {
		t.Fatalf("slice length: expected %d, got %d", len(origSlice), len(targetSlice))
	}
	for i := range origSlice {
		if !compareNestedParent(origSlice[i], targetSlice[i]) {
			t.Errorf("design 8.4: []*NestedParent[%d] roundtrip: result not equal to original", i)
		}
	}
}

// TestMemberTypeSliceOfPointerModelFieldTypeValue 对应设计 5.4 成员类型为 []*T 的补充说明：不要求 item 为 nil，但需保证 []*T 成员在 Model/Field/Type/Value 上符合预期，且对象值相互转换正确。
func TestMemberTypeSliceOfPointerModelFieldTypeValue(t *testing.T) {
	entity := NewNestedSlicePtrParent()

	// Local：Model / Field / Type / Value 符合预期
	localModel, err := local.GetEntityModel(entity, nil)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}
	childrenField := localModel.GetField("children")
	if childrenField == nil {
		t.Fatal("Local: GetField(\"children\") should not be nil")
	}
	if !models.IsSliceField(childrenField) {
		t.Error("Local: field \"children\" should be slice type (IsSliceField)")
	}
	elemType := childrenField.GetType().Elem()
	if elemType == nil {
		t.Error("Local: Field.GetType().Elem() for children should not be nil")
	}
	if elemType != nil && !models.IsStructType(elemType.GetValue()) {
		t.Errorf("Local: children element type should be struct type, got %v", elemType.GetValue())
	}
	childrenVal := childrenField.GetValue()
	if childrenVal == nil {
		t.Fatal("Local: GetValue() for children should not be nil")
	}
	if childrenVal != nil && !childrenVal.IsValid() {
		t.Error("Local: children Value should be valid")
	}
	unpacked := childrenVal.UnpackValue()
	if len(unpacked) != len(entity.Children) {
		t.Errorf("Local: UnpackValue() length expected %d, got %d", len(entity.Children), len(unpacked))
	}

	// Remote：Object 赋 ObjectValue 后，Model / Field / Type / Value 符合预期
	obj, err := helper.GetObject(entity)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	objValue, err := helper.GetObjectValue(entity)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}
	_, err = remote.SetModelValue(obj, remote.NewValue(objValue), true)
	if err != nil {
		t.Fatalf("SetModelValue failed: %v", err)
	}
	remoteChildrenField := obj.GetField("children")
	if remoteChildrenField == nil {
		t.Fatal("Remote: GetField(\"children\") should not be nil")
	}
	if !models.IsSliceField(remoteChildrenField) {
		t.Error("Remote: field \"children\" should be slice type (IsSliceField)")
	}
	remoteElemType := remoteChildrenField.GetType().Elem()
	if remoteElemType == nil {
		t.Error("Remote: Field.GetType().Elem() for children should not be nil")
	}
	if remoteElemType != nil && !models.IsStructType(remoteElemType.GetValue()) {
		t.Errorf("Remote: children element type should be struct type, got %v", remoteElemType.GetValue())
	}
	remoteChildrenVal := remoteChildrenField.GetValue()
	if remoteChildrenVal == nil {
		t.Fatal("Remote: GetValue() for children should not be nil")
	}
	remoteUnpacked := remoteChildrenVal.UnpackValue()
	if len(remoteUnpacked) != len(entity.Children) {
		t.Errorf("Remote: UnpackValue() length expected %d, got %d", len(entity.Children), len(remoteUnpacked))
	}

	// 相互转换：Local→Remote→Local 往返后与起点一致（设计 5.4 []*T 对象值相互转换）
	jsonData, err := remote.EncodeObjectValue(objValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}
	decoded, err := remote.DecodeObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}
	var target NestedSlicePtrParent
	if err = helper.UpdateEntity(decoded, &target); err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}
	if !compareNestedSlicePtrParent(entity, &target) {
		t.Error("design 5.4: []*T member object value conversion Local→Remote→Local: result not equal to original")
	}
}

// TestDesignSetModelValueObjectValue 对应设计 7.4.5/3.5：外部通过 SetModelValue(vModel, NewValue(objVal)) 将 ObjectValue 赋给 Object。
func TestDesignSetModelValueObjectValue(t *testing.T) {
	entity := NewBasicTypes()
	obj, err := helper.GetObject(entity)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	objValue, err := helper.GetObjectValue(entity)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}
	_, err = remote.SetModelValue(obj, remote.NewValue(objValue), true)
	if err != nil {
		t.Fatalf("SetModelValue failed: %v", err)
	}
	// 通过 Model 读回，与 ObjectValue 一致
	for _, fv := range objValue.GetValue() {
		name := fv.GetName()
		expected := fv.Get()
		field := obj.GetField(name)
		if field == nil {
			t.Errorf("Object.GetField(%q) nil", name)
			continue
		}
		if field.GetValue() == nil {
			t.Errorf("Object field %q GetValue() nil", name)
			continue
		}
		actual := field.GetValue().Get()
		if !valuesEqual(expected, actual) {
			t.Errorf("field %s: expected %v, got %v", name, expected, actual)
		}
	}
}

// valuesEqual 比较两值是否一致（设计要求 Remote 与 Local 均使用 bool 表示 TypeBooleanValue，见设计 10.2 修复后）。
func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return reflect.DeepEqual(a, b)
}
