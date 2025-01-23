package helper

import (
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
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
var f32Val = float32(12.34)
var f64Val = 23.45
var strVal = "hello world"
var tsVal = time.Now()
var bVal = true
var iArray = []int{iVal, iVal}
var strArray = []string{strVal, strVal}
var tsArray = []time.Time{tsVal, tsVal}
var bArray = []bool{bVal, bVal}
var iPtr = &iVal
var tsPtr = &tsVal
var iArrayPtr = &iArray
var iPtrArray = []*int{iPtr, iPtr}
var iPtrArrayPtr = &iPtrArray

var emptyBase = &Base{}
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

var emptyCompose = &Compose{}
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

func TestModel(t *testing.T) {
	base1 := emptyBase
	err := testValue(t, base1)
	if err != nil {
		t.Error("test base1 failed")
		return
	}

	base2 := baseVal
	err = testValue(t, base2)
	if err != nil {
		t.Error("test base2 failed")
		return
	}

	compose1 := emptyCompose
	err = testValue(t, compose1)
	if err != nil {
		t.Error("test compose1 failed")
		return
	}

	compose2 := composeVal
	err = testValue(t, compose2)
	if err != nil {
		t.Error("test compose2 failed")
		return
	}
}

func testValue(t *testing.T, valPtr any) *cd.Result {
	lModel, lErr := local.GetEntityModel(valPtr)
	if lErr != nil {
		t.Errorf("local.GetEntityModel failed. err:%s", lErr.Error())
		return lErr
	}

	baseObject, baseErr := remote.GetObject(valPtr)
	if baseErr != nil {
		t.Errorf("remote.GetObject failed, err:%s", baseErr.Error())
		return baseErr
	}
	rModel, rErr := remote.GetEntityModel(baseObject)
	if rErr != nil {
		t.Errorf("remote.GetEntityModel failed. err:%s", rErr.Error())
		return rErr
	}

	baseObjectVal, baseValErr := remote.GetObjectValue(valPtr)
	if baseValErr != nil {
		t.Errorf("remote.GetObjectValue failed, err:%s", baseValErr.Error())
		return baseValErr
	}

	byteVal, byteErr := remote.EncodeObjectValue(baseObjectVal)
	if byteErr != nil {
		t.Errorf("remote.EncodeObjectValue failed, err:%v", byteErr.Error())
		return byteErr
	}

	rawObjectVal, rawObjectErr := remote.DecodeObjectValue(byteVal)
	if rawObjectErr != nil {
		t.Errorf("remote.DecodeObjectValue failed, err:%v", rawObjectErr.Error())
		return rawObjectErr
	}

	rVal, rErr := remote.GetEntityValue(rawObjectVal)
	if rErr != nil {
		t.Errorf("remote.GetEntityValue failed. err:%s", rErr.Error())
		return rErr
	}

	rModel, rErr = remote.SetModelValue(rModel, rVal)
	if rErr != nil {
		t.Errorf("remote.SetModelValue failed. err:%s", rErr.Error())
		return rErr
	}

	if !model.CompareModel(lModel, rModel) {
		t.Errorf("CompareModel failed")
		return cd.NewResult(cd.UnExpected, "compare model failed")
	}
	return nil
}

func TestPerson(t *testing.T) {
	age := 40
	addr := "test addr"
	person := &Person{
		ID:   12,
		Name: "test",
		Age:  &age,
		Addr: &addr,
	}

	personObjectVal, personObjectErr := GetObjectValue(person)
	if personObjectErr != nil {
		t.Errorf("GetObjectValue failed, error:%s", personObjectErr.Error())
		return
	}

	personObjectVal.SetFieldValue("age", 32)
	personObjectVal.SetFieldValue("addr", "hey boy!")

	nPerson := &Person{}
	entityErr := UpdateEntity(personObjectVal, nPerson)
	if entityErr != nil {
		t.Errorf("UpdateEntity failed, error:%s", entityErr.Error())
		return
	}
	if *nPerson.Age != 32 {
		t.Errorf("UpdateEntity failed")
		return
	}
	if *nPerson.Addr != "hey boy!" {
		t.Errorf("UpdateEntity failed")
		return
	}
}
