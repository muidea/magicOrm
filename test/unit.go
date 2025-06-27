package test

import (
	"time"
)

// Unit 单元信息
type Unit struct {
	ID        int       `orm:"id key auto" view:"detail,lite"`
	I8        int8      `orm:"i8" view:"detail,lite"`
	I16       int16     `orm:"i16" view:"detail,lite"`
	I32       int32     `orm:"i32" view:"detail,lite"`
	I64       uint64    `orm:"i64" view:"detail,lite"`
	Name      string    `orm:"name" view:"detail,lite"`
	Value     float32   `orm:"value" view:"detail,lite"`
	F64       float64   `orm:"f64" view:"detail,lite"`
	TimeStamp time.Time `orm:"ts" view:"detail,lite"`
	Flag      bool      `orm:"flag" view:"detail,lite"`
	IArray    []int     `orm:"iArray" view:"detail,lite"`
	FArray    []float32 `orm:"fArray" view:"detail,lite"`
	StrArray  []string  `orm:"strArray" view:"detail,lite"`
}

// ExtUnit ExtUnit
type ExtUnit struct {
	ID   int   `orm:"id key auto" view:"detail,lite"`
	Unit *Unit `orm:"unit" view:"detail,lite"`
}

// ExtUnitList ExtUnitList
type ExtUnitList struct {
	ID       int    `orm:"id key auto" view:"detail,lite"`
	Unit     Unit   `orm:"unit" view:"detail,lite"`
	UnitList []Unit `orm:"unitlist" view:"detail,lite"`
}
