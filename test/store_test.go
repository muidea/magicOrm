package test

import (
	"testing"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestLocalStore(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider(localOwner)

	o1, err := orm.NewOrm(localProvider, config, "xyz")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&SKUInfo{}, &Product{}, &Store{}, &GoodsInfo{}, &StockIn{}}
	_, err = registerModel(localProvider, objList)
	if err != nil {
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	skuInfo001 := &SKUInfo{
		SKU:         "sk0001",
		Description: "test sku info",
		Image:       []string{"abc_url", "bcd_url", "cde_url"},
		Namespace:   "xyz",
	}

	skuInfoModel, skuInfoErr := localProvider.GetEntityModel(skuInfo001)
	if skuInfoErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, erro:%s", skuInfoErr.Error())
		return
	}

	err = o1.Drop(skuInfoModel)
	if err != nil {
		t.Errorf("o1.Drop failed. err:%s", err.Error())
		return
	}

	err = o1.Create(skuInfoModel)
	if err != nil {
		t.Errorf("o1.Create failed. err:%s", err.Error())
		return
	}

	skuInfoModel, skuInfoErr = o1.Insert(skuInfoModel)
	if skuInfoErr != nil {
		t.Errorf("o1.Insert failed, erro:%s", skuInfoErr.Error())
		return
	}
	skuInfo001 = skuInfoModel.Interface(true, model.DetailView).(*SKUInfo)

	product001 := &Product{
		Name:        "pro001",
		Description: "test product",
		SKUInfo:     []*SKUInfo{skuInfo001},
		Image:       []string{"abc_url", "bcd_url", "cde_url"},
		Expire:      123,
	}
	productModel, productErr := localProvider.GetEntityModel(product001)
	if productErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, erro:%s", productErr.Error())
		return
	}
	err = o1.Drop(productModel)
	if err != nil {
		t.Errorf("o1.Drop failed. err:%s", err.Error())
		return
	}

	err = o1.Create(productModel)
	if err != nil {
		t.Errorf("o1.Create failed. err:%s", err.Error())
		return
	}

	productModel, productErr = o1.Insert(productModel)
	if productErr != nil {
		t.Errorf("o1.Insert failed, erro:%s", productErr.Error())
		return
	}
	product001 = productModel.Interface(true, model.DetailView).(*Product)

	store001 := &Store{
		Code: "store001",
		Name: "store001",
	}
	storeModel, storeErr := localProvider.GetEntityModel(store001)
	if storeErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, erro:%s", storeErr.Error())
		return
	}
	err = o1.Create(storeModel)
	if err != nil {
		t.Errorf("o1.Create failed. err:%s", err.Error())
		return
	}

	storeModel, storeErr = o1.Insert(storeModel)
	if storeErr != nil {
		t.Errorf("o1.Insert failed, erro:%s", storeErr.Error())
		return
	}
	store001 = storeModel.Interface(true, model.DetailView).(*Store)

	goodsInfo := &GoodsInfo{}
	goodsInfoModel, goodsInfoErr := localProvider.GetEntityModel(goodsInfo)
	if goodsInfoErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, erro:%s", goodsInfoErr.Error())
		return
	}
	err = o1.Create(goodsInfoModel)
	if err != nil {
		t.Errorf("o1.Create failed. err:%s", err.Error())
		return
	}

	stockIn001 := &StockIn{
		SN: "si001",
		GoodsInfo: []GoodsInfo{
			{
				SKU:     "sk0001",
				Product: product001,
				Count:   100,
				Price:   23.45,
			},
		},
		Description: "test stockIn",
		Store:       store001,
	}
	stockInModel, stockInErr := localProvider.GetEntityModel(stockIn001)
	if stockInErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, erro:%s", stockInErr.Error())
		return
	}
	err = o1.Create(stockInModel)
	if err != nil {
		t.Errorf("o1.Create failed. err:%s", err.Error())
		return
	}

	stockInModel, stockInErr = o1.Insert(stockInModel)
	if stockInErr != nil {
		t.Errorf("o1.Insert failed, erro:%s", stockInErr.Error())
		return
	}
	_ = stockInModel.Interface(true, model.DetailView).(*StockIn)
}
