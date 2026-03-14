package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/remote"
)

func loadVMIRemoteObject(t *testing.T, relativePath string) *remote.Object {
	t.Helper()

	filePath := filepath.Join("..", relativePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile(%s) failed: %v", relativePath, err)
	}

	object := &remote.Object{}
	if err := json.Unmarshal(data, object); err != nil {
		t.Fatalf("Unmarshal(%s) failed: %v", relativePath, err)
	}

	return object
}

func loadAllVMIRemoteObjects(t *testing.T) []*remote.Object {
	t.Helper()

	root := filepath.Join("..", "test", "vmi", "entity")
	paths := make([]string, 0, 16)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		relativePath, err := filepath.Rel("..", path)
		if err != nil {
			return err
		}
		paths = append(paths, relativePath)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%s) failed: %v", root, err)
	}

	sort.Strings(paths)

	objects := make([]*remote.Object, 0, len(paths))
	for _, path := range paths {
		objects = append(objects, loadVMIRemoteObject(t, path))
	}

	return objects
}

func registerAllVMIRemoteObjects(t *testing.T, remoteProvider Provider) map[string]*remote.Object {
	t.Helper()

	objects := loadAllVMIRemoteObjects(t)
	result := make(map[string]*remote.Object, len(objects))
	for _, object := range objects {
		model, err := remoteProvider.RegisterModel(object)
		if err != nil {
			t.Fatalf("RegisterModel(%s) failed: %v", object.GetPkgKey(), err)
		}
		if model == nil {
			t.Fatalf("RegisterModel(%s) returned nil model", object.GetPkgKey())
		}

		result[object.GetPkgKey()] = object
	}

	return result
}

func requireRemoteObjectField(t *testing.T, object *remote.Object, fieldName string) models.Field {
	t.Helper()

	field := object.GetField(fieldName)
	if field == nil {
		t.Fatalf("%s.%s should exist", object.GetPkgKey(), fieldName)
	}

	return field
}

func requireRemoteFieldValue[T any](t *testing.T, model models.Model, fieldName string) T {
	t.Helper()

	field := model.GetField(fieldName)
	if field == nil {
		t.Fatalf("%s.%s should exist", model.GetPkgKey(), fieldName)
	}
	if field.GetValue() == nil || !field.GetValue().IsValid() {
		t.Fatalf("%s.%s should be assigned", model.GetPkgKey(), fieldName)
	}

	value, ok := field.GetValue().Get().(T)
	if !ok {
		t.Fatalf("%s.%s should be %T, got %T", model.GetPkgKey(), fieldName, *new(T), field.GetValue().Get())
	}

	return value
}

func requireRemoteFieldInvalid(t *testing.T, model models.Model, fieldName string) {
	t.Helper()

	field := model.GetField(fieldName)
	if field == nil {
		t.Fatalf("%s.%s should exist", model.GetPkgKey(), fieldName)
	}
	if field.GetValue() != nil && field.GetValue().IsValid() {
		t.Fatalf("%s.%s should stay invalid for current view, got %#v", model.GetPkgKey(), fieldName, field.GetValue().Get())
	}
}

func requireRemoteObjectValueField[T any](t *testing.T, objectValue *remote.ObjectValue, fieldName string) T {
	t.Helper()

	if objectValue == nil {
		t.Fatal("objectValue should not be nil")
	}

	value := objectValue.GetFieldValue(fieldName)
	if value == nil {
		t.Fatalf("%s.%s should be assigned", objectValue.GetPkgKey(), fieldName)
	}

	typedValue, ok := value.(T)
	if !ok {
		t.Fatalf("%s.%s should be %T, got %T", objectValue.GetPkgKey(), fieldName, *new(T), value)
	}

	return typedValue
}

func TestRemoteProviderRegistersAndFetchesAllVMIModels(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", nil)
	objects := registerAllVMIRemoteObjects(t, remoteProvider)

	for pkgKey, object := range objects {
		model, err := remoteProvider.GetEntityModel(object, true)
		if err != nil {
			t.Fatalf("GetEntityModel(%s) failed: %v", pkgKey, err)
		}
		if model == nil {
			t.Fatalf("GetEntityModel(%s) returned nil", pkgKey)
		}
		if model.GetPkgKey() != pkgKey {
			t.Fatalf("GetEntityModel(%s) got pkgKey=%s", pkgKey, model.GetPkgKey())
		}
	}
}

