package test

type SKUInfo struct {
	SKU         string   `json:"sku" orm:"sku key" view:"view,lite"`
	Description string   `json:"description" orm:"description" view:"view,lite"`
	Image       []string `json:"image" orm:"image" view:"view,lite"`
	Namespace   string   `json:"namespace" orm:"namespace" view:"view,lite"`
}

type Product struct {
	ID          int64      `json:"id" orm:"id key auto" view:"view,lite"`
	Name        string     `json:"name" orm:"name" view:"view,lite"`
	Description string     `json:"description" orm:"description" view:"view,lite"`
	SKUInfo     []*SKUInfo `json:"skuInfo" orm:"skuInfo" view:"view,lite"`
	Image       []string   `json:"image" orm:"image" view:"view,lite"`
	Expire      int        `json:"expire" orm:"expire" view:"view,lite"`
}

type Store struct {
	ID   int64  `json:"id" orm:"id key auto" view:"view,lite"`
	Code string `json:"code" orm:"code" view:"view,lite"`
	Name string `json:"name" view:"view,lite"`
}

type GoodsInfo struct {
	ID      int64    `json:"id" orm:"id key auto" view:"view,lite"`
	SKU     string   `json:"sku" orm:"sku" view:"view,lite"`
	Product *Product `json:"product" orm:"product" view:"view,lite"`
	Count   int      `json:"count" orm:"count" view:"view,lite"`
	Price   float64  `json:"price" orm:"price" view:"view,lite"`
}

type StockIn struct {
	ID          int64       `json:"id" orm:"id key auto" view:"view,lite"`
	SN          string      `json:"sn" orm:"sn" view:"view,lite"`
	GoodsInfo   []GoodsInfo `json:"goodsInfo" view:"view,lite"`
	Description string      `json:"description" orm:"description" view:"view,lite"`
	Store       *Store      `json:"store" orm:"store" view:"view,lite"`
}
