//go:build local || all
// +build local all

package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestLocalStore(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider("localOwner")

	o1, err := orm.NewOrm(localProvider, config, "xyz")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&SKUInfo{}, &Product{}, &Store{}, &GoodsInfo{}, &StockIn{}}
	mList, mErr := registerLocalModel(localProvider, objList)
	if mErr != nil {
		t.Errorf("register model failed. err:%s", mErr.Error())
		return
	}

	err = dropModel(o1, mList)
	if err != nil {
		t.Errorf("dropModel failed. err:%s", err.Error())
		return
	}
	err = createModel(o1, mList)
	if err != nil {
		t.Errorf("createModel failed. err:%s", err.Error())
		return
	}

	product001 := getLocalProduct()
	productModel, productErr := localProvider.GetEntityModel(product001)
	if productErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, erro:%s", productErr.Error())
		return
	}

	productModel, productErr = o1.Insert(productModel)
	if productErr != nil {
		t.Errorf("o1.Insert failed, erro:%s", productErr.Error())
		return
	}
	product001 = productModel.Interface(true).(*Product)

	store001 := getLocalStore()
	storeModel, storeErr := localProvider.GetEntityModel(store001)
	if storeErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, erro:%s", storeErr.Error())
		return
	}
	storeModel, storeErr = o1.Insert(storeModel)
	if storeErr != nil {
		t.Errorf("o1.Insert failed, erro:%s", storeErr.Error())
		return
	}
	store001 = storeModel.Interface(true).(*Store)
	stockIn001 := getLocalStockIn(product001, store001)
	stockInModel, stockInErr := localProvider.GetEntityModel(stockIn001)
	if stockInErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, erro:%s", stockInErr.Error())
		return
	}
	stockInModel, stockInErr = o1.Insert(stockInModel)
	if stockInErr != nil {
		t.Errorf("o1.Insert failed, erro:%s", stockInErr.Error())
		return
	}
	stockIn001 = stockInModel.Interface(true).(*StockIn)

	queryByIDStockIn := &StockIn{
		ID:    stockIn001.ID,
		Store: &Store{},
	}

	queryByIDStockInModel, queryByIDStockInErr := localProvider.GetEntityModel(queryByIDStockIn)
	if queryByIDStockInErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, erro:%s", queryByIDStockInErr.Error())
		return
	}
	queryByIDStockInModel, queryByIDStockInErr = o1.Query(queryByIDStockInModel)
	if queryByIDStockInErr != nil {
		t.Errorf("o1.Query failed, erro:%s", queryByIDStockInErr.Error())
		return
	}
	queryByIDStockIn = queryByIDStockInModel.Interface(true).(*StockIn)
	if queryByIDStockIn.ID != stockIn001.ID || queryByIDStockIn.SN != stockIn001.SN || len(queryByIDStockIn.GoodsInfo) != len(stockIn001.GoodsInfo) || queryByIDStockIn.Store == nil {
		t.Errorf("queryByIDStockIn failed")
		return
	}
}
