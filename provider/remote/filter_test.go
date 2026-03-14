package remote

import (
	"encoding/json"
	"testing"

	"github.com/muidea/magicOrm/models"
)

func testRemoteFilterObject() *Object {
	return &Object{
		Name:    "User",
		PkgPath: "github.com/test/pkg",
		Fields: []*Field{
			{
				Name: "id",
				Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &SpecImpl{FieldName: "id", PrimaryKey: true},
			},
			{
				Name: "name",
				Type: &TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &SpecImpl{FieldName: "name"},
			},
			{
				Name: "group",
				Type: &TypeImpl{
					Name:    "Group",
					PkgPath: "github.com/test/pkg",
					Value:   models.TypeSliceValue,
					ElemType: &TypeImpl{
						Name:    "Group",
						PkgPath: "github.com/test/pkg",
						Value:   models.TypeStructValue,
						IsPtr:   true,
					},
				},
				Spec: &SpecImpl{FieldName: "group"},
			},
		},
	}
}

func TestObjectFilterInAcceptsSliceObjectValue(t *testing.T) {
	filter := NewFilter(testRemoteFilterObject())
	groupValue := &SliceObjectValue{
		Name:    "Group",
		PkgPath: "github.com/test/pkg",
		Values: []*ObjectValue{
			{
				ID:      "1",
				Name:    "Group",
				PkgPath: "github.com/test/pkg",
				Fields: []*FieldValue{
					{Name: "id", Value: int64(1)},
				},
			},
		},
	}

	if err := filter.In("group", groupValue); err != nil {
		t.Fatalf("In should accept *SliceObjectValue, got %v", err)
	}

	item := filter.GetFilterItem("group")
	if item == nil {
		t.Fatal("GetFilterItem(group) should not be nil")
	}
	got, ok := item.OprValue().Get().(*SliceObjectValue)
	if !ok {
		t.Fatalf("expected filter item value to be *SliceObjectValue, got %T", item.OprValue().Get())
	}
	if got.Values == nil || len(got.Values) != 1 || got.Values[0].ID != "1" {
		t.Fatalf("unexpected filter item value: %#v", got)
	}
}

func TestObjectFilterMaskModelDoesNotMutateBoundObject(t *testing.T) {
	object := testRemoteFilterObject()
	filter := NewFilter(object)

	if err := filter.ValueMask(&ObjectValue{
		Name:    "User",
		PkgPath: "github.com/test/pkg",
		Fields: []*FieldValue{
			{Name: "name", Value: "masked"},
		},
	}); err != nil {
		t.Fatalf("ValueMask failed: %v", err)
	}

	maskedModel := filter.MaskModel()
	if maskedModel == nil {
		t.Fatal("MaskModel should not return nil")
	}
	maskedName := maskedModel.GetField("name")
	if maskedName == nil || !maskedName.GetValue().IsValid() || maskedName.GetValue().Get() != "masked" {
		t.Fatalf("masked model should contain masked field value, got %#v", maskedName)
	}

	boundName := object.GetField("name")
	if boundName == nil {
		t.Fatal("bound object should still expose name field")
	}
	if boundName.GetValue().IsValid() {
		t.Fatalf("MaskModel should not mutate bound object, got %#v", boundName.GetValue().Get())
	}
}

func TestObjectFilterValueMaskRejectsMismatchedModel(t *testing.T) {
	filter := NewFilter(testRemoteFilterObject())

	err := filter.ValueMask(&ObjectValue{
		Name:    "OtherUser",
		PkgPath: "github.com/test/other",
		Fields: []*FieldValue{
			{Name: "name", Value: "masked"},
		},
	})
	if err == nil {
		t.Fatal("ValueMask should reject mismatched model value")
	}
}

