package remote

import (
	"encoding/json"
	"testing"

	"github.com/muidea/magicOrm/models"
)

func buildRewardPolicyObjectValue() *ObjectValue {
	orderValue := &ObjectValue{
		ID:      "11",
		Name:    "valueItem",
		PkgPath: "/vmi/bill/rewardPolicy",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(11)},
			{Name: "level", Value: 1},
			{Name: "type", Value: 1},
			{Name: "value", Value: 12.5},
		},
	}

	return &ObjectValue{
		ID:      "2001",
		Name:    "rewardPolicy",
		PkgPath: "/vmi/bill",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(2001)},
			{Name: "name", Value: "promotion"},
			{Name: "description", Value: "order reward"},
			{Name: "partner", Value: 5.5},
			{Name: "order", Value: orderValue},
			{
				Name: "item",
				Value: &SliceObjectValue{
					Name:    "valueItem",
					PkgPath: "/vmi/bill/rewardPolicy",
					Values: []*ObjectValue{
						orderValue,
						{
							ID:      "12",
							Name:    "valueItem",
							PkgPath: "/vmi/bill/rewardPolicy",
							Fields: []*FieldValue{
								{Name: "id", Value: int64(12)},
								{Name: "level", Value: 2},
								{Name: "type", Value: 1},
								{Name: "value", Value: 18.75},
							},
						},
					},
				},
			},
			{
				Name: "scope",
				Value: &ObjectValue{
					ID:      "21",
					Name:    "valueScope",
					PkgPath: "/vmi/bill/rewardPolicy",
					Fields: []*FieldValue{
						{Name: "id", Value: int64(21)},
						{Name: "lowValue", Value: 100.0},
						{Name: "highValue", Value: 999.0},
					},
				},
			},
			{
				Name: "status",
				Value: &ObjectValue{
					ID:      "9",
					Name:    "status",
					PkgPath: "/vmi",
					Fields: []*FieldValue{
						{Name: "id", Value: int64(9)},
						{Name: "value", Value: 2},
						{Name: "name", Value: "published"},
					},
				},
			},
		},
	}
}

func TestDecodeObjectValuePreservesNilBackedSliceRelationShell(t *testing.T) {
	product := loadVMIObject(t, "test/vmi/entity/product/product.json")

	objectValue := &ObjectValue{
		ID:      "1001",
		Name:    product.GetName(),
		PkgPath: product.GetPkgPath(),
		Fields: []*FieldValue{
			{
				Name: "skuInfo",
				Value: &SliceObjectValue{
					Name:    "skuInfo",
					PkgPath: "/vmi/product",
					Values:  nil,
				},
			},
		},
	}

	encoded, err := EncodeObjectValue(objectValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}

	decoded, err := DecodeObjectValue(encoded)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}

	skuInfoValue, ok := decoded.GetFieldValue("skuInfo").(*SliceObjectValue)
	if !ok {
		t.Fatalf("decoded skuInfo should remain *SliceObjectValue, got %#v", decoded.GetFieldValue("skuInfo"))
	}
	if skuInfoValue.Name != "skuInfo" || skuInfoValue.PkgPath != "/vmi/product" {
		t.Fatalf("decoded skuInfo shell mismatch, got %#v", skuInfoValue)
	}
	if skuInfoValue.Values != nil {
		t.Fatalf("decoded skuInfo should preserve nil-backed relation shell, got %#v", skuInfoValue.Values)
	}
}

func TestConvertObjectValuePreservesMapBackedNilSliceRelationShell(t *testing.T) {
	product := loadVMIObject(t, "test/vmi/entity/product/product.json")

	objectValue := &ObjectValue{
		Name:    product.GetName(),
		PkgPath: product.GetPkgPath(),
		Fields: []*FieldValue{
			{
				Name: "skuInfo",
				Value: map[string]any{
					NameTag:    "skuInfo",
					PkgPathTag: "/vmi/product",
					ValuesTag:  nil,
				},
			},
		},
	}

	converted, err := ConvertObjectValue(objectValue)
	if err != nil {
		t.Fatalf("ConvertObjectValue failed: %v", err)
	}

	skuInfoValue, ok := converted.GetFieldValue("skuInfo").(*SliceObjectValue)
	if !ok {
		t.Fatalf("converted skuInfo should remain *SliceObjectValue, got %#v", converted.GetFieldValue("skuInfo"))
	}
	if skuInfoValue.Name != "skuInfo" || skuInfoValue.PkgPath != "/vmi/product" {
		t.Fatalf("converted skuInfo shell mismatch, got %#v", skuInfoValue)
	}
	if skuInfoValue.Values != nil {
		t.Fatalf("converted skuInfo should preserve nil-backed relation shell, got %#v", skuInfoValue.Values)
	}
}

