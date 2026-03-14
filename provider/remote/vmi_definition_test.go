package remote

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/muidea/magicOrm/models"
)

func loadVMIObject(t *testing.T, relativePath string) *Object {
	t.Helper()

	filePath := filepath.Join("..", "..", relativePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile(%s) failed: %v", relativePath, err)
	}

	object := &Object{}
	if err := json.Unmarshal(data, object); err != nil {
		t.Fatalf("Unmarshal(%s) failed: %v", relativePath, err)
	}

	return object
}

func requireField(t *testing.T, object *Object, name string) models.Field {
	t.Helper()

	field := object.GetField(name)
	if field == nil {
		t.Fatalf("%s.%s should exist", object.GetName(), name)
	}
	return field
}

func TestVMIEntityDefinitionsDecodeAndVerify(t *testing.T) {
	root := filepath.Join("..", "..", "test", "vmi", "entity")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("ReadFile(%s) failed: %v", path, err)
			return nil
		}

		object := &Object{}
		if err := json.Unmarshal(data, object); err != nil {
			t.Errorf("Unmarshal(%s) failed: %v", path, err)
			return nil
		}

		if err := models.VerifyModel(object); err != nil {
			t.Errorf("VerifyModel(%s) failed: %v", path, err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir failed: %v", err)
	}
}

func TestVMIProductFieldKinds(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/product/product.json")

	skuInfo := requireField(t, object, "skuInfo")
	if !models.IsSliceField(skuInfo) || !models.IsStructType(skuInfo.GetType().Elem().GetValue()) {
		t.Fatalf("product.skuInfo should be slice of struct, got type=%v elem=%v", skuInfo.GetType().GetValue(), skuInfo.GetType().Elem().GetValue())
	}

	image := requireField(t, object, "image")
	if !models.IsSliceField(image) || !models.IsBasicType(image.GetType().Elem().GetValue()) {
		t.Fatalf("product.image should be slice of basic value, got type=%v elem=%v", image.GetType().GetValue(), image.GetType().Elem().GetValue())
	}

	status := requireField(t, object, "status")
	if !models.IsStructField(status) || !models.IsPtrField(status) {
		t.Fatalf("product.status should be pointer struct, got type=%v isPtr=%v", status.GetType().GetValue(), status.GetType().IsPtrType())
	}
}

func TestVMIStockInFieldKinds(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/store/stockin.json")

	goodsInfo := requireField(t, object, "goodsInfo")
	if !models.IsSliceField(goodsInfo) || !models.IsStructType(goodsInfo.GetType().Elem().GetValue()) {
		t.Fatalf("stockin.goodsInfo should be slice of struct, got type=%v elem=%v", goodsInfo.GetType().GetValue(), goodsInfo.GetType().Elem().GetValue())
	}

	store := requireField(t, object, "store")
	if !models.IsStructField(store) || !models.IsPtrField(store) {
		t.Fatalf("stockin.store should be pointer struct, got type=%v isPtr=%v", store.GetType().GetValue(), store.GetType().IsPtrType())
	}

	status := requireField(t, object, "status")
	if !models.IsStructField(status) || !models.IsPtrField(status) {
		t.Fatalf("stockin.status should be pointer struct, got type=%v isPtr=%v", status.GetType().GetValue(), status.GetType().IsPtrType())
	}
}

func TestVMIRewardPolicyFieldKinds(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/bill/rewardPolicy/rewardPolicy.json")

	item := requireField(t, object, "item")
	if !models.IsSliceField(item) || !models.IsStructType(item.GetType().Elem().GetValue()) {
		t.Fatalf("rewardPolicy.item should be slice of struct, got type=%v elem=%v", item.GetType().GetValue(), item.GetType().Elem().GetValue())
	}

	scope := requireField(t, object, "scope")
	if !models.IsStructField(scope) || models.IsPtrField(scope) {
		t.Fatalf("rewardPolicy.scope should be value struct, got type=%v isPtr=%v", scope.GetType().GetValue(), scope.GetType().IsPtrType())
	}

	status := requireField(t, object, "status")
	if !models.IsStructField(status) || !models.IsPtrField(status) {
		t.Fatalf("rewardPolicy.status should be pointer struct, got type=%v isPtr=%v", status.GetType().GetValue(), status.GetType().IsPtrType())
	}
}

func TestVMIStoreFieldKinds(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/store/store.json")

	goods := requireField(t, object, "goods")
	if !models.IsSliceField(goods) || !models.IsStructType(goods.GetType().Elem().GetValue()) || !goods.GetType().Elem().IsPtrType() {
		t.Fatalf("store.goods should be slice of pointer struct, got type=%v elem=%v elemPtr=%v", goods.GetType().GetValue(), goods.GetType().Elem().GetValue(), goods.GetType().Elem().IsPtrType())
	}

	shelf := requireField(t, object, "shelf")
	if !models.IsSliceField(shelf) || !models.IsStructType(shelf.GetType().Elem().GetValue()) || !shelf.GetType().Elem().IsPtrType() {
		t.Fatalf("store.shelf should be slice of pointer struct, got type=%v elem=%v elemPtr=%v", shelf.GetType().GetValue(), shelf.GetType().Elem().GetValue(), shelf.GetType().Elem().IsPtrType())
	}
}
