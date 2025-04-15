package helper

import (
	"testing"
	"time"
)

var idVal = 100
var i8Val = int8(8)
var i16Val = int16(16)
var i32Val = int32(32)
var i64Val = int64(64)
var iVal = 200
var ui8Val = uint8(18)
var ui16Val = uint16(116)
var ui32Val = uint32(132)
var ui64Val = uint64(164)
var uiVal = uint(400)
var strVal = "hello world"
var tsVal = time.Now()
var bVal = true
var iArray = []int{iVal, iVal}
var strArray = []string{strVal, strVal}
var tsArray = []time.Time{tsVal, tsVal}
var bArray = []bool{bVal, bVal}
var iPtr = &iVal
var tsPtr = &tsVal

var baseVal = &Base{
	ID:        idVal,
	I8:        i8Val,
	I16:       i16Val,
	I32:       i32Val,
	I64:       i64Val,
	IVal:      iVal,
	UI8:       ui8Val,
	UI16:      ui16Val,
	UI32:      ui32Val,
	UI64:      ui64Val,
	UIVal:     uiVal,
	Str:       strVal,
	TimeStamp: tsVal,
	Flag:      bVal,
	IArray:    iArray,
	StrArray:  strArray,
	TsArray:   tsArray,
	BArray:    bArray,
	IPtr:      iPtr,
	TSPtr:     tsPtr,
}

var basePtrArrayVal = []*Base{baseVal, baseVal}
var composeVal = &Compose{
	ID:              idVal,
	Name:            strVal,
	Base:            *baseVal,
	BasePtr:         baseVal,
	BaseArray:       []Base{*baseVal, *baseVal},
	BasePtrArray:    basePtrArrayVal,
	BasePtrArrayPtr: &basePtrArrayVal,
}

type Base struct {
	ID        int         `orm:"id key auto"`
	I8        int8        `orm:"i8"`
	I16       int16       `orm:"i16"`
	I32       int32       `orm:"i32"`
	I64       int64       `orm:"i64"`
	IVal      int         `orm:"iVal"`
	UI8       uint8       `orm:"ui8"`
	UI16      uint16      `orm:"ui16"`
	UI32      uint32      `orm:"ui32"`
	UI64      uint64      `orm:"ui64"`
	UIVal     uint        `orm:"uiVal"`
	F32       float32     `orm:"f32"`
	F64       float64     `orm:"f64"`
	Str       string      `orm:"name"`
	TimeStamp time.Time   `orm:"ts"`
	Flag      bool        `orm:"flag"`
	IArray    []int       `orm:"iArray"`
	StrArray  []string    `orm:"strArray"`
	TsArray   []time.Time `orm:"tsArray"`
	BArray    []bool      `orm:"bArray"`
	IPtr      *int        `orm:"iPtr"`
	TSPtr     *time.Time  `orm:"tsPtr"`
	IArrayPtr *[]int      `orm:"iArrayPtr"`
}

type Compose struct {
	ID              int      `orm:"id key auto"`
	Name            string   `orm:"name"`
	Base            Base     `orm:"base"`
	BasePtr         *Base    `orm:"basePtr"`
	BaseArray       []Base   `orm:"baseArray"`
	BasePtrArray    []*Base  `orm:"basePtrArray"`
	BasePtrArrayPtr *[]*Base `orm:"basePtrArrayPtr"`
}

type Person struct {
	ID   int     `orm:"id key auto"`
	Name string  `orm:"name"`
	Age  *int    `orm:"age"`
	Addr *string `orm:"addr"`
}

func TestPerson(t *testing.T) {
	age := 40
	//addr := "test addr"
	person := &Person{
		ID:   12,
		Name: "test",
		Age:  &age,
	}

	personObjectVal, personObjectErr := GetObjectValue(person)
	if personObjectErr != nil {
		t.Errorf("GetObjectValue failed, error:%s", personObjectErr.Error())
		return
	}

	nPerson := &Person{}
	entityErr := UpdateEntity(personObjectVal, nPerson)
	if entityErr != nil {
		t.Errorf("UpdateEntity failed, error:%s", entityErr.Error())
		return
	}
	if nPerson.ID != 12 || nPerson.Name != "test" || nPerson.Age == nil || nPerson.Addr != nil {
		t.Errorf("UpdateEntity failed")
		return
	}
}
