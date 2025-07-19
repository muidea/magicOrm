package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestRemoteStore(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider("remote")

	o1, err := orm.NewOrm(remoteProvider, config, "remote")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&SKUInfo{}, &Product{}, &Store{}, &GoodsInfo{}, &StockIn{}}
	mList, mErr := registerRemoteModel(remoteProvider, objList)
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
	productValue, productErr := getRemoteProduct(product001)
	if productErr != nil {
		t.Errorf("getRemoteProduct failed, err:%s", productErr.Error())
		return
	}
	productModel, productErr := remoteProvider.GetEntityModel(productValue)
	if productErr != nil {
		t.Errorf("remoteProvider.GetEntityModel failed, erro:%s", productErr.Error())
		return
	}

	productModel, productErr = o1.Insert(productModel)
	if productErr != nil {
		t.Errorf("o1.Insert failed, erro:%s", productErr.Error())
		return
	}
	productValue = productModel.Interface(true).(*remote.ObjectValue)
	curProduct := &Product{}
	err = helper.UpdateEntity(productValue, curProduct)
	if err != nil {
		t.Errorf("helper.UpdateEntity failed, err:%s", err.Error())
		return
	}
	if curProduct.Name != product001.Name || curProduct.Expire != product001.Expire {
		t.Errorf("product failed")
		return
	}

	store001 := getLocalStore()
	storeValue, storeErr := getRemoteStore(store001)
	if storeErr != nil {
		t.Errorf("getRemoteStore failed, err:%s", storeErr.Error())
		return
	}
	storeModel, storeErr := remoteProvider.GetEntityModel(storeValue)
	if storeErr != nil {
		t.Errorf("remoteProvider.GetEntityModel failed, erro:%s", storeErr.Error())
		return
	}
	storeModel, storeErr = o1.Insert(storeModel)
	if storeErr != nil {
		t.Errorf("o1.Insert failed, erro:%s", storeErr.Error())
		return
	}
	storeValue = storeModel.Interface(true).(*remote.ObjectValue)
	curStore := &Store{}
	err = helper.UpdateEntity(storeValue, curStore)
	if err != nil {
		t.Errorf("helper.UpdateEntity failed, err:%s", err.Error())
		return
	}
	if curStore.Name != store001.Name || curStore.Code != store001.Code {
		t.Errorf("store failed")
		return
	}

	stockIn001 := getLocalStockIn(curProduct, curStore)
	stockInValue, stockInErr := getRemoteStockIn(stockIn001)
	if stockInErr != nil {
		t.Errorf("getRemoteStockIn failed, err:%s", stockInErr.Error())
		return
	}
	stockInModel, stockInErr := remoteProvider.GetEntityModel(stockInValue)
	if stockInErr != nil {
		t.Errorf("remoteProvider.GetEntityModel failed, erro:%s", stockInErr.Error())
		return
	}
	stockInModel, stockInErr = o1.Insert(stockInModel)
	if stockInErr != nil {
		t.Errorf("o1.Insert failed, erro:%s", stockInErr.Error())
		return
	}
	stockInValue = stockInModel.Interface(true).(*remote.ObjectValue)
	curStockIn := &StockIn{GoodsInfo: []GoodsInfo{}, Store: &Store{}}
	err = helper.UpdateEntity(stockInValue, curStockIn)
	if err != nil {
		t.Errorf("helper.UpdateEntity failed, err:%s", err.Error())
		return
	}
	if curStockIn.SN != stockIn001.SN ||
		len(curStockIn.GoodsInfo) != len(stockIn001.GoodsInfo) ||
		curStockIn.Store == nil {
		t.Errorf("stockIn failed")
		return
	}

	queryByIDStockIn := &StockIn{
		ID:    curStockIn.ID,
		Store: &Store{},
	}

	queryStockInValue, queryStockInErr := getRemoteStockIn(queryByIDStockIn)
	if queryStockInErr != nil {
		t.Errorf("getRemoteStockIn failed, err:%s", queryStockInErr.Error())
		return
	}
	queryByIDStockInModel, queryByIDStockInErr := remoteProvider.GetEntityModel(queryStockInValue)
	if queryByIDStockInErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, erro:%s", queryByIDStockInErr.Error())
		return
	}
	queryByIDStockInModel, queryByIDStockInErr = o1.Query(queryByIDStockInModel)
	if queryByIDStockInErr != nil {
		t.Errorf("o1.Query failed, erro:%s", queryByIDStockInErr.Error())
		return
	}
	queryStockInValue = queryByIDStockInModel.Interface(true).(*remote.ObjectValue)
	curQueryStockIn := &StockIn{}
	err = helper.UpdateEntity(queryStockInValue, curQueryStockIn)
	if err != nil {
		t.Errorf("helper.UpdateEntity failed, err:%s", err.Error())
		return
	}
	if curQueryStockIn.ID != curStockIn.ID ||
		curQueryStockIn.SN != curStockIn.SN ||
		len(curQueryStockIn.GoodsInfo) != len(curStockIn.GoodsInfo) ||
		curQueryStockIn.Store == nil {
		t.Errorf("queryByIDStockIn failed")
		return
	}
}
