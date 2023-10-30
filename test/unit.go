package test

import (
	"time"
)

// Unit 单元信息
type Unit struct {
	ID        int       `orm:"id key auto" view:"view,lite"`
	I8        int8      `orm:"i8" view:"view,lite"`
	I16       int16     `orm:"i16" view:"view,lite"`
	I32       int32     `orm:"i32" view:"view,lite"`
	I64       uint64    `orm:"i64" view:"view,lite"`
	Name      string    `orm:"name" view:"view,lite"`
	Value     float32   `orm:"value" view:"view,lite"`
	F64       float64   `orm:"f64" view:"view,lite"`
	TimeStamp time.Time `orm:"ts" view:"view,lite"`
	Flag      bool      `orm:"flag" view:"view,lite"`
	IArray    []int     `orm:"iArray" view:"view,lite"`
	FArray    []float32 `orm:"fArray" view:"view,lite"`
	StrArray  []string  `orm:"strArray" view:"view,lite"`
}

// ExtUnit ExtUnit
type ExtUnit struct {
	ID   int   `orm:"id key auto" view:"view,lite"`
	Unit *Unit `orm:"unit" view:"view,lite"`
}

// ExtUnitList ExtUnitList
type ExtUnitList struct {
	ID       int    `orm:"id key auto" view:"view,lite"`
	Unit     Unit   `orm:"unit" view:"view,lite"`
	UnitList []Unit `orm:"unitlist" view:"view,lite"`
}
