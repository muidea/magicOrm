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

func TestModel(t *testing.T) {
	now := time.Now()

	test := reflect.ValueOf(Test{})
	if test.CanSet() {
		t.Error("invalid value")
		return
	}

	ptr := &Test{}
	test = reflect.ValueOf(ptr)
	if test.CanSet() {
		t.Error("invalid value")
		return
	}

	test = test.Elem()
	if !test.CanSet() {
		t.Error("invalid value")
		return
	}

	_, err := getTypeModel(test.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}

	testInfo, testErr := getValueModel(test)
	if testErr != nil {
		t.Errorf("getValueModel failed, err:%s", testErr.Error())
		return
	}
	fields := testInfo.GetFields()
	for _, val := range fields {
		if val.IsAssigned() {
			t.Errorf("invalid filed,name:%s", val.GetName())
			return
		}
	}

	test = reflect.ValueOf(Test{ID: 10, Val: 123, Base: Base{ID: 12}, Base2: BT{ID: 23}})
	testInfo, testErr = getValueModel(test)
	if testErr != nil {
		t.Errorf("getValueModel failed, err:%s", testErr.Error())
		return
	}
	fields = testInfo.GetFields()
	for _, val := range fields {
		if !val.IsAssigned() {
			t.Errorf("invalid filed,name:%s", val.GetName())
			return
		}
	}

	val := reflect.ValueOf(&Unit{T1: Test{ID: 12, Val: 123}, TimeStamp: now}).Elem()
	_, err = getTypeModel(val.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}

	modelInfo, err := getValueModel(val)
	if modelInfo == nil || err != nil {
		t.Errorf("getValueModel failed, err:%s", err.Error())
		return
	}

	fields = modelInfo.GetFields()
	if len(fields) != 5 {
		t.Errorf("get value Model failed")
		return
	}

	modelInfo.Dump()
}

func TestModelValue(t *testing.T) {
	now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-01-02 15:04:05", time.Local)
	unit := Unit{Name: "AA", T1: Test{Val: 123}, TimeStamp: now}
	unitVal := reflect.ValueOf(&unit).Elem()
	_, err := getTypeModel(unitVal.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}

	modelInfo, err := getValueModel(unitVal)
	if err != nil {
		t.Errorf("getValueModel failed, err:%s", err.Error())
		return
	}

	id := 123320
	pk := modelInfo.GetPrimaryField()
	if pk == nil {
		t.Errorf("GetPrimaryField faield")
		return
	}
	err = pk.UpdateValue(reflect.ValueOf(id))
	if err != nil {
		t.Errorf("Set value failed, err:%s", err.Error())
		return
	}

	name := "abcdfrfe"
	err = modelInfo.UpdateFieldValue("Name", reflect.ValueOf(name))
	if err != nil {
		t.Errorf("UpdateField value failed, err:%s", err.Error())
		return
	}

	now = time.Now()
	tsVal := reflect.ValueOf(now)
	err = modelInfo.UpdateFieldValue("TimeStamp", tsVal)
	if err != nil {
		t.Errorf("UpdateField value failed, err:%s", err.Error())
		return
	}

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
