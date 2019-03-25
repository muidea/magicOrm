package local

import (
	"log"
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
	cache := NewCache()
	now := time.Now()
	modelInfo, err := getObjectModel(&Unit{T1: Test{ID: 12, Val: 123}, TimeStamp: now}, cache)
	if modelInfo == nil || err != nil {
		t.Errorf("getObjectModel failed, err:%s", err.Error())
		return
	}

	modelInfo.Dump(cache)
}

func TestModelValue(t *testing.T) {
	cache := NewCache()
	now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-01-02 15:04:05", time.Local)
	unit := &Unit{Name: "AA", T1: Test{Val: 123}, TimeStamp: now}
	modelInfo, err := getObjectModel(unit, cache)
	if err != nil {
		t.Errorf("getObjectModel failed, err:%s", err.Error())
		return
	}

	log.Print(*unit)

	id := int64(123320)
	pk := modelInfo.GetPrimaryField()
	if pk == nil {
		t.Errorf("GetPrimaryField faield")
		return
	}
	fv := pk.GetValue()
	fv.Set(reflect.ValueOf(id))

	name := "abcdfrfe"
	modelInfo.UpdateFieldValue("Name", reflect.ValueOf(name))

	now = time.Now()
	tsVal := reflect.ValueOf(now)
	modelInfo.UpdateFieldValue("TimeStamp", tsVal)

	log.Print(*unit)
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

	cache := NewCache()
	f32Info, err := getObjectModel(&Demo{AB: &AB{}}, cache)
	if err != nil {
		t.Errorf("getObjectModel failed, err:%s", err.Error())
		return
	}

	f32Info.Dump(cache)

	i64Info, err := getObjectModel(&CD{}, cache)
	if err != nil {
		t.Errorf("getObjectModel failed, err:%s", err.Error())
	}

	i64Info.Dump(cache)
}

type TT struct {
	Aa int `orm:"aa key auto"`
	Bb int `orm:"bb"`
	Tt *TT `orm:"tt"`
}

func TestGetModelValue(t *testing.T) {
	cache := NewCache()
	t1 := &TT{Aa: 12, Bb: 23}
	t1Info, t1Err := getObjectModel(t1, cache)
	if t1Err != nil {
		t.Errorf("getObjectModel t1 failed, err:%s", t1Err.Error())
		return
	}

	t2 := &TT{Aa: 34, Bb: 45}
	//reflect.TypeOf(t2)
	t2Info, t2Err := getValueModel(reflect.ValueOf(t2), cache)
	if t1Err != nil {
		t.Errorf("getObjectModel t2 failed, err:%s", t2Err.Error())
		return
	}

	t1Info.Dump(cache)
	t2Info.Dump(cache)
}