func TestCompareObjectValueIncludesID(t *testing.T) {
	left := &ObjectValue{
		ID:      "1001",
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*FieldValue{
			{Name: "name", Value: "apple"},
		},
	}
	right := &ObjectValue{
		ID:      "1002",
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*FieldValue{
			{Name: "name", Value: "apple"},
		},
	}

	if CompareObjectValue(left, right) {
		t.Fatalf("CompareObjectValue should treat different IDs as different objects")
	}
}

func TestDecodeObjectValueRewardPolicyNestedRoundTrip(t *testing.T) {
	rewardPolicyValue := buildRewardPolicyObjectValue()

	encoded, err := EncodeObjectValue(rewardPolicyValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue(rewardPolicy) failed: %v", err)
	}

	decoded, err := DecodeObjectValue(encoded)
	if err != nil {
		t.Fatalf("DecodeObjectValue(rewardPolicy) failed: %v", err)
	}

	if !CompareObjectValue(rewardPolicyValue, decoded) {
		t.Fatalf("DecodeObjectValue(rewardPolicy) mismatch, got %#v", decoded)
	}
}

func TestConvertObjectValueRewardPolicyNestedRoundTrip(t *testing.T) {
	rewardPolicyValue := buildRewardPolicyObjectValue()

	encoded, err := EncodeObjectValue(rewardPolicyValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue(rewardPolicy) failed: %v", err)
	}

	rawValue := &ObjectValue{}
	if err := json.Unmarshal(encoded, rawValue); err != nil {
		t.Fatalf("json.Unmarshal(raw rewardPolicy) failed: %v", err)
	}

	converted, err := ConvertObjectValue(rawValue)
	if err != nil {
		t.Fatalf("ConvertObjectValue(rewardPolicy) failed: %v", err)
	}

	if !CompareObjectValue(rewardPolicyValue, converted) {
		t.Fatalf("ConvertObjectValue(rewardPolicy) mismatch, got %#v", converted)
	}
}

func TestDecodeSliceObjectValueRewardPolicyRoundTrip(t *testing.T) {
	sliceValue := &SliceObjectValue{
		Name:    "rewardPolicy",
		PkgPath: "/vmi/bill",
		Values: []*ObjectValue{
			buildRewardPolicyObjectValue(),
		},
	}

	encoded, err := EncodeSliceObjectValue(sliceValue)
	if err != nil {
		t.Fatalf("EncodeSliceObjectValue(rewardPolicy) failed: %v", err)
	}

	decoded, err := DecodeSliceObjectValue(encoded)
	if err != nil {
		t.Fatalf("DecodeSliceObjectValue(rewardPolicy) failed: %v", err)
	}

	if !CompareSliceObjectValue(sliceValue, decoded) {
		t.Fatalf("DecodeSliceObjectValue(rewardPolicy) mismatch, got %#v", decoded)
	}
}

func TestDecodeObjectValuePreservesNilBackedStructRelationShell(t *testing.T) {
	objectValue := &ObjectValue{
		ID:      "1001",
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*FieldValue{
			{
				Name: "status",
				Value: &ObjectValue{
					Name:    "status",
					PkgPath: "/vmi",
					Fields:  nil,
				},
			},
		},
	}

	encoded, err := EncodeObjectValue(objectValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue(status shell) failed: %v", err)
	}

	decoded, err := DecodeObjectValue(encoded)
	if err != nil {
		t.Fatalf("DecodeObjectValue(status shell) failed: %v", err)
	}

	statusValue, ok := decoded.GetFieldValue("status").(*ObjectValue)
	if !ok {
		t.Fatalf("decoded status should remain *ObjectValue, got %#v", decoded.GetFieldValue("status"))
	}
	if statusValue.Name != "status" || statusValue.PkgPath != "/vmi" {
		t.Fatalf("decoded status shell mismatch, got %#v", statusValue)
	}
	if statusValue.Fields != nil {
		t.Fatalf("decoded status should preserve nil-backed struct shell, got %#v", statusValue.Fields)
	}
}

