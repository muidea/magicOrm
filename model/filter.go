package model

import (
	"github.com/muidea/magicCommon/foundation/util"
)

type OprFunc func(string, interface{}) string

// FilterItem FilterItem
type FilterItem interface {
	OprFunc() OprFunc
	OprValue() Value
}

// Sorter sort Item
type Sorter interface {
	Name() string
	AscSort() bool
}

// Filter orm query filter
type Filter interface {
	Equal(key string, val interface{}) error
	NotEqual(key string, val interface{}) error
	Below(key string, val interface{}) error
	Above(key string, val interface{}) error
	In(key string, val interface{}) error
	NotIn(key string, val interface{}) error
	Like(key string, val interface{}) error
	Page(filter *util.Pagination)
	Sort(sorter *util.SortFilter)
	ValueMask(val interface{}) error

	GetFilterItem(name string) FilterItem
	Pagination() (limit, offset int, paging bool)
	Sorter() Sorter
	MaskModel() Model
}
