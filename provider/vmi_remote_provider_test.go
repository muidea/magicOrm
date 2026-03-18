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
		{objectKey: "/vmi/product/productInfo", fieldName: "product", wantKey: "/vmi/product"},
		{objectKey: "/vmi/store/goodsInfo", fieldName: "product", wantKey: "/vmi/product/productInfo"},
		{objectKey: "/vmi/store/goodsInfo", fieldName: "shelf", wantKey: "/vmi/warehouse/shelf"},
		{objectKey: "/vmi/order", fieldName: "goods", wantKey: "/vmi/order/goodsItem"},
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

func TestRemoteProviderVMIGoodsInfoRoundTripAndFilter(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", nil)
	registerAllVMIRemoteObjects(t, remoteProvider)

	productValue := &remote.ObjectValue{
		Name:    "productInfo",
		PkgPath: "/vmi/product",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(81)},
			{Name: "sku", Value: "sku-info-001"},
			{Name: "description", Value: "default product info"},
		},
	}
	shelfSlice := &remote.SliceObjectValue{
		Name:    "shelf",
		PkgPath: "/vmi/warehouse",
		Values: []*remote.ObjectValue{
			{
				Name:    "shelf",
				PkgPath: "/vmi/warehouse",
				Fields: []*remote.FieldValue{
					{Name: "id", Value: int64(301)},
				},
			},
			{
				Name:    "shelf",
				PkgPath: "/vmi/warehouse",
				Fields: []*remote.FieldValue{
					{Name: "id", Value: int64(302)},
				},
			},
		},
	}
	goodsInfoValue := &remote.ObjectValue{
		Name:    "goodsInfo",
		PkgPath: "/vmi/store",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
			{Name: "sku", Value: "sku-001"},
			{Name: "product", Value: productValue},
			{Name: "type", Value: 1},
			{Name: "count", Value: 8},
			{Name: "price", Value: 19.5},
			{Name: "shelf", Value: shelfSlice},
		},
	}

	model, err := remoteProvider.GetEntityModel(goodsInfoValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(goodsInfoValue) failed: %v", err)
	}

	gotValue, ok := model.Interface(true).(*remote.ObjectValue)
	if !ok {
		t.Fatalf("Interface(true) should return *remote.ObjectValue, got %T", model.Interface(true))
	}
	if gotValue.ID != "1001" {
		t.Fatalf("goodsInfo ID mismatch, got %s", gotValue.ID)
	}
	if requireRemoteObjectValueField[int64](t, gotValue, "id") != int64(1001) {
		t.Fatalf("goodsInfo.id mismatch, got %#v", gotValue.GetFieldValue("id"))
	}
	if requireRemoteObjectValueField[string](t, gotValue, "sku") != "sku-001" {
		t.Fatalf("goodsInfo.sku mismatch, got %#v", gotValue.GetFieldValue("sku"))
	}
	if requireRemoteObjectValueField[int](t, gotValue, "type") != 1 {
		t.Fatalf("goodsInfo.type mismatch, got %#v", gotValue.GetFieldValue("type"))
	}
	if requireRemoteObjectValueField[int](t, gotValue, "count") != 8 {
		t.Fatalf("goodsInfo.count mismatch, got %#v", gotValue.GetFieldValue("count"))
	}
	if requireRemoteObjectValueField[float64](t, gotValue, "price") != 19.5 {
		t.Fatalf("goodsInfo.price mismatch, got %#v", gotValue.GetFieldValue("price"))
	}
	gotProductValue := requireRemoteObjectValueField[*remote.ObjectValue](t, gotValue, "product")
	if !remote.CompareObjectValue(productValue, gotProductValue) {
		t.Fatalf("goodsInfo.product mismatch, got %#v", gotProductValue)
	}
	gotShelfValue := requireRemoteObjectValueField[*remote.SliceObjectValue](t, gotValue, "shelf")
	if !remote.CompareSliceObjectValue(shelfSlice, gotShelfValue) {
		t.Fatalf("goodsInfo.shelf mismatch, got %#v", gotShelfValue)
	}

	filter, err := remoteProvider.GetModelFilter(model)
	if err != nil {
		t.Fatalf("GetModelFilter(goodsInfoModel) failed: %v", err)
	}

	if err := filter.Equal("product", productValue); err != nil {
		t.Fatalf("filter.Equal(product) failed: %v", err)
	}
	if err := filter.In("shelf", shelfSlice); err != nil {
		t.Fatalf("filter.In(shelf) failed: %v", err)
	}

	productItem := filter.GetFilterItem("product")
	if productItem == nil || productItem.OprCode() != models.EqualOpr {
		t.Fatalf("product filter item mismatch: %#v", productItem)
	}
	gotProduct, ok := productItem.OprValue().Get().(*remote.ObjectValue)
	if !ok || !remote.CompareObjectValue(productValue, gotProduct) {
		t.Fatalf("product filter value mismatch, got %#v", productItem.OprValue().Get())
	}

	shelfItem := filter.GetFilterItem("shelf")
	if shelfItem == nil || shelfItem.OprCode() != models.InOpr {
		t.Fatalf("shelf filter item mismatch: %#v", shelfItem)
	}
	gotShelf, ok := shelfItem.OprValue().Get().(*remote.SliceObjectValue)
	if !ok || !remote.CompareSliceObjectValue(shelfSlice, gotShelf) {
		t.Fatalf("shelf filter value mismatch, got %#v", shelfItem.OprValue().Get())
	}

	maskValue := &remote.ObjectValue{
		Name:    "goodsInfo",
		PkgPath: "/vmi/store",
		Fields: []*remote.FieldValue{
			{
				Name: "product",
				Value: &remote.ObjectValue{
					Name:    "productInfo",
					PkgPath: "/vmi/product",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(81)},
					},
				},
			},
			{
				Name: "shelf",
				Value: &remote.SliceObjectValue{
					Name:    "shelf",
					PkgPath: "/vmi/warehouse",
					Values:  []*remote.ObjectValue{},
				},
			},
		},
	}
	if err := filter.ValueMask(maskValue); err != nil {
		t.Fatalf("filter.ValueMask(goodsInfo mask) failed: %v", err)
	}

	maskedModel := filter.MaskModel()
	if maskedModel == nil {
		t.Fatal("MaskModel() should not return nil")
	}

	maskedProduct := requireRemoteFieldValue[*remote.ObjectValue](t, maskedModel, "product")
	if maskedProduct.GetFieldValue("id") != int64(81) {
		t.Fatalf("masked product should keep explicit id, got %#v", maskedProduct)
	}

	maskedShelf := requireRemoteFieldValue[*remote.SliceObjectValue](t, maskedModel, "shelf")
	if maskedShelf.Values == nil || len(maskedShelf.Values) != 0 {
		t.Fatalf("masked shelf should be assigned empty slice, got %#v", maskedShelf)
	}

	originalShelf := requireRemoteFieldValue[*remote.SliceObjectValue](t, model, "shelf")
	if len(originalShelf.Values) != 2 {
		t.Fatalf("MaskModel should not mutate original shelf, got %#v", originalShelf)
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
	rewardPolicyValue := &remote.ObjectValue{
		Name:    "rewardPolicy",
		PkgPath: "/vmi/credit",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(2001)},
			{Name: "name", Value: "promotion"},
			{Name: "description", Value: "credit reward"},
			{Name: "policy", Value: "order.total * 2"},
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
	if requireRemoteObjectValueField[string](t, gotValue, "description") != "credit reward" {
		t.Fatalf("rewardPolicy.description mismatch, got %#v", gotValue.GetFieldValue("description"))
	}
	if requireRemoteObjectValueField[string](t, gotValue, "policy") != "order.total * 2" {
		t.Fatalf("rewardPolicy.policy mismatch, got %#v", gotValue.GetFieldValue("policy"))
	}
	gotStatusValue := requireRemoteObjectValueField[*remote.ObjectValue](t, gotValue, "status")
	if !remote.CompareObjectValue(statusValue, gotStatusValue) {
		t.Fatalf("rewardPolicy.status mismatch, got %#v", gotStatusValue)
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
				Name: "shelf",
				Value: &remote.ObjectValue{
					Name:    "shelf",
					PkgPath: "/vmi/warehouse",
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
	requireRemoteFieldInvalid(t, maskModel, "shelf")
}

func TestRemoteProviderVMIEncodeValueCompressesRelationPrimaryKeys(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", nil)
	objects := registerAllVMIRemoteObjects(t, remoteProvider)

	goodsInfoObject := objects["/vmi/store/goodsInfo"]
	if goodsInfoObject == nil {
		t.Fatal("missing registered /vmi/store/goodsInfo model")
	}
	productField := goodsInfoObject.GetField("product")
	shelfField := goodsInfoObject.GetField("shelf")
	if productField == nil || shelfField == nil {
		t.Fatal("/vmi/store/goodsInfo product or shelf field should exist")
	}

	productValue := &remote.ObjectValue{
		Name:    "productInfo",
		PkgPath: "/vmi/product",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(81)},
			{Name: "sku", Value: "sku-encoded"},
		},
	}
	encodedProduct, err := remoteProvider.EncodeValue(productValue, productField.GetType())
	if err != nil {
		t.Fatalf("EncodeValue(product relation) failed: %v", err)
	}
	if encodedProduct != int64(81) {
		t.Fatalf("encoded product should be primary key 81, got %#v", encodedProduct)
	}

	shelfValue := &remote.ObjectValue{
		Name:    "shelf",
		PkgPath: "/vmi/warehouse",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(301)},
			{Name: "capacity", Value: 20},
		},
	}
	encodedShelf, err := remoteProvider.EncodeValue(shelfValue, shelfField.GetType())
	if err != nil {
		t.Fatalf("EncodeValue(shelf relation item) failed: %v", err)
	}
	if encodedShelf != int64(301) {
		t.Fatalf("encoded shelf should be primary key 301, got %#v", encodedShelf)
	}

	productPrimary := objects["/vmi/product/productInfo"].GetPrimaryField()
	if productPrimary == nil {
		t.Fatal("missing /vmi/product/productInfo primary field")
	}
	decodedProduct, err := remoteProvider.DecodeValue(encodedProduct, productPrimary.GetType())
	if err != nil {
		t.Fatalf("DecodeValue(product primary) failed: %v", err)
	}
	if decodedProduct != int64(81) {
		t.Fatalf("decoded product primary mismatch, got %#v", decodedProduct)
	}
}
