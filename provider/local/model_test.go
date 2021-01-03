package local

import (
	"reflect"
	"testing"
	"time"
)

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int64 `json:"id" orm:"id key"`
	// Name 名称
	Name      string    `json:"name" orm:"name"`
	Value     float32   `json:"value" orm:"value"`
	TimeStamp time.Time `json:"timeStamp" orm:"timeStamp"`
	T1        Test      `orm:"t1"`
}

type BT struct {
	ID  int `orm:"id key"`
	Val int `orm:"val"`
}

type Base struct {
	ID  int `orm:"id key"`
	Val int `orm:"val"`
	Bt  BT  `orm:"bt"`
}

type Test struct {
	ID    int  `orm:"id key"`
	Val   int  `orm:"val"`
	Base  Base `orm:"b1"`
	Base2 BT   `orm:"b2"`
}

func TestModelValue(t *testing.T) {
	now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-01-02 15:04:05", time.Local)
	unit := Unit{Name: "AA", T1: Test{Val: 123}, TimeStamp: now}
	unitVal := reflect.ValueOf(&unit).Elem()
	unitInfo, unitErr := getTypeModel(unitVal.Type())
	if unitErr != nil {
		t.Errorf("getValueModel failed, unitErr:%s", unitErr.Error())
		return
	}

	id := int64(123320)
	iVal := newValue(reflect.ValueOf(id))
	pk := unitInfo.GetPrimaryField()
	if pk == nil {
		t.Errorf("GetPrimaryField faield")
		return
	}
	unitErr = pk.SetValue(iVal)
	if unitErr != nil {
		t.Errorf("Set value failed, unitErr:%s", unitErr.Error())
		return
	}

	name := "abcdfrfe"
	nVal := newValue(reflect.ValueOf(name))
	unitErr = unitInfo.SetFieldValue("Name", nVal)
	if unitErr != nil {
		t.Errorf("UpdateField value failed, unitErr:%s", unitErr.Error())
		return
	}

	now = time.Now()
	tsVal := newValue(reflect.ValueOf(now))
	unitErr = unitInfo.SetFieldValue("TimeStamp", tsVal)
	if unitErr != nil {
		t.Errorf("UpdateField value failed, unitErr:%s", unitErr.Error())
		return
	}

	uVal := unitInfo.Interface()
	unit = uVal.Get().(reflect.Value).Interface().(Unit)
	if unit.ID != int64(id) {
		t.Errorf("update id field failed")
		return
	}
	if unit.Name != name {
		t.Errorf("update id field failed")
		return
	}
	if !unit.TimeStamp.Equal(now) {
		t.Errorf("update timeStamp failed")
		return
	}
}

func TestReference(t *testing.T) {
	type AB struct {
		F32 float32 `orm:"f32 key"`
	}

	type CD struct {
		AB  AB  `orm:"ab"`
		I64 int `orm:"i64 key"`
	}

	type Demo struct {
		II int   `orm:"ii key"`
		AB *AB   `orm:"ab"`
		CD []int `orm:"cd"`
		EF []*AB `orm:"ef"`
	}

	abVal := reflect.ValueOf(&AB{})
	cdVal := reflect.ValueOf(&CD{}).Elem()
	demoVal := reflect.ValueOf(&Demo{AB: &AB{}}).Elem()
	_, err := getTypeModel(abVal.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}
	_, err = getTypeModel(cdVal.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}
	_, err = getTypeModel(demoVal.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}
	f32Info, err := getValueModel(demoVal)
	if err != nil {
		t.Errorf("getValueModel failed, err:%s", err.Error())
		return
	}

	f32Info.Dump()

	i64Info, err := getValueModel(cdVal)
	if err != nil {
		t.Errorf("getValueModel failed, err:%s", err.Error())
	}

	i64Info.Dump()
}

type TT struct {
	Aa int `orm:"aa key auto"`
	Bb int `orm:"bb"`
	Tt *TT `orm:"tt"`
}

func TestGetModelValue(t *testing.T) {
	t1 := TT{Aa: 12, Bb: 23}
	ttVal := reflect.ValueOf(&t1).Elem()
	_, err := getTypeModel(ttVal.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}
	t1Info, t1Err := getValueModel(ttVal)
	if t1Err != nil {
		t.Errorf("getValueModel t1 failed, err:%s", t1Err.Error())
		return
	}

	t2 := &TT{Aa: 34, Bb: 45}
	//reflect.TypeOf(t2)
	t2Info, t2Err := getValueModel(reflect.ValueOf(t2).Elem())
	if t1Err != nil {
		t.Errorf("getValueModel t2 failed, err:%s", t2Err.Error())
		return
	}

	t1Info.Dump()
	t2Info.Dump()
}