func TestRemoteProviderVMIResolvesReferencedTypeModels(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", nil)
	objects := registerAllVMIRemoteObjects(t, remoteProvider)

	cases := []struct {
		objectKey string
		fieldName string
		wantKey   string
	}{
		{objectKey: "/vmi/product", fieldName: "status", wantKey: "/vmi/status"},
		{objectKey: "/vmi/product", fieldName: "skuInfo", wantKey: "/vmi/product/skuInfo"},
		{objectKey: "/vmi/bill/rewardPolicy", fieldName: "scope", wantKey: "/vmi/bill/rewardPolicy/valueScope"},
		{objectKey: "/vmi/bill/rewardPolicy", fieldName: "item", wantKey: "/vmi/bill/rewardPolicy/valueItem"},
		{objectKey: "/vmi/store", fieldName: "goods", wantKey: "/vmi/store/goods"},
		{objectKey: "/vmi/store", fieldName: "shelf", wantKey: "/vmi/warehouse/shelf"},
	}

	for _, tc := range cases {
		object := objects[tc.objectKey]
		if object == nil {
			t.Fatalf("missing registered object %s", tc.objectKey)
		}

		field := requireRemoteObjectField(t, object, tc.fieldName)
		typeModel, err := remoteProvider.GetTypeModel(field.GetType())
		if err != nil {
			t.Fatalf("GetTypeModel(%s.%s) failed: %v", tc.objectKey, tc.fieldName, err)
		}
		if typeModel == nil {
			t.Fatalf("GetTypeModel(%s.%s) returned nil", tc.objectKey, tc.fieldName)
		}
		if typeModel.GetPkgKey() != tc.wantKey {
			t.Fatalf("GetTypeModel(%s.%s) got %s, want %s", tc.objectKey, tc.fieldName, typeModel.GetPkgKey(), tc.wantKey)
		}
	}
}

