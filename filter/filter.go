package filter

import (
	"muidea.com/magicCommon/foundation/util"
	"muidea.com/magicOrm/model"
)

// Filter orm query filter
type Filter interface {
	Equle(key string, val interface{}) error
	NotEqule(key string, val interface{}) error
	Below(key string, val interface{}) error
	Above(key string, val interface{}) error
	In(key string, val interface{}) error
	NotIn(key string, val interface{}) error
	PageFilter(filter *util.PageFilter)
	Builder(structInfo model.StructInfo) (string, error)
}
