package remote

import (
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

func TestObjectViewPrimaryAndCompareHelpers(t *testing.T) {
	liteOnlySpec := &SpecImpl{FieldName: "name", ViewDeclare: []models.ViewDeclare{models.LiteView}}
	detailOnlySpec := &SpecImpl{FieldName: "description", ViewDeclare: []models.ViewDeclare{models.DetailView}}
	object := &Object{
		Name:     "product",
		PkgPath:  "/vmi",
		viewSpec: models.LiteView,
		Fields: []*Field{
			{Name: "id", Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}, Spec: &SpecImpl{FieldName: "id", PrimaryKey: true, ViewDeclare: []models.ViewDeclare{models.LiteView}}, value: NewValue(int64(1001))},
			{Name: "name", Type: &TypeImpl{Name: "string", Value: models.TypeStringValue}, Spec: liteOnlySpec, value: NewValue("apple")},
			{Name: "description", Type: &TypeImpl{Name: "string", Value: models.TypeStringValue}, Spec: detailOnlySpec},
		},
	}

	if object.fieldInActiveView(nil) {
		t.Fatal("fieldInActiveView(nil) should be false")
	}
	if !object.fieldInActiveView(object.Fields[1]) {
		t.Fatal("fieldInActiveView(lite field) should be true")
	}
	if object.fieldInActiveView(object.Fields[2]) {
		t.Fatal("fieldInActiveView(detail field on lite model) should be false")
	}

	if err := object.SetPrimaryFieldValue(float64(2002)); err != nil {
		t.Fatalf("SetPrimaryFieldValue failed: %v", err)
	}
	if got := object.GetPrimaryField().GetValue().Get(); got != int64(2002) {
		t.Fatalf("SetPrimaryFieldValue mismatch, got %#v", got)
	}
	if err := object.innerSetPrimaryFieldValue(nil, false); err != nil {
		t.Fatalf("innerSetPrimaryFieldValue(nil) failed: %v", err)
	}
	if object.GetPrimaryField().GetValue().IsValid() {
		t.Fatalf("innerSetPrimaryFieldValue(nil) should clear primary value")
	}
	if err := object.SetPrimaryFieldValue(int64(3003)); err != nil {
		t.Fatalf("SetPrimaryFieldValue reset failed: %v", err)
	}

	exported, ok := object.Interface(false).(*ObjectValue)
	if !ok {
		t.Fatalf("Object.Interface should export *ObjectValue, got %#v", object.Interface(false))
	}
	if exported.ID != "3003" {
		t.Fatalf("Object.Interface primary mismatch, got %q", exported.ID)
	}
	if exported.GetFieldValue("description") != nil {
		t.Fatalf("Object.Interface should skip invalid detail field, got %#v", exported.GetFieldValue("description"))
	}

	objectWithoutPrimary := &Object{Name: "plain", PkgPath: "/vmi", Fields: []*Field{{Name: "name", Type: &TypeImpl{Name: "string", Value: models.TypeStringValue}, Spec: &SpecImpl{FieldName: "name"}, value: NewValue("plain")}}}
	exportedWithoutPrimary := objectWithoutPrimary.Interface(false).(*ObjectValue)
	if exportedWithoutPrimary.ID != "" {
		t.Fatalf("Object.Interface without primary should keep empty ID, got %q", exportedWithoutPrimary.ID)
	}

	if object.GetField("missing") != nil {
		t.Fatal("GetField(missing) should be nil")
	}
	if !CompareObject(object, object.Copy(models.OriginView).(*Object)) {
		t.Fatal("CompareObject on copied object should be true")
	}
	if CompareObject(object, &Object{Name: object.Name, PkgPath: "/other", Fields: object.Fields}) {
		t.Fatal("CompareObject different pkgPath should be false")
	}
	if CompareObject(object, &Object{Name: object.Name, PkgPath: object.PkgPath, Fields: object.Fields[:2]}) {
		t.Fatal("CompareObject different field count should be false")
	}

	emptySlice := &SliceObjectValue{Name: "skuInfo", PkgPath: "/vmi/product"}
	if emptySlice.IsAssigned() {
		t.Fatal("SliceObjectValue without values should be unassigned")
	}
	transferred := TransferObjectValue("skuInfo", "/vmi/product", []*ObjectValue{{Name: "skuInfo", PkgPath: "/vmi/product"}})
	if !transferred.IsAssigned() || transferred.GetPkgKey() != "/vmi/product/skuInfo" {
		t.Fatalf("TransferObjectValue mismatch, got %#v", transferred)
	}
}