func TestRemoteProviderVMIProductRoundTripAndFilter(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", nil)
	registerAllVMIRemoteObjects(t, remoteProvider)

	statusValue := &remote.ObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(2)},
			{Name: "value", Value: 1},
			{Name: "name", Value: "enabled"},
		},
	}
	skuInfoValue := &remote.ObjectValue{
		Name:    "skuInfo",
		PkgPath: "/vmi/product",
		Fields: []*remote.FieldValue{
			{Name: "sku", Value: "sku-001"},
			{Name: "description", Value: "default sku"},
			{Name: "image", Value: []string{"sku-a.png", "sku-b.png"}},
		},
	}
	skuInfoSlice := &remote.SliceObjectValue{
		Name:    "skuInfo",
		PkgPath: "/vmi/product",
		Values:  []*remote.ObjectValue{skuInfoValue},
	}
	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
			{Name: "name", Value: "apple"},
			{Name: "description", Value: "fresh apple"},
			{Name: "skuInfo", Value: skuInfoSlice},
			{Name: "image", Value: []string{"main.png"}},
			{Name: "expire", Value: 30},
			{Name: "status", Value: statusValue},
			{Name: "tags", Value: []string{"fruit", "fresh"}},
		},
	}

	model, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(productValue) failed: %v", err)
	}

	gotValue, ok := model.Interface(true).(*remote.ObjectValue)
	if !ok {
		t.Fatalf("Interface(true) should return *remote.ObjectValue, got %T", model.Interface(true))
	}
	if gotValue.ID != "1001" {
		t.Fatalf("product ID mismatch, got %s", gotValue.ID)
	}
	if requireRemoteObjectValueField[int64](t, gotValue, "id") != int64(1001) {
		t.Fatalf("product.id mismatch, got %#v", gotValue.GetFieldValue("id"))
	}
	if requireRemoteObjectValueField[string](t, gotValue, "name") != "apple" {
		t.Fatalf("product.name mismatch, got %#v", gotValue.GetFieldValue("name"))
	}
	if requireRemoteObjectValueField[string](t, gotValue, "description") != "fresh apple" {
		t.Fatalf("product.description mismatch, got %#v", gotValue.GetFieldValue("description"))
	}
	if requireRemoteObjectValueField[int](t, gotValue, "expire") != 30 {
		t.Fatalf("product.expire mismatch, got %#v", gotValue.GetFieldValue("expire"))
	}
	gotImage := requireRemoteObjectValueField[[]string](t, gotValue, "image")
	if len(gotImage) != 1 || gotImage[0] != "main.png" {
		t.Fatalf("product.image mismatch, got %#v", gotImage)
	}
	gotTags := requireRemoteObjectValueField[[]string](t, gotValue, "tags")
	if len(gotTags) != 2 || gotTags[0] != "fruit" || gotTags[1] != "fresh" {
		t.Fatalf("product.tags mismatch, got %#v", gotTags)
	}
	gotStatusValue := requireRemoteObjectValueField[*remote.ObjectValue](t, gotValue, "status")
	if !remote.CompareObjectValue(statusValue, gotStatusValue) {
		t.Fatalf("product.status mismatch, got %#v", gotStatusValue)
	}
	gotSKUInfoValue := requireRemoteObjectValueField[*remote.SliceObjectValue](t, gotValue, "skuInfo")
	if !remote.CompareSliceObjectValue(skuInfoSlice, gotSKUInfoValue) {
		t.Fatalf("product.skuInfo mismatch, got %#v", gotSKUInfoValue)
	}

	filter, err := remoteProvider.GetModelFilter(model)
	if err != nil {
		t.Fatalf("GetModelFilter(productModel) failed: %v", err)
	}

	if err := filter.Equal("status", statusValue); err != nil {
		t.Fatalf("filter.Equal(status) failed: %v", err)
	}
	if err := filter.In("skuInfo", skuInfoSlice); err != nil {
		t.Fatalf("filter.In(skuInfo) failed: %v", err)
	}

	statusItem := filter.GetFilterItem("status")
	if statusItem == nil || statusItem.OprCode() != models.EqualOpr {
		t.Fatalf("status filter item mismatch: %#v", statusItem)
	}
	gotStatus, ok := statusItem.OprValue().Get().(*remote.ObjectValue)
	if !ok || !remote.CompareObjectValue(statusValue, gotStatus) {
		t.Fatalf("status filter value mismatch, got %#v", statusItem.OprValue().Get())
	}

	skuInfoItem := filter.GetFilterItem("skuInfo")
	if skuInfoItem == nil || skuInfoItem.OprCode() != models.InOpr {
		t.Fatalf("skuInfo filter item mismatch: %#v", skuInfoItem)
	}
	gotSKUInfo, ok := skuInfoItem.OprValue().Get().(*remote.SliceObjectValue)
	if !ok || !remote.CompareSliceObjectValue(skuInfoSlice, gotSKUInfo) {
		t.Fatalf("skuInfo filter value mismatch, got %#v", skuInfoItem.OprValue().Get())
	}

	maskValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{
				Name: "status",
				Value: &remote.ObjectValue{
					Name:    "status",
					PkgPath: "/vmi",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(2)},
					},
				},
			},
			{
				Name: "skuInfo",
				Value: &remote.SliceObjectValue{
					Name:    "skuInfo",
					PkgPath: "/vmi/product",
					Values:  []*remote.ObjectValue{},
				},
			},
		},
	}
	if err := filter.ValueMask(maskValue); err != nil {
		t.Fatalf("filter.ValueMask(product mask) failed: %v", err)
	}

	maskedModel := filter.MaskModel()
	if maskedModel == nil {
		t.Fatal("MaskModel() should not return nil")
	}

	maskedStatus := requireRemoteFieldValue[*remote.ObjectValue](t, maskedModel, "status")
	if maskedStatus.GetFieldValue("id") != int64(2) {
		t.Fatalf("masked status should keep explicit id, got %#v", maskedStatus)
	}

	maskedSKUInfo := requireRemoteFieldValue[*remote.SliceObjectValue](t, maskedModel, "skuInfo")
	if maskedSKUInfo.Values == nil || len(maskedSKUInfo.Values) != 0 {
		t.Fatalf("masked skuInfo should be assigned empty slice, got %#v", maskedSKUInfo)
	}

	originalSKUInfo := requireRemoteFieldValue[*remote.SliceObjectValue](t, model, "skuInfo")
	if len(originalSKUInfo.Values) != 1 {
		t.Fatalf("MaskModel should not mutate original skuInfo, got %#v", originalSKUInfo)
	}
}

