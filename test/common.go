package test

import (
	"time"

	orm "github.com/muidea/magicOrm"

	"github.com/muidea/magicOrm/provider/remote"
)

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

func registerModel(orm orm.Orm, objList []interface{}) (err error) {
	for _, val := range objList {
		err = orm.RegisterModel(val, "default")
		if err != nil {
			return
		}
	}

	return
}

func getObjectValue(val interface{}) (ret *remote.ObjectValue, err error) {
	objVal, objErr := remote.GetObjectValue(val)
	if objErr != nil {
		err = objErr
		return
	}

	data, dataErr := remote.EncodeObjectValue(objVal)
	if dataErr != nil {
		err = dataErr
		return
	}
	ret, err = remote.DecodeObjectValue(data)
	if err != nil {
		return
	}

	return
}
