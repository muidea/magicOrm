package model

import (
	"github.com/muidea/magicCommon/foundation/util"
)

const (
	EqualOpr    = iota // =
	NotEqualOpr        // !=
	BelowOpr           // <
	AboveOpr           // >
	InOpr              // in
	NotInOpr           // not in
	LikeOpr            // like
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
	Equal(key string, val any) error
	NotEqual(key string, val any) error
	Below(key string, val any) error
	Above(key string, val any) error
	In(key string, val any) error
	NotIn(key string, val any) error
	Like(key string, val any) error
	Page(page *util.Pagination)
	Sort(sorter *util.SortFilter)
	ValueMask(val any) error

	GetFilterItem(key string) FilterItem
	Pagination() (limit, offset int, paging bool)
	Sorter() Sorter
	MaskModel() Model
}
