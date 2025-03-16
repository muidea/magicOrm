package test

import (
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

type Simple struct {
	ID        int       `orm:"id key auto" view:"detail,lite"`
	I8        int8      `orm:"i8" view:"detail,lite"`
	I16       int16     `orm:"i16" view:"detail,lite"`
	I32       int32     `orm:"i32" view:"detail,lite"`
	I64       uint64    `orm:"i64" view:"detail,lite"`
	Name      string    `orm:"name" view:"detail,lite"`
	Value     float32   `orm:"value" view:"detail,lite"`
	F64       float64   `orm:"f64" view:"detail,lite"`
	TimeStamp time.Time `orm:"ts datetime" view:"detail,lite"`
	Flag      bool      `orm:"flag" view:"detail,lite"`
}

func (l *Simple) IsSame(r *Simple) bool {
	if l.ID != r.ID {
		return false
	}
	if l.I8 != r.I8 {
		return false
	}
	if l.I16 != r.I16 {
		return false
	}
	if l.I32 != r.I32 {
		return false
	}
	if l.I64 != r.I64 {
		return false
	}
	if l.Name != r.Name {
		return false
	}
	if l.Value != r.Value {
		return false
	}
	if l.F64 != r.F64 {
		return false
	}
	if l.TimeStamp.Sub(r.TimeStamp) >= time.Second {
		return false
	}
	if l.Flag != r.Flag {
		return false
	}

	return true
}

type Reference struct {
	ID          int       `orm:"id key auto" view:"detail,lite"`
	Name        string    `orm:"name" view:"detail,lite"`
	FValue      float32   `orm:"value" view:"detail,lite"`
	F64         float64   `orm:"f64" view:"detail,lite"`
	TimeStamp   time.Time `orm:"ts" view:"detail,lite"`
	Flag        bool      `orm:"flag" view:"detail,lite"`
	IArray      []int     `orm:"iArray" view:"detail,lite"`
	FArray      []float32 `orm:"fArray" view:"detail,lite"`
	StrArray    []string  `orm:"strArray" view:"detail,lite"`
	BArray      []bool    `orm:"bArray" view:"detail,lite"`
	PtrArray    *[]string `orm:"ptrArray" view:"detail,lite"`
	StrPtrArray []string  `orm:"strPtrArray" view:"detail,lite"`
	PtrStrArray *[]string `orm:"ptrStrArray" view:"detail,lite"`
}

func (l *Reference) IsSame(r *Reference) bool {
	if l.ID != r.ID {
		return false
	}
	if l.Name != r.Name {
		return false
	}
	if l.F64 != r.F64 {
		return false
	}
	if l.FValue != r.FValue {
		return false
	}
	if l.TimeStamp != r.TimeStamp {
		return false
	}
	if l.Flag != r.Flag {
		return false
	}
	if len(l.IArray) != len(r.IArray) {
		return false
	}
	if len(l.FArray) != len(r.FArray) {
		return false
	}
	if len(l.StrArray) != len(r.StrArray) {
		return false
	}
	if len(l.BArray) != len(r.BArray) {
		return false
	}
	if l.PtrArray != nil && len(*l.PtrArray) > 0 {
		if r.PtrArray == nil {
			return false
		}
		if len(*l.PtrArray) != len(*r.PtrArray) {
			return false
		}
	}
	if l.PtrArray == nil {
		if r.PtrArray != nil && len(*r.PtrArray) > 0 {
			return false
		}
	}
	if len(l.StrPtrArray) != len(r.StrPtrArray) {
		return false
	}
	if l.PtrStrArray != nil && len(*l.PtrStrArray) > 0 {
		if r.PtrStrArray == nil {
			return false
		}
		if len(*l.PtrStrArray) != len(*r.PtrStrArray) {
			return false
		}
	}
	if l.PtrStrArray == nil {
		if r.PtrStrArray != nil && len(*r.PtrStrArray) > 0 {
			return false
		}
	}

	return true
}

type Compose struct {
	ID   int    `orm:"id key auto" view:"detail,lite"`
	Name string `orm:"name" view:"detail,lite"`
	// 1
	Simple Simple `orm:"simple" view:"detail,lite"`
	// 3
	SimplePtr *Simple `orm:"simplePtr" view:"detail,lite"`
	// 2
	SimpleArray []Simple `orm:"simpleArray" view:"detail,lite"`
	// 4
	SimplePtrArray    []*Simple    `orm:"simplePtrArray" view:"detail,lite"`
	SimpleArrayPtr    *[]Simple    `orm:"simpleArrayPtr" view:"detail,lite"`
	Reference         Reference    `orm:"reference" view:"detail,lite"`
	ReferencePtr      *Reference   `orm:"referencePtr" view:"detail,lite"`
	ReferenceArray    []Reference  `orm:"referenceArray" view:"detail,lite"`
	ReferencePtrArray []*Reference `orm:"referencePtrArray" view:"detail,lite"`
	ComposePtr        *Compose     `orm:"composePtr" view:"detail,lite"`
}

func (l *Compose) IsSame(r *Compose) bool {
	if l.ID != r.ID {
		return false
	}
	if l.Name != r.Name {
		return false
	}
	if l.Simple.ID != r.Simple.ID {
		return false
	}
	if l.SimplePtr != nil {
		if r.SimplePtr == nil {
			return false
		}
		if l.SimplePtr.ID != r.SimplePtr.ID {
			return false
		}
	}
	if l.SimplePtr == nil {
		if r.SimplePtr != nil {
			return false
		}
	}
	if len(l.SimpleArray) != len(r.SimpleArray) {
		return false
	}
	if len(l.SimplePtrArray) != len(r.SimplePtrArray) {
		return false
	}
	if l.SimpleArrayPtr != nil && len(*l.SimpleArrayPtr) > 0 {
		if r.SimpleArrayPtr == nil {
			return false
		}
		if len(*l.SimpleArrayPtr) != len(*r.SimpleArrayPtr) {
			return false
		}
	}
	if l.SimpleArrayPtr == nil {
		if r.SimpleArrayPtr != nil && len(*r.SimpleArrayPtr) > 0 {
			return false
		}
	}
	if l.Reference.ID != r.Reference.ID {
		return false
	}
	if l.ReferencePtr != nil {
		if r.ReferencePtr == nil {
			return false
		}
		if l.ReferencePtr.ID != r.ReferencePtr.ID {
			return false
		}
	}
	if l.ReferencePtr == nil {
		if r.ReferencePtr != nil {
			return false
		}
	}
	if len(l.ReferenceArray) != len(r.ReferenceArray) {
		return false
	}
	if len(l.ReferencePtrArray) != len(r.ReferencePtrArray) {
		return false
	}
	if l.ComposePtr != nil {
		if r.ComposePtr == nil {
			return false
		}

		if l.ComposePtr.ID != r.ComposePtr.ID {
			return false
		}
	}
	if l.ComposePtr == nil {
		if r.ComposePtr != nil {
			return false
		}
	}

	return true
}

func registerModel(provider provider.Provider, objList []any) (ret []model.Model, err *cd.Result) {
	for _, val := range objList {
		m, mErr := provider.RegisterModel(val)
		if mErr != nil {
			err = mErr
			return
		}

		ret = append(ret, m)
	}

	return
}

func createModel(orm orm.Orm, modelList []model.Model) (err *cd.Result) {
	for _, val := range modelList {
		err = orm.Create(val)
		if err != nil {
			return
		}
	}

	return
}

func dropModel(orm orm.Orm, modelList []model.Model) (err *cd.Result) {
	for _, val := range modelList {
		err = orm.Drop(val)
		if err != nil {
			return
		}
	}

	return
}

func getObjectValue(val any) (ret *remote.ObjectValue, err *cd.Result) {
	objVal, objErr := helper.GetObjectValue(val)
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
