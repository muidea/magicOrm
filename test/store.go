package test

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/provider/remote"
)

type SKUInfo struct {
	SKU         string   `json:"sku" orm:"sku key" view:"detail,lite"`
	Description string   `json:"description" orm:"description" view:"detail,lite"`
	Image       []string `json:"image" orm:"image" view:"detail,lite"`
	Namespace   string   `json:"namespace" orm:"namespace" view:"detail,lite"`
}

type Product struct {
	ID          int64     `json:"id" orm:"id key auto" view:"detail,lite"`
	Name        string    `json:"name" orm:"name" view:"detail,lite"`
	Description string    `json:"description" orm:"description" view:"detail,lite"`
	SKUInfo     []SKUInfo `json:"skuInfo" orm:"skuInfo" view:"detail,lite"`
	Image       []string  `json:"image" orm:"image" view:"detail,lite"`
	Expire      int       `json:"expire" orm:"expire" view:"detail,lite"`
}

type Store struct {
	ID   int64  `json:"id" orm:"id key auto" view:"detail,lite"`
	Code string `json:"code" orm:"code" view:"detail,lite"`
	Name string `json:"name" view:"detail,lite"`
}

type GoodsInfo struct {
	ID      int64    `json:"id" orm:"id key auto" view:"detail,lite"`
	SKU     string   `json:"sku" orm:"sku" view:"detail,lite"`
	Product *Product `json:"product" orm:"product" view:"detail,lite"`
	Count   int      `json:"count" orm:"count" view:"detail,lite"`
	Price   float64  `json:"price" orm:"price" view:"detail,lite"`
}

type StockIn struct {
	ID          int64       `json:"id" orm:"id key auto" view:"detail,lite"`
	SN          string      `json:"sn" orm:"sn" view:"detail,lite"`
	GoodsInfo   []GoodsInfo `json:"goodsInfo" view:"detail,lite"`
	Description string      `json:"description" orm:"description" view:"detail,lite"`
	Store       *Store      `json:"store" orm:"store" view:"detail,lite"`
}

func getLocalProduct() *Product {
	return &Product{
		Name:        "product",
		Description: "product description",
		SKUInfo: []SKUInfo{
			{
				SKU:         "sku001",
				Description: "sku001 description",
				Image:       []string{"image"},
				Namespace:   "xyz",
			},
			{
				SKU:         "sku002",
				Description: "sku002 description",
				Image:       []string{"image"},
				Namespace:   "xyz",
			},
		},
		Image:  []string{"image"},
		Expire: 1,
	}
}

func getLocalStore() *Store {
	return &Store{
		Code: "store001",
		Name: "store001",
	}
}

func getLocalStockIn(productPtr *Product, storePtr *Store) *StockIn {
	return &StockIn{
		SN:          "stockin001",
		GoodsInfo:   []GoodsInfo{{SKU: "sku001", Product: productPtr, Count: 1, Price: 1}},
		Description: "stockin001 description",
		Store:       storePtr,
	}
}

func getRemoteProduct(productPtr *Product) (ret *remote.ObjectValue, err *cd.Error) {
	ret, err = getObjectValue(productPtr)
	return
}

func getRemoteStore(storePtr *Store) (ret *remote.ObjectValue, err *cd.Error) {
	ret, err = getObjectValue(storePtr)
	return
}

func getRemoteStockIn(stockInPtr *StockIn) (ret *remote.ObjectValue, err *cd.Error) {
	ret, err = getObjectValue(stockInPtr)
	return
}
