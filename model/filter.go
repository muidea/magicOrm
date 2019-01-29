package model

import (
	"muidea.com/magicCommon/foundation/util"
)

// FilterItem FilterItem
type FilterItem interface {
	Verify(fType FieldType) (err error)
	FilterStr(name string) (string, error)
}

// Filter orm query filter
type Filter interface {
	Equle(key string, val interface{}) error
	NotEqule(key string, val interface{}) error
	Below(key string, val interface{}) error
	Above(key string, val interface{}) error
	In(key string, val interface{}) error
	NotIn(key string, val interface{}) error
	Like(key string, val interface{}) error
	PageFilter(filter *util.PageFilter)

	Items() map[string]FilterItem
	Pagination() (limit, offset int, paging bool)
}
