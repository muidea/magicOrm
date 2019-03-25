package test

import "time"

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID  int    `json:"id" orm:"id key auto"`
	I8  int8   `orm:"i8"`
	I16 int16  `orm:"i16"`
	I32 int32  `orm:"i32"`
	I64 uint64 `orm:"i64"`
	// Name 名称
	Name      string    `json:"name" orm:"name"`
	Value     float32   `json:"value" orm:"value"`
	F64       float64   `orm:"f64"`
	TimeStamp time.Time `json:"timeStamp" orm:"ts"`
	Flag      bool      `orm:"flag"`
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
