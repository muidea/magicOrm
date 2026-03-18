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

	image := requireField(t, object, "image")
	if !models.IsSliceField(image) || !models.IsBasicType(image.GetType().Elem().GetValue()) {
		t.Fatalf("product.image should be slice of basic value, got type=%v elem=%v", image.GetType().GetValue(), image.GetType().Elem().GetValue())
	}

	tags := requireField(t, object, "tags")
	if !models.IsSliceField(tags) || !models.IsBasicType(tags.GetType().Elem().GetValue()) {
		t.Fatalf("product.tags should be slice of basic value, got type=%v elem=%v", tags.GetType().GetValue(), tags.GetType().Elem().GetValue())
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

func TestVMIOrderFieldKinds(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/order/order.json")

	goods := requireField(t, object, "goods")
	if !models.IsSliceField(goods) || !models.IsStructType(goods.GetType().Elem().GetValue()) {
		t.Fatalf("order.goods should be slice of struct, got type=%v elem=%v", goods.GetType().GetValue(), goods.GetType().Elem().GetValue())
	}

	customer := requireField(t, object, "customer")
	if !models.IsStructField(customer) || !models.IsPtrField(customer) {
		t.Fatalf("order.customer should be pointer struct, got type=%v isPtr=%v", customer.GetType().GetValue(), customer.GetType().IsPtrType())
	}

	store := requireField(t, object, "store")
	if !models.IsStructField(store) || !models.IsPtrField(store) {
		t.Fatalf("order.store should be pointer struct, got type=%v isPtr=%v", store.GetType().GetValue(), store.GetType().IsPtrType())
	}

	status := requireField(t, object, "status")
	if !models.IsStructField(status) || !models.IsPtrField(status) {
		t.Fatalf("order.status should be pointer struct, got type=%v isPtr=%v", status.GetType().GetValue(), status.GetType().IsPtrType())
	}
}

func TestVMIStoreFieldKinds(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/store/store.json")

	shelf := requireField(t, object, "shelf")
	if !models.IsSliceField(shelf) || !models.IsStructType(shelf.GetType().Elem().GetValue()) || !shelf.GetType().Elem().IsPtrType() {
		t.Fatalf("store.shelf should be slice of pointer struct, got type=%v elem=%v elemPtr=%v", shelf.GetType().GetValue(), shelf.GetType().Elem().GetValue(), shelf.GetType().Elem().IsPtrType())
	}
}

func TestVMIRewardPolicyFieldKinds(t *testing.T) {
	object := loadVMIObject(t, "test/vmi/entity/credit/rewardPolicy/rewardPolicy.json")

	policy := requireField(t, object, "policy")
	if !models.IsBasicField(policy) {
		t.Fatalf("rewardPolicy.policy should be basic field, got type=%v", policy.GetType().GetValue())
	}

	status := requireField(t, object, "status")
	if !models.IsStructField(status) || !models.IsPtrField(status) {
		t.Fatalf("rewardPolicy.status should be pointer struct, got type=%v isPtr=%v", status.GetType().GetValue(), status.GetType().IsPtrType())
	}
}
