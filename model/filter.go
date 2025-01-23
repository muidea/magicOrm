package model

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
	Equal(key string, val any) *cd.Result
	NotEqual(key string, val any) *cd.Result
	Below(key string, val any) *cd.Result
	Above(key string, val any) *cd.Result
	In(key string, val any) *cd.Result
	NotIn(key string, val any) *cd.Result
	Like(key string, val any) *cd.Result
	Pagination(pageNum, pageSize int)
	Sort(fieldName string, ascFlag bool)
	ValueMask(val any) *cd.Result

	GetFilterItem(key string) FilterItem
	Paginationer() Paginationer
	Sorter() Sorter
	MaskModel() Model
}