func TestRemoteProviderVMIRewardPolicyRoundTrip(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", nil)
	registerAllVMIRemoteObjects(t, remoteProvider)

	statusValue := &remote.ObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(9)},
			{Name: "value", Value: 2},
			{Name: "name", Value: "published"},
		},
	}
	orderValue := &remote.ObjectValue{
		Name:    "valueItem",
		PkgPath: "/vmi/bill/rewardPolicy",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(11)},
			{Name: "level", Value: 1},
			{Name: "type", Value: 1},
			{Name: "value", Value: 12.5},
		},
	}
	itemSlice := &remote.SliceObjectValue{
		Name:    "valueItem",
		PkgPath: "/vmi/bill/rewardPolicy",
		Values: []*remote.ObjectValue{
			orderValue,
			{
				Name:    "valueItem",
				PkgPath: "/vmi/bill/rewardPolicy",
				Fields: []*remote.FieldValue{
					{Name: "id", Value: int64(12)},
					{Name: "level", Value: 2},
					{Name: "type", Value: 1},
					{Name: "value", Value: 18.75},
				},
			},
		},
	}
	scopeValue := &remote.ObjectValue{
		Name:    "valueScope",
		PkgPath: "/vmi/bill/rewardPolicy",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(21)},
			{Name: "lowValue", Value: 100.0},
			{Name: "highValue", Value: 999.0},
		},
	}
	rewardPolicyValue := &remote.ObjectValue{
		Name:    "rewardPolicy",
		PkgPath: "/vmi/bill",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(2001)},
			{Name: "name", Value: "promotion"},
			{Name: "description", Value: "order reward"},
			{Name: "partner", Value: 5.5},
			{Name: "order", Value: orderValue},
			{Name: "item", Value: itemSlice},
			{Name: "scope", Value: scopeValue},
			{Name: "ratio", Value: 1.25},
			{Name: "status", Value: statusValue},
		},
	}

	model, err := remoteProvider.GetEntityModel(rewardPolicyValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(rewardPolicyValue) failed: %v", err)
	}

	gotValue, ok := model.Interface(true).(*remote.ObjectValue)
	if !ok {
		t.Fatalf("Interface(true) should return *remote.ObjectValue, got %T", model.Interface(true))
	}
	if gotValue.ID != "2001" {
		t.Fatalf("rewardPolicy ID mismatch, got %s", gotValue.ID)
	}
	if requireRemoteObjectValueField[int64](t, gotValue, "id") != int64(2001) {
		t.Fatalf("rewardPolicy.id mismatch, got %#v", gotValue.GetFieldValue("id"))
	}
	if requireRemoteObjectValueField[string](t, gotValue, "name") != "promotion" {
		t.Fatalf("rewardPolicy.name mismatch, got %#v", gotValue.GetFieldValue("name"))
	}
	if requireRemoteObjectValueField[string](t, gotValue, "description") != "order reward" {
		t.Fatalf("rewardPolicy.description mismatch, got %#v", gotValue.GetFieldValue("description"))
	}
	if requireRemoteObjectValueField[float64](t, gotValue, "partner") != 5.5 {
		t.Fatalf("rewardPolicy.partner mismatch, got %#v", gotValue.GetFieldValue("partner"))
	}
	if requireRemoteObjectValueField[float64](t, gotValue, "ratio") != 1.25 {
		t.Fatalf("rewardPolicy.ratio mismatch, got %#v", gotValue.GetFieldValue("ratio"))
	}
	gotStatusValue := requireRemoteObjectValueField[*remote.ObjectValue](t, gotValue, "status")
	if !remote.CompareObjectValue(statusValue, gotStatusValue) {
		t.Fatalf("rewardPolicy.status mismatch, got %#v", gotStatusValue)
	}
	gotOrderValue := requireRemoteObjectValueField[*remote.ObjectValue](t, gotValue, "order")
	if !remote.CompareObjectValue(orderValue, gotOrderValue) {
		t.Fatalf("rewardPolicy.order mismatch, got %#v", gotOrderValue)
	}
	gotScopeValue := requireRemoteObjectValueField[*remote.ObjectValue](t, gotValue, "scope")
	if !remote.CompareObjectValue(scopeValue, gotScopeValue) {
		t.Fatalf("rewardPolicy.scope mismatch, got %#v", gotScopeValue)
	}
	gotItemValues := requireRemoteObjectValueField[*remote.SliceObjectValue](t, gotValue, "item")
	if !remote.CompareSliceObjectValue(itemSlice, gotItemValues) {
		t.Fatalf("rewardPolicy.item mismatch, got %#v", gotItemValues)
	}

	gotOrder := requireRemoteFieldValue[*remote.ObjectValue](t, model, "order")
	if !remote.CompareObjectValue(orderValue, gotOrder) {
		t.Fatalf("rewardPolicy.order mismatch, got %#v", gotOrder)
	}

	gotScope := requireRemoteFieldValue[*remote.ObjectValue](t, model, "scope")
	if !remote.CompareObjectValue(scopeValue, gotScope) {
		t.Fatalf("rewardPolicy.scope mismatch, got %#v", gotScope)
	}

	gotItems := requireRemoteFieldValue[*remote.SliceObjectValue](t, model, "item")
	if !remote.CompareSliceObjectValue(itemSlice, gotItems) {
		t.Fatalf("rewardPolicy.item mismatch, got %#v", gotItems)
	}
}