func TestConvertObjectValueConvertsBasicSlicesFromRawJSON(t *testing.T) {
	objectValue := &ObjectValue{
		Name:    "raw",
		PkgPath: "/vmi/test",
		Fields: []*FieldValue{
			{Name: "names", Value: []any{"alpha", "beta"}},
			{Name: "flags", Value: []any{true, false}},
			{Name: "scores", Value: []any{1.5, 2.5}},
		},
	}

	converted, err := ConvertObjectValue(objectValue)
	if err != nil {
		t.Fatalf("ConvertObjectValue(raw basic slices) failed: %v", err)
	}

	names, ok := converted.GetFieldValue("names").([]string)
	if !ok || len(names) != 2 || names[0] != "alpha" || names[1] != "beta" {
		t.Fatalf("names should convert to []string, got %#v", converted.GetFieldValue("names"))
	}
	flags, ok := converted.GetFieldValue("flags").([]bool)
	if !ok || len(flags) != 2 || !flags[0] || flags[1] {
		t.Fatalf("flags should convert to []bool, got %#v", converted.GetFieldValue("flags"))
	}
	scores, ok := converted.GetFieldValue("scores").([]float64)
	if !ok || len(scores) != 2 || scores[0] != 1.5 || scores[1] != 2.5 {
		t.Fatalf("scores should convert to []float64, got %#v", converted.GetFieldValue("scores"))
	}
}

func TestObjectSetFieldValueStructAndSliceStruct(t *testing.T) {
	product := loadVMIObject(t, "test/vmi/entity/product/product.json")

	statusValue := ObjectValue{
		ID:      "9",
		Name:    "status",
		PkgPath: "/vmi",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(9)},
			{Name: "name", Value: "published"},
		},
	}
	if err := product.SetFieldValue("status", statusValue); err != nil {
		t.Fatalf("SetFieldValue(status) failed: %v", err)
	}

	skuInfoValue := SliceObjectValue{
		Name:    "skuInfo",
		PkgPath: "/vmi/product",
		Values: []*ObjectValue{
			{
				ID:      "sku-001",
				Name:    "skuInfo",
				PkgPath: "/vmi/product",
				Fields: []*FieldValue{
					{Name: "sku", Value: "sku-001"},
				},
			},
		},
	}
	if err := product.SetFieldValue("skuInfo", skuInfoValue); err != nil {
		t.Fatalf("SetFieldValue(skuInfo) failed: %v", err)
	}

	gotStatus, ok := product.GetField("status").GetValue().Get().(*ObjectValue)
	if !ok || !CompareObjectValue(&statusValue, gotStatus) {
		t.Fatalf("status field mismatch, got %#v", product.GetField("status").GetValue().Get())
	}
	gotSKUInfo, ok := product.GetField("skuInfo").GetValue().Get().(*SliceObjectValue)
	if !ok || !CompareSliceObjectValue(&skuInfoValue, gotSKUInfo) {
		t.Fatalf("skuInfo field mismatch, got %#v", product.GetField("skuInfo").GetValue().Get())
	}

	if err := product.SetFieldValue("status", "illegal"); err == nil {
		t.Fatalf("SetFieldValue(status, string) should fail")
	}
	if err := product.SetFieldValue("skuInfo", "illegal"); err == nil {
		t.Fatalf("SetFieldValue(skuInfo, string) should fail")
	}
}

