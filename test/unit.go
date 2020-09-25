package test

import (
	"time"
)

// Unit 单元信息
type Unit struct {
	ID        int       `orm:"id key auto"`
	I8        int8      `orm:"i8"`
	I16       int16     `orm:"i16"`
	I32       int32     `orm:"i32"`
	I64       uint64    `orm:"i64"`
	Name      string    `orm:"name"`
	Value     float32   `orm:"value"`
	F64       float64   `orm:"f64"`
	TimeStamp time.Time `orm:"ts"`
	Flag      bool      `orm:"flag"`
	IArray    []int     `orm:"iArray"`
	FArray    []float32 `orm:"fArray"`
	StrArray  []string  `orm:"strArray"`
}

// ExtUnit ExtUnit
type ExtUnit struct {
	ID   int   `orm:"id key auto"`
	Unit *Unit `orm:"unit"`
}

// ExtUnitList ExtUnitList
type ExtUnitList struct {
	ID       int    `orm:"id key auto"`
	Unit     Unit   `orm:"unit"`
	UnitList []Unit `orm:"unitlist"`
}