func TestRemoteProviderVMIGetEntityFilterRespectsLiteView(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", nil)
	registerAllVMIRemoteObjects(t, remoteProvider)

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(3001)},
			{Name: "name", Value: "banana"},
			{Name: "description", Value: "detail only"},
			{Name: "expire", Value: 7},
			{
				Name: "status",
				Value: &remote.ObjectValue{
					Name:    "status",
					PkgPath: "/vmi",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(3)},
						{Name: "value", Value: 1},
						{Name: "name", Value: "enabled"},
					},
				},
			},
			{Name: "tags", Value: []string{"lite"}},
		},
	}

	filter, err := remoteProvider.GetEntityFilter(productValue, models.LiteView)
	if err != nil {
		t.Fatalf("GetEntityFilter(productValue, LiteView) failed: %v", err)
	}

	maskModel := filter.MaskModel()
	if maskModel == nil {
		t.Fatal("MaskModel() should not return nil")
	}

	if requireRemoteFieldValue[int64](t, maskModel, "id") != int64(3001) {
		t.Fatalf("product.id mismatch, got %#v", maskModel.GetField("id").GetValue().Get())
	}
	if requireRemoteFieldValue[string](t, maskModel, "name") != "banana" {
		t.Fatalf("product.name mismatch, got %#v", maskModel.GetField("name").GetValue().Get())
	}
	if requireRemoteFieldValue[string](t, maskModel, "description") != "detail only" {
		t.Fatalf("product.description mismatch, got %#v", maskModel.GetField("description").GetValue().Get())
	}
	gotTags := requireRemoteFieldValue[[]string](t, maskModel, "tags")
	if len(gotTags) != 1 || gotTags[0] != "lite" {
		t.Fatalf("product.tags mismatch, got %#v", gotTags)
	}

	requireRemoteFieldInvalid(t, maskModel, "expire")
	requireRemoteFieldInvalid(t, maskModel, "status")

	maskValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "name", Value: "masked banana"},
			{Name: "tags", Value: []string{"masked"}},
			{Name: "expire", Value: 99},
			{
				Name: "status",
				Value: &remote.ObjectValue{
					Name:    "status",
					PkgPath: "/vmi",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(33)},
					},
				},
			},
		},
	}
	if err := filter.ValueMask(maskValue); err != nil {
		t.Fatalf("ValueMask(product lite) failed: %v", err)
	}

	maskedByValueModel := filter.MaskModel()
	if maskedByValueModel == nil {
		t.Fatal("MaskModel() with ValueMask should not return nil")
	}
	if requireRemoteFieldValue[string](t, maskedByValueModel, "name") != "masked banana" {
		t.Fatalf("masked product.name mismatch, got %#v", maskedByValueModel.GetField("name").GetValue().Get())
	}
	maskedTags := requireRemoteFieldValue[[]string](t, maskedByValueModel, "tags")
	if len(maskedTags) != 1 || maskedTags[0] != "masked" {
		t.Fatalf("masked product.tags mismatch, got %#v", maskedTags)
	}
	requireRemoteFieldInvalid(t, maskedByValueModel, "expire")
	requireRemoteFieldInvalid(t, maskedByValueModel, "status")
}