func TestObjectSetFieldValueBasicAndBasicSlice(t *testing.T) {
	product := loadVMIObject(t, "test/vmi/entity/product/product.json")

	if err := product.SetFieldValue("name", "apple"); err != nil {
		t.Fatalf("SetFieldValue(name) failed: %v", err)
	}
	if err := product.SetFieldValue("expire", float64(30)); err != nil {
		t.Fatalf("SetFieldValue(expire) failed: %v", err)
	}
	if err := product.SetFieldValue("image", []string{"main.png", "thumb.png"}); err != nil {
		t.Fatalf("SetFieldValue(image) failed: %v", err)
	}

	if got := product.GetField("name").GetValue().Get(); got != "apple" {
		t.Fatalf("name field mismatch, got %#v", got)
	}
	if got := product.GetField("expire").GetValue().Get(); got != 30 {
		t.Fatalf("expire field mismatch, got %#v", got)
	}
	imageValue, ok := product.GetField("image").GetValue().Get().([]string)
	if !ok || len(imageValue) != 2 || imageValue[0] != "main.png" || imageValue[1] != "thumb.png" {
		t.Fatalf("image field mismatch, got %#v", product.GetField("image").GetValue().Get())
	}

	if err := product.SetFieldValue("expire", "illegal"); err == nil {
		t.Fatalf("SetFieldValue(expire, string) should fail")
	}
}

func TestObjectResetAndVerify(t *testing.T) {
	product := loadVMIObject(t, "test/vmi/entity/product/product.json")
	if product.GetShowName() != "产品信息" {
		t.Fatalf("GetShowName mismatch, got %q", product.GetShowName())
	}
	if err := product.SetFieldValue("name", "apple"); err != nil {
		t.Fatalf("SetFieldValue(name) failed: %v", err)
	}
	if err := product.SetFieldValue("image", []string{"main.png"}); err != nil {
		t.Fatalf("SetFieldValue(image) failed: %v", err)
	}
	if err := product.SetFieldValue("status", &ObjectValue{
		ID:      "9",
		Name:    "status",
		PkgPath: "/vmi",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(9)},
		},
	}); err != nil {
		t.Fatalf("SetFieldValue(status) failed: %v", err)
	}

	product.Reset()
	if product.GetField("name").GetValue().Get() != "" {
		t.Fatalf("Reset should clear basic string field, got %#v", product.GetField("name").GetValue().Get())
	}
	imageValue, ok := product.GetField("image").GetValue().Get().([]string)
	if !ok || imageValue == nil || len(imageValue) != 0 {
		t.Fatalf("Reset should clear basic slice field, got %#v", product.GetField("image").GetValue().Get())
	}
	statusValue, ok := product.GetField("status").GetValue().Get().(*ObjectValue)
	if !ok || statusValue == nil || statusValue.Name != "status" || statusValue.PkgPath != "/vmi" {
		t.Fatalf("Reset should restore struct shell, got %#v", product.GetField("status").GetValue().Get())
	}

	if err := product.Verify(); err != nil {
		t.Fatalf("Verify(valid product) failed: %v", err)
	}

	invalidObject := &Object{}
	if err := invalidObject.Verify(); err == nil {
		t.Fatal("Verify should reject object without name")
	}

	invalidFieldObject := &Object{
		Name:    "broken",
		PkgPath: "/vmi",
		Fields: []*Field{
			{
				Name: "id",
				Type: &TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &SpecImpl{PrimaryKey: true, ValueDeclare: models.AutoIncrement},
			},
		},
	}
	if err := invalidFieldObject.Verify(); err == nil {
		t.Fatal("Verify should reject invalid field declaration")
	}

	sliceObject := &SliceObjectValue{Name: "skuInfo", PkgPath: "/vmi/product"}
	if sliceObject.GetPkgKey() != "/vmi/product/skuInfo" {
		t.Fatalf("SliceObjectValue.GetPkgKey mismatch, got %q", sliceObject.GetPkgKey())
	}
}

func TestEncodeDecodeObjectValuePreservesAssignedState(t *testing.T) {
	raw := &ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(0), Assigned: true},
			{Name: "status", Value: nil, Assigned: true},
			{Name: "name", Value: "", Assigned: false},
		},
	}

	data, err := EncodeObjectValue(raw)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}

	decoded, err := DecodeObjectValue(data)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}

	if !decoded.Fields[0].Assigned || decoded.Fields[0].Value != float64(0) {
		t.Fatalf("explicit zero assigned state should be preserved, got %#v", decoded.Fields[0])
	}
	if !decoded.Fields[1].Assigned || decoded.Fields[1].Value != nil {
		t.Fatalf("explicit nil assigned state should be preserved, got %#v", decoded.Fields[1])
	}
	if decoded.Fields[2].Assigned {
		t.Fatalf("unassigned zero value should remain unassigned, got %#v", decoded.Fields[2])
	}
}
