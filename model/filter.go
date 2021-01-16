package model

import (
	"github.com/muidea/magicCommon/foundation/util"
)

// FilterItem FilterItem
type FilterItem interface {
	FilterStr(name string, fType Type) (string, error)
}

// Sorter sort Item
type Sorter interface {
	Name() string
	SortStr(tagName string) string
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
	ValueMask(val interface{}) error
	Page(filter *util.PageFilter)
	Sort(sorter *util.SortFilter)

	GetFilterItem(name string) FilterItem
	Pagination() (limit, offset int, paging bool)
	MaskModel() Model
	Sorter() Sorter
}