func TestObjectFilterOperationsAndHelpers(t *testing.T) {
	filter := NewFilter(testRemoteFilterObject())
	if filter.GetName() != "User" || filter.GetPkgPath() != "github.com/test/pkg" {
		t.Fatalf("filter identity mismatch, got %s %s", filter.GetName(), filter.GetPkgPath())
	}
	if filter.Paginationer() != nil || filter.Sorter() != nil {
		t.Fatal("empty filter should not expose pagination or sorter")
	}

	if err := filter.Equal("name", "alice"); err != nil {
		t.Fatalf("Equal failed: %v", err)
	}
	if got, ok := filter.GetString("name"); !ok || got != "alice" {
		t.Fatalf("GetString mismatch, got %q ok=%v", got, ok)
	}

	if err := filter.Equal("id", float64(12)); err != nil {
		t.Fatalf("Equal(id) failed: %v", err)
	}
	filter.EqualFilter = append(filter.EqualFilter, &FieldValue{Name: "id_raw", Value: float64(12)})
	if got, ok := filter.GetInt("id_raw"); !ok || got != 12 {
		t.Fatalf("GetInt mismatch, got %d ok=%v", got, ok)
	}

	if err := filter.NotEqual("name", "bob"); err != nil {
		t.Fatalf("NotEqual failed: %v", err)
	}
	if item := filter.GetFilterItem("name"); item == nil || item.OprCode() != models.EqualOpr {
		t.Fatalf("GetFilterItem should prefer equal filter, got %#v", item)
	}

	statusOnly := NewFilter(testRemoteFilterObject())
	if err := statusOnly.NotEqual("name", "bob"); err != nil {
		t.Fatalf("NotEqual(name) failed: %v", err)
	}
	if item := statusOnly.GetFilterItem("name"); item == nil || item.OprCode() != models.NotEqualOpr {
		t.Fatalf("GetFilterItem(not equal) mismatch, got %#v", item)
	}

	belowFilter := NewFilter(testRemoteFilterObject())
	if err := belowFilter.Below("id", 5); err != nil {
		t.Fatalf("Below failed: %v", err)
	}
	if item := belowFilter.GetFilterItem("id"); item == nil || item.OprCode() != models.BelowOpr || item.OprValue().Get() != int64(5) {
		t.Fatalf("Below filter mismatch, got %#v", item)
	}

	aboveFilter := NewFilter(testRemoteFilterObject())
	if err := aboveFilter.Above("id", 7); err != nil {
		t.Fatalf("Above failed: %v", err)
	}
	if item := aboveFilter.GetFilterItem("id"); item == nil || item.OprCode() != models.AboveOpr || item.OprValue().Get() != int64(7) {
		t.Fatalf("Above filter mismatch, got %#v", item)
	}

	notInFilter := NewFilter(testRemoteFilterObject())
	if err := notInFilter.NotIn("group", &SliceObjectValue{Name: "Group", PkgPath: "github.com/test/pkg"}); err != nil {
		t.Fatalf("NotIn failed: %v", err)
	}
	if item := notInFilter.GetFilterItem("group"); item == nil || item.OprCode() != models.NotInOpr {
		t.Fatalf("NotIn filter mismatch, got %#v", item)
	}

	likeFilter := NewFilter(testRemoteFilterObject())
	if err := likeFilter.Like("name", "ali%"); err != nil {
		t.Fatalf("Like failed: %v", err)
	}
	if item := likeFilter.GetFilterItem("name"); item == nil || item.OprCode() != models.LikeOpr || item.OprValue().Get() != "ali%" {
		t.Fatalf("Like filter mismatch, got %#v", item)
	}

	filter.Pagination(2, 5)
	page := filter.Paginationer()
	if page == nil || page.Limit() != 5 || page.Offset() != 5 {
		t.Fatalf("Pagination mismatch, got %#v", page)
	}

	filter.Sort("name", true)
	sorter := filter.Sorter()
	if sorter == nil || sorter.Name() != "name" || !sorter.AscSort() {
		t.Fatalf("Sorter mismatch, got %#v", sorter)
	}
}

func TestObjectFilterValueMaskRawMessageAndMissingKeys(t *testing.T) {
	filter := NewFilter(testRemoteFilterObject())
	maskBytes, err := json.Marshal(&ObjectValue{
		Name:    "User",
		PkgPath: "github.com/test/pkg",
		Fields: []*FieldValue{
			{Name: "name", Value: "raw"},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if err := filter.ValueMask(json.RawMessage(maskBytes)); err != nil {
		t.Fatalf("ValueMask(raw message) failed: %v", err)
	}

	maskedModel := filter.MaskModel()
	if got := maskedModel.GetField("name").GetValue().Get(); got != "raw" {
		t.Fatalf("MaskModel(raw message) mismatch, got %#v", got)
	}

	if item := filter.GetFilterItem("missing"); item != nil {
		t.Fatalf("GetFilterItem(missing) should be nil, got %#v", item)
	}
	if got, ok := filter.GetString("missing"); ok || got != "" {
		t.Fatalf("GetString(missing) mismatch, got %q ok=%v", got, ok)
	}
	if got, ok := filter.GetInt("missing"); ok || got != 0 {
		t.Fatalf("GetInt(missing) mismatch, got %d ok=%v", got, ok)
	}
}

func TestObjectFilterValueMaskAdditionalBranches(t *testing.T) {
	filter := NewFilter(testRemoteFilterObject())

	var nilMask *ObjectValue
	if err := filter.ValueMask(nilMask); err != nil {
		t.Fatalf("ValueMask(nil *ObjectValue) should be ignored, got %v", err)
	}
	if filter.MaskValue != nil {
		t.Fatalf("ValueMask(nil *ObjectValue) should not set mask, got %#v", filter.MaskValue)
	}

	if err := filter.ValueMask(123); err == nil {
		t.Fatal("ValueMask(illegal type) should fail")
	}
	if err := filter.ValueMask(nil); err == nil {
		t.Fatal("ValueMask(nil) should fail")
	}
}