func TestRemoteProviderVMISetModelValueRespectsLiteView(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", nil)
	objects := registerAllVMIRemoteObjects(t, remoteProvider)

	productModel := objects["/vmi/product"]
	if productModel == nil {
		t.Fatal("missing registered /vmi/product model")
	}

	liteModel := productModel.Copy(models.LiteView)
	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(4001)},
			{Name: "name", Value: "orange"},
			{Name: "description", Value: "detail field"},
			{Name: "expire", Value: 10},
			{
				Name: "status",
				Value: &remote.ObjectValue{
					Name:    "status",
					PkgPath: "/vmi",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(4)},
						{Name: "value", Value: 2},
						{Name: "name", Value: "disabled"},
					},
				},
			},
			{Name: "tags", Value: []string{"citrus"}},
		},
	}

	entityValue, err := remote.GetEntityValue(productValue)
	if err != nil {
		t.Fatalf("GetEntityValue(productValue) failed: %v", err)
	}

	liteModel, err = remoteProvider.SetModelValue(liteModel, entityValue)
	if err != nil {
		t.Fatalf("SetModelValue(lite product) failed: %v", err)
	}

	if requireRemoteFieldValue[int64](t, liteModel, "id") != int64(4001) {
		t.Fatalf("product.id mismatch, got %#v", liteModel.GetField("id").GetValue().Get())
	}
	if requireRemoteFieldValue[string](t, liteModel, "name") != "orange" {
		t.Fatalf("product.name mismatch, got %#v", liteModel.GetField("name").GetValue().Get())
	}
	if requireRemoteFieldValue[string](t, liteModel, "description") != "detail field" {
		t.Fatalf("product.description mismatch, got %#v", liteModel.GetField("description").GetValue().Get())
	}
	gotTags := requireRemoteFieldValue[[]string](t, liteModel, "tags")
	if len(gotTags) != 1 || gotTags[0] != "citrus" {
		t.Fatalf("product.tags mismatch, got %#v", gotTags)
	}

	requireRemoteFieldInvalid(t, liteModel, "expire")
	requireRemoteFieldInvalid(t, liteModel, "status")
}

