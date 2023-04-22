package model

import (
	"github.com/muidea/magicCommon/foundation/util"
)

const (
	Equal = iota
	NotEqual
	Below
	Above
	In
	NotIn
	Like
)

type OprCode int

// FilterItem FilterItem
type FilterItem interface {
	OprCode() OprCode
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

	GetFilterItem(key string) FilterItem
	Pagination() (limit, offset int, paging bool)
	Sorter() Sorter
	MaskModel() Model
}