func TestObjectDecodeAndMarshalErrorPaths(t *testing.T) {
	var fn func()
	if _, err := marshalHelper(&fn); err == nil {
		t.Fatal("marshalHelper(func) should fail")
	}

	if _, err := unmarshalHelper([]byte("{"), &ObjectValue{}, ConvertObjectValue); err == nil {
		t.Fatal("unmarshalHelper(invalid json) should fail")
	}
	if _, err := unmarshalHelper([]byte(`{"name":"product","pkgPath":"/vmi","fields":[]}`), &ObjectValue{}, func(val *ObjectValue) (*ObjectValue, *cd.Error) {
		return nil, cd.NewError(cd.Unexpected, "decode failed")
	}); err == nil {
		t.Fatal("unmarshalHelper(decode error) should fail")
	}

	if _, err := decodeObjectValueFromMap(map[string]any{}); err == nil {
		t.Fatal("decodeObjectValueFromMap(missing keys) should fail")
	}
	if _, err := decodeObjectValueFromMap(map[string]any{NameTag: 1, PkgPathTag: "/vmi", FieldsTag: []any{}}); err == nil {
		t.Fatal("decodeObjectValueFromMap(non-string name) should fail")
	}
	if _, err := decodeObjectValueFromMap(map[string]any{NameTag: "product", PkgPathTag: "/vmi", FieldsTag: "illegal"}); err == nil {
		t.Fatal("decodeObjectValueFromMap(non-slice fields) should fail")
	}
	if _, err := decodeObjectValueFromMap(map[string]any{NameTag: "product", PkgPathTag: "/vmi", FieldsTag: []any{"illegal"}}); err == nil {
		t.Fatal("decodeObjectValueFromMap(invalid field item) should fail")
	}

	if _, err := decodeSliceObjectValueFromMap(map[string]any{}); err == nil {
		t.Fatal("decodeSliceObjectValueFromMap(missing keys) should fail")
	}
	if _, err := decodeSliceObjectValueFromMap(map[string]any{NameTag: 1, PkgPathTag: "/vmi", ValuesTag: []any{}}); err == nil {
		t.Fatal("decodeSliceObjectValueFromMap(non-string name) should fail")
	}
	if _, err := decodeSliceObjectValueFromMap(map[string]any{NameTag: "skuInfo", PkgPathTag: "/vmi", ValuesTag: "illegal"}); err == nil {
		t.Fatal("decodeSliceObjectValueFromMap(non-slice values) should fail")
	}
	if _, err := decodeSliceObjectValueFromMap(map[string]any{NameTag: "skuInfo", PkgPathTag: "/vmi", ValuesTag: []any{"illegal"}}); err == nil {
		t.Fatal("decodeSliceObjectValueFromMap(invalid object item) should fail")
	}

	if _, err := decodeItemValue(nil); err == nil {
		t.Fatal("decodeItemValue(nil) should fail")
	}
	if _, err := decodeItemValue(map[string]any{}); err == nil {
		t.Fatal("decodeItemValue(missing name) should fail")
	}
	if _, err := decodeItemValue(map[string]any{NameTag: 1}); err == nil {
		t.Fatal("decodeItemValue(non-string name) should fail")
	}
	if _, err := decodeItemValue(map[string]any{NameTag: "status", ValueTag: map[string]any{NameTag: "status", PkgPathTag: "/vmi", FieldsTag: "illegal"}}); err == nil {
		t.Fatal("decodeItemValue(invalid nested object) should fail")
	}

	if got := convertAnySlice([]any{1.5, "bad"}); len(got.([]float64)) != 0 {
		t.Fatalf("convertAnySlice(mixed float) should return empty []float64, got %#v", got)
	}
	if got := convertAnySlice([]any{"ok", 1}); len(got.([]string)) != 0 {
		t.Fatalf("convertAnySlice(mixed string) should return empty []string, got %#v", got)
	}
	if got := convertAnySlice([]any{true, "bad"}); len(got.([]bool)) != 0 {
		t.Fatalf("convertAnySlice(mixed bool) should return empty []bool, got %#v", got)
	}
	if got := convertAnySlice([]any{1}); len(got.([]int)) != 1 || got.([]int)[0] != 1 {
		t.Fatalf("convertAnySlice(int) should return []int{1}, got %#v", got)
	}

	item, err := ConvertItem(&FieldValue{Name: "noop", Value: map[string]any{"foo": "bar"}})
	if err != nil {
		t.Fatalf("ConvertItem(non-object map) failed: %v", err)
	}
	if item != nil {
		t.Fatalf("ConvertItem(non-object map) should return nil, got %#v", item)
	}
}