func TestRemoteProviderVMIGetTypeFilterRespectsLiteView(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", nil)
	objects := registerAllVMIRemoteObjects(t, remoteProvider)

	stockInObject := objects["/vmi/store/stockin"]
	if stockInObject == nil {
		t.Fatal("missing registered /vmi/store/stockin model")
	}
	goodsInfoField := stockInObject.GetField("goodsInfo")
	if goodsInfoField == nil {
		t.Fatal("/vmi/store/stockin.goodsInfo should exist")
	}

	filter, err := remoteProvider.GetTypeFilter(goodsInfoField.GetType(), models.LiteView)
	if err != nil {
		t.Fatalf("GetTypeFilter(goodsInfo, LiteView) failed: %v", err)
	}
	maskValue := &remote.ObjectValue{
		Name:    "goodsInfo",
		PkgPath: "/vmi/store",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(51)},
			{Name: "sku", Value: "sku-51"},
			{Name: "count", Value: 8},
			{Name: "price", Value: 19.5},
			{
				Name: "product",
				Value: &remote.ObjectValue{
					Name:    "product",
					PkgPath: "/vmi",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(6001)},
					},
				},
			},
			{
				Name: "status",
				Value: &remote.ObjectValue{
					Name:    "status",
					PkgPath: "/vmi",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(71)},
					},
				},
			},
		},
	}
	if err := filter.ValueMask(maskValue); err != nil {
		t.Fatalf("ValueMask(goodsInfo lite) failed: %v", err)
	}

	maskModel := filter.MaskModel()
	if maskModel == nil {
		t.Fatal("MaskModel() should not return nil")
	}
	if requireRemoteFieldValue[int64](t, maskModel, "id") != int64(51) {
		t.Fatalf("goodsInfo.id mismatch, got %#v", maskModel.GetField("id").GetValue().Get())
	}
	if requireRemoteFieldValue[string](t, maskModel, "sku") != "sku-51" {
		t.Fatalf("goodsInfo.sku mismatch, got %#v", maskModel.GetField("sku").GetValue().Get())
	}
	if requireRemoteFieldValue[int](t, maskModel, "count") != 8 {
		t.Fatalf("goodsInfo.count mismatch, got %#v", maskModel.GetField("count").GetValue().Get())
	}
	if requireRemoteFieldValue[float64](t, maskModel, "price") != 19.5 {
		t.Fatalf("goodsInfo.price mismatch, got %#v", maskModel.GetField("price").GetValue().Get())
	}
	productValue := requireRemoteFieldValue[*remote.ObjectValue](t, maskModel, "product")
	if productValue.GetFieldValue("id") != int64(6001) {
		t.Fatalf("goodsInfo.product mismatch, got %#v", productValue)
	}
	requireRemoteFieldInvalid(t, maskModel, "status")
}

func TestRemoteProviderVMIEncodeValueCompressesRelationPrimaryKeys(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", nil)
	objects := registerAllVMIRemoteObjects(t, remoteProvider)

	productObject := objects["/vmi/product"]
	if productObject == nil {
		t.Fatal("missing registered /vmi/product model")
	}
	statusField := productObject.GetField("status")
	skuInfoField := productObject.GetField("skuInfo")
	if statusField == nil || skuInfoField == nil {
		t.Fatal("/vmi/product status or skuInfo field should exist")
	}

	statusValue := &remote.ObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(9)},
			{Name: "value", Value: 2},
			{Name: "name", Value: "published"},
		},
	}
	encodedStatus, err := remoteProvider.EncodeValue(statusValue, statusField.GetType())
	if err != nil {
		t.Fatalf("EncodeValue(status relation) failed: %v", err)
	}
	if encodedStatus != int64(9) {
		t.Fatalf("encoded status should be primary key 9, got %#v", encodedStatus)
	}

	skuInfoValue := &remote.ObjectValue{
		Name:    "skuInfo",
		PkgPath: "/vmi/product",
		Fields: []*remote.FieldValue{
			{Name: "sku", Value: "sku-encoded"},
			{Name: "description", Value: "encoded"},
		},
	}
	encodedSKU, err := remoteProvider.EncodeValue(skuInfoValue, skuInfoField.GetType())
	if err != nil {
		t.Fatalf("EncodeValue(skuInfo relation item) failed: %v", err)
	}
	if encodedSKU != "sku-encoded" {
		t.Fatalf("encoded skuInfo should be primary key sku-encoded, got %#v", encodedSKU)
	}

	statusPrimary := objects["/vmi/status"].GetPrimaryField()
	if statusPrimary == nil {
		t.Fatal("missing /vmi/status primary field")
	}
	decodedStatus, err := remoteProvider.DecodeValue(encodedStatus, statusPrimary.GetType())
	if err != nil {
		t.Fatalf("DecodeValue(status primary) failed: %v", err)
	}
	if decodedStatus != int64(9) {
		t.Fatalf("decoded status primary mismatch, got %#v", decodedStatus)
	}
}
