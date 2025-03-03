package test

import (
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
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
	H1 Simple `orm:"simple" view:"detail,lite"`
	// 3
	R3 *Simple `orm:"ptrSimple" view:"detail,lite"`
	// 2
	H2 []Simple `orm:"simpleArray" view:"detail,lite"`
	// 4
	R4           []*Simple    `orm:"simplePtrArray" view:"detail,lite"`
	PR4          *[]Simple    `orm:"ptrSimpleArray" view:"detail,lite"`
	Reference    Reference    `orm:"reference" view:"detail,lite"`
	PtrReference *Reference   `orm:"ptrReference" view:"detail,lite"`
	RefArray     []Reference  `orm:"refArray" view:"detail,lite"`
	RefPtrArray  []*Reference `orm:"refPtrArray" view:"detail,lite"`
	PtrRefArray  []*Reference `orm:"ptrRefArray" view:"detail,lite"`
	PtrCompose   *Compose     `orm:"ptrCompose" view:"detail,lite"`
}

func (l *Compose) IsSame(r *Compose) bool {
	if l.ID != r.ID {
		return false
	}
	if l.Name != r.Name {
		return false
	}
	if l.H1.ID != r.H1.ID {
		return false
	}
	if l.R3 != nil {
		if r.R3 == nil {
			return false
		}
		if l.R3.ID != r.R3.ID {
			return false
		}
	}
	if l.R3 == nil {
		if r.R3 != nil {
			return false
		}
	}
	if len(l.H2) != len(r.H2) {
		return false
	}
	if len(l.R4) != len(r.R4) {
		return false
	}
	if l.PR4 != nil && len(*l.PR4) > 0 {
		if r.PR4 == nil {
			return false
		}
		if len(*l.PR4) != len(*r.PR4) {
			return false
		}
	}
	if l.PR4 == nil {
		if r.PR4 != nil && len(*r.PR4) > 0 {
			return false
		}
	}
	if l.Reference.ID != r.Reference.ID {
		return false
	}
	if l.PtrReference != nil {
		if r.PtrReference == nil {
			return false
		}
		if l.PtrReference.ID != r.PtrReference.ID {
			return false
		}
	}
	if l.PtrReference == nil {
		if r.PtrReference != nil {
			return false
		}
	}
	if len(l.RefArray) != len(r.RefArray) {
		return false
	}
	if len(l.RefPtrArray) != len(r.RefPtrArray) {
		return false
	}
	if len(l.PtrRefArray) > 0 {
		if r.PtrRefArray == nil {
			return false
		}
		if len(l.PtrRefArray) != len(r.PtrRefArray) {
			return false
		}
	}
	if l.PtrRefArray == nil {
		if len(r.PtrRefArray) > 0 {
			return false
		}
	}
	if l.PtrCompose != nil {
		if r.PtrCompose == nil {
			return false
		}

		if l.PtrCompose.ID != r.PtrCompose.ID {
			return false
		}
	}
	if l.PtrCompose == nil {
		if r.PtrCompose != nil {
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
