package models

import (
	cd "github.com/muidea/magicCommon/def"
)

const (
	EqualOpr    = iota // =
	NotEqualOpr        // !=
	BelowOpr           // <
	AboveOpr           // >
	InOpr              // in
	NotInOpr           // !in
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

type Paginationer interface {
	Limit() int64
	Offset() int64
}

// Filter orm query filter
type Filter interface {
	GetName() string
	GetPkgPath() string
	Equal(key string, val any) *cd.Error
	NotEqual(key string, val any) *cd.Error
	Below(key string, val any) *cd.Error
	Above(key string, val any) *cd.Error
	In(key string, val any) *cd.Error
	NotIn(key string, val any) *cd.Error
	Like(key string, val any) *cd.Error
	Pagination(pageNum, pageSize int64)
	Sort(fieldName string, ascFlag bool)
	ValueMask(val any) *cd.Error

	GetFilterItem(key string) FilterItem
	Paginationer() Paginationer
	Sorter() Sorter
	MaskModel() Model
}
