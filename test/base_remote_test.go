package test

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

const remoteOwner = "remote"

//var config = orm.NewConfig("localhost:5432", "magicplatform_db", "postgres", "rootkit", "")

func TestRemoteSimple(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	remoteProvider := provider.NewRemoteProvider(remoteOwner)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Simple{}}
	mList, mErr := registerRemoteModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("registerRemoteModel failed, err:%s", mErr.Error())
		return
	}

	err = dropModel(o1, mList)
	if err != nil {
		t.Errorf("drop reference schema failed, err:%s", err.Error())
		return
	}

	err = createModel(o1, mList)
	if err != nil {
		t.Errorf("create reference schema failed, err:%s", err.Error())
		return
	}

	ts, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	s1 := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}

	s1Value, s1Err := getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}

	byteVal, byteErr := remote.EncodeObjectValue(s1Value)
	if byteErr != nil {
		t.Errorf("encode object value failed, err:%s", byteErr.Error())
		return
	}

	s1Value, s1Err = remote.DecodeObjectValue(byteVal)
	if s1Err != nil {
		t.Errorf("decode object value failed, err:%s", s1Err.Error())
		return
	}
	s1Model, s1Err := remoteProvider.GetEntityModel(s1Value)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	s1Model, s1Err = o1.Insert(s1Model)
	if s1Err != nil {
		err = s1Err
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s1Model.Interface(true).(*remote.ObjectValue), s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	s1.Name = "hello"
	s1Value, s1Err = getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}

	s1Model, s1Err = remoteProvider.GetEntityModel(s1Value)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	s1Model, s1Err = o1.Update(s1Model)
	if s1Err != nil {
		err = s1Err
		t.Errorf("update reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s1Model.Interface(true).(*remote.ObjectValue), s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	s2 := &Simple{ID: s1.ID}
	s2Value, s2Err := getObjectValue(s2)
	if s2Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s2Err.Error())
		return
	}
	s2Model, s2Err := remoteProvider.GetEntityModel(s2Value)
	if s2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s2Err.Error())
		return
	}
	s2Model, s2Err = o1.Query(s2Model)
	if s2Err != nil {
		err = s2Err
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s2Model.Interface(true).(*remote.ObjectValue), s2)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}
	if !s1.IsSame(s2) {
		t.Errorf("Query reference failed.")
	}

}

func TestRemoteReference(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	remoteProvider := provider.NewRemoteProvider(remoteOwner)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Reference{}}
	mList, mErr := registerRemoteModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("register model failed. err:%s", mErr.Error())
		return
	}

	err = dropModel(o1, mList)
	if err != nil {
		t.Errorf("drop reference schema failed, err:%s", err.Error())
		return
	}

	err = createModel(o1, mList)
	if err != nil {
		t.Errorf("create reference schema failed, err:%s", err.Error())
		return
	}

	ts, _ := time.ParseInLocation(util.CSTLayout, "2018-01-02 15:04:05", time.Local)
	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []string{strValue, strValue}
	s1 := &Reference{
		Name:        strValue,
		FValue:      fValue,
		F64:         23.456,
		TimeStamp:   ts,
		Flag:        flag,
		IArray:      iArray,
		FArray:      fArray,
		StrArray:    strArray,
		BArray:      bArray,
		PtrArray:    &strArray,
		StrPtrArray: strPtrArray,
		PtrStrArray: &strPtrArray,
	}

	s1Value, s1Err := getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}

	s1Model, s1Err := remoteProvider.GetEntityModel(s1Value)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}

	s1Model, s1Err = o1.Insert(s1Model)
	if s1Err != nil {
		err = s1Err
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s1Model.Interface(true).(*remote.ObjectValue), s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	s1.Name = "hello"
	s1Value, s1Err = getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}
	s1Model, s1Err = remoteProvider.GetEntityModel(s1Value)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	s1Model, s1Err = o1.Update(s1Model)
	if s1Err != nil {
		err = s1Err
		t.Errorf("update reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s1Model.Interface(true).(*remote.ObjectValue), s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	fValue2 := float32(0.0)
	var ts2 time.Time
	var bVal bool
	strArray2 := []string{}
	ptrArray2 := []string{}
	s2 := &Reference{
		ID:          s1.ID,
		FValue:      fValue2,
		TimeStamp:   ts2,
		Flag:        bVal,
		IArray:      []int{},
		FArray:      []float32{},
		StrArray:    []string{},
		BArray:      []bool{},
		PtrArray:    &strArray2,
		StrPtrArray: []string{},
		PtrStrArray: &ptrArray2,
	}
	s2Value, s2Err := getObjectValue(s2)
	if s2Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s2Err.Error())
		return
	}
	s2Model, s2Err := remoteProvider.GetEntityModel(s2Value)
	if s2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s2Err.Error())
		return
	}
	s2Model, s2Err = o1.Query(s2Model)
	if s2Err != nil {
		err = s2Err
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s2Model.Interface(true).(*remote.ObjectValue), s2)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}
	if !s1.IsSame(s2) {
		t.Errorf("Query reference failed.")
		return
	}

	s4 := &Reference{
		ID:     s1.ID,
		IArray: []int{},
	}
	s4Value, s4Err := getObjectValue(s4)
	if s4Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s4Err.Error())
		return
	}
	s4Model, s4Err := remoteProvider.GetEntityModel(s4Value)
	if s4Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s4Err.Error())
		return
	}
	s4Model, s4Err = o1.Query(s4Model)
	if s4Err != nil {
		err = s4Err
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s4Model.Interface(true).(*remote.ObjectValue), s4)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}
	if s4.Name != s2.Name {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	if s4.IArray == nil || s4.PtrStrArray != nil || s4.PtrArray != nil {
		t.Errorf("query reference failed")
		return
	}
}

func TestRemoteCompose(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	remoteProvider := provider.NewRemoteProvider(remoteOwner)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Simple{}, &Reference{}, &Compose{}}
	mList, mErr := registerRemoteModel(remoteProvider, objList)
	if mErr != nil {
		err = mErr
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	err = dropModel(o1, mList)
	if err != nil {
		t.Errorf("drop model failed. err:%s", err.Error())
		return
	}

	err = createModel(o1, mList)
	if err != nil {
		t.Errorf("create model failed. err:%s", err.Error())
		return
	}

	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	s1 := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}

	s1Val, s1Err := getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}

	s1Model, s1Err := remoteProvider.GetEntityModel(s1Val)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	s1Model, err = o1.Insert(s1Model)
	if err != nil {
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}
	s1Val = s1Model.Interface(true).(*remote.ObjectValue)
	err = helper.UpdateEntity(s1Val, s1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []string{strValue, strValue}
	r1 := &Reference{
		Name:        strValue,
		FValue:      fValue,
		F64:         23.456,
		TimeStamp:   ts,
		Flag:        flag,
		IArray:      iArray,
		FArray:      fArray,
		StrArray:    strArray,
		BArray:      bArray,
		PtrArray:    &strArray,
		StrPtrArray: strPtrArray,
		PtrStrArray: &strPtrArray,
	}
	r1Val, r1Err := getObjectValue(r1)
	if r1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", r1Err.Error())
		return
	}

	r1Model, r1Err := remoteProvider.GetEntityModel(r1Val)
	if r1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", r1Err.Error())
		return
	}

	r1Model, err = o1.Insert(r1Model)
	if err != nil {
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}

	r1Val = r1Model.Interface(true).(*remote.ObjectValue)
	err = helper.UpdateEntity(r1Val, r1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	refPtrArray := []*Reference{r1}
	c1 := &Compose{
		Name:              strValue,
		Simple:            *s1,
		SimplePtr:         s1,
		SimpleArray:       []Simple{*s1, *s1},
		SimplePtrArray:    []*Simple{s1, s1},
		Reference:         *r1,
		ReferencePtr:      r1,
		ReferenceArray:    []Reference{*r1, *r1, *r1},
		ReferencePtrArray: refPtrArray,
	}
	c1Val, c1Err := getObjectValue(c1)
	if c1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", c1Err.Error())
		return
	}
	c1Model, c1Err := remoteProvider.GetEntityModel(c1Val)
	if c1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c1Err.Error())
		return
	}
	c1Model, err = o1.Insert(c1Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	c1Val = c1Model.Interface(true).(*remote.ObjectValue)

	err = helper.UpdateEntity(c1Val, c1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	c2 := &Compose{
		Name:              strValue,
		Simple:            *s1,
		SimplePtr:         s1,
		SimpleArray:       []Simple{*s1, *s1},
		SimplePtrArray:    []*Simple{s1, s1},
		SimpleArrayPtr:    &[]Simple{},
		Reference:         *r1,
		ReferencePtr:      r1,
		ReferenceArray:    []Reference{*r1, *r1, *r1},
		ReferencePtrArray: refPtrArray,
		ComposePtr:        c1,
	}
	c2Val, c2Err := getObjectValue(c2)
	if c2Err != nil {
		t.Errorf("getObjectValue failed, err:%s", c2Err.Error())
		return
	}
	c2Model, c2Err := remoteProvider.GetEntityModel(c2Val)
	if c2Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", c2Err.Error())
		return
	}

	c2Model, err = o1.Insert(c2Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	c2Val = c2Model.Interface(true).(*remote.ObjectValue)
	err = helper.UpdateEntity(c2Val, c2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	c3 := &Compose{
		ID:                c2.ID,
		SimplePtr:         &Simple{},
		SimpleArray:       []Simple{},
		SimplePtrArray:    []*Simple{},
		SimpleArrayPtr:    &[]Simple{},
		ReferencePtr:      &Reference{},
		ReferenceArray:    []Reference{},
		ReferencePtrArray: []*Reference{},
		ComposePtr:        &Compose{},
	}
	c3Val, c3Err := getObjectValue(c3)
	if c3Err != nil {
		t.Errorf("getObjectValue failed, err:%s", c3Err.Error())
		return
	}
	c3Model, c3Err := remoteProvider.GetEntityModel(c3Val)
	if c3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c3Err.Error())
		return
	}

	c3Model, err = o1.Query(c3Model)
	if err != nil {
		t.Errorf("query compose failed, err:%s", err.Error())
		return
	}
	c3Val = c3Model.Interface(true).(*remote.ObjectValue)
	err = helper.UpdateEntity(c3Val, c3)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	if c3.IsSame(c1) {
		t.Error("query compose failed")
		return
	}
	if !c3.IsSame(c2) {
		t.Error("query compose failed")
		return
	}
}

func TestRemoteQuery(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	remoteProvider := provider.NewRemoteProvider(remoteOwner)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	s1 := &Simple{}
	s1Def, s1Err := helper.GetObject(s1)
	if s1Err != nil {
		t.Errorf("GetObject failed, err:%s", s1Err.Error())
		return
	}

	r1 := &Reference{}
	r1Def, r1Err := helper.GetObject(r1)
	if r1Err != nil {
		t.Errorf("GetObject failed, err:%s", r1Err.Error())
		return
	}

	c1 := &Compose{}
	c1Def, c1Err := helper.GetObject(c1)
	if c1Err != nil {
		t.Errorf("GetObject failed, err:%s", c1Err.Error())
		return
	}

	objList := []any{s1Def, r1Def, c1Def}
	mList, mErr := registerLocalModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("register model failed. err:%s", mErr.Error())
		return
	}

	for _, val := range mList {
		err = o1.Drop(val)
		if err != nil {
			t.Errorf("drop object failed, err:%s", err.Error())
		}
	}

	for _, val := range mList {
		err = o1.Create(val)
		if err != nil {
			t.Errorf("create object failed, err:%s", err.Error())
		}
	}

	ts, _ := time.Parse("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000")
	s1 = &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}
	s1Value, s1Err := getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}
	s1Model, s1Err := remoteProvider.GetEntityModel(s1Value)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}

	s1Model, s1Err = o1.Insert(s1Model)
	if s1Err != nil {
		err = s1Err
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s1Model.Interface(true).(*remote.ObjectValue), s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []string{strValue, strValue}
	r1 = &Reference{
		Name:        strValue,
		FValue:      fValue,
		F64:         23.456,
		TimeStamp:   ts,
		Flag:        flag,
		IArray:      iArray,
		FArray:      fArray,
		StrArray:    strArray,
		BArray:      bArray,
		PtrArray:    &strArray,
		StrPtrArray: strPtrArray,
		PtrStrArray: &strPtrArray,
	}
	r1Value, r1Err := getObjectValue(r1)
	if r1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", r1Err.Error())
		return
	}
	r1Model, r1Err := remoteProvider.GetEntityModel(r1Value)
	if r1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", r1Err.Error())
		return
	}
	r1Model, r1Err = o1.Insert(r1Model)
	if r1Err != nil {
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(r1Model.Interface(true).(*remote.ObjectValue), r1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	refPtrArray := []*Reference{r1}
	c1 = &Compose{
		Name:              strValue,
		Simple:            *s1,
		SimplePtr:         s1,
		SimpleArray:       []Simple{*s1, *s1},
		SimplePtrArray:    []*Simple{s1, s1},
		Reference:         *r1,
		ReferencePtr:      r1,
		ReferenceArray:    []Reference{*r1, *r1, *r1},
		ReferencePtrArray: refPtrArray,
	}
	c1Value, c1Err := getObjectValue(c1)
	if c1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", c1Err.Error())
		return
	}
	c1Model, c1Err := remoteProvider.GetEntityModel(c1Value)
	if c1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c1Err.Error())
		return
	}

	c1Model, c1Err = o1.Insert(c1Model)
	if c1Err != nil {
		err = c1Err
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(c1Model.Interface(true).(*remote.ObjectValue), c1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	strValue = "123"
	c2 := &Compose{
		Name:              strValue,
		Simple:            *s1,
		SimplePtr:         s1,
		SimpleArray:       []Simple{*s1, *s1},
		SimplePtrArray:    []*Simple{s1, s1},
		Reference:         *r1,
		ReferencePtr:      r1,
		ReferenceArray:    []Reference{*r1, *r1, *r1},
		ReferencePtrArray: refPtrArray,
		ComposePtr:        c1,
	}
	c2Value, c2Err := getObjectValue(c2)
	if c2Err != nil {
		t.Errorf("getObjectValue failed, err:%s", c2Err.Error())
		return
	}
	c2Model, c2Err := remoteProvider.GetEntityModel(c2Value)
	if c2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c2Err.Error())
		return
	}
	c2Model, c2Err = o1.Insert(c2Model)
	if c2Err != nil {
		err = c2Err
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(c2Model.Interface(true).(*remote.ObjectValue), c2)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	c3 := &Compose{ID: c2.ID, SimplePtr: &Simple{}, SimpleArrayPtr: &[]Simple{}, ReferencePtr: &Reference{}, ComposePtr: &Compose{}}
	c3Value := c2Value
	c3Model, c3Err := remoteProvider.GetEntityModel(c3Value)
	if c3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c3Err.Error())
		return
	}
	c3Model, c3Err = o1.Insert(c3Model)
	if c3Err != nil {
		err = c3Err
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}

	err = helper.UpdateEntity(c3Model.Interface(true).(*remote.ObjectValue), c3)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	c4 := &Compose{}
	c4Value := c2Value
	c4Model, c4Err := remoteProvider.GetEntityModel(c4Value)
	if c4Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c4Err.Error())
		return
	}
	c4Model, c4Err = o1.Insert(c4Model)
	if c4Err != nil {
		err = c4Err
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}

	err = helper.UpdateEntity(c4Model.Interface(true).(*remote.ObjectValue), c4)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	cComposePtr := &Compose{}
	cObjectValue, cObjectValueErr := helper.GetObjectValue(cComposePtr)
	if cObjectValueErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", cObjectValueErr.Error())
		return
	}
	cObjectPtr, cObjectErr := helper.GetObject(cComposePtr)
	if cObjectErr != nil {
		t.Errorf("GetObject failed, err:%s", cObjectErr.Error())
		return
	}

	filter, err := remoteProvider.GetModelFilter(cObjectPtr)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	filter.ValueMask(cObjectValue)
	cModelList, cModelErr := o1.BatchQuery(filter)
	if cModelErr != nil {
		err = cModelErr
		t.Errorf("batch query compose failed, err:%s", err.Error())
		return
	}
	if len(cModelList) != 4 {
		t.Errorf("batch query compose failed")
		return
	}

	maskVal, maskErr := helper.GetObjectValue(&Compose{SimplePtr: &Simple{}})
	if maskErr != nil {
		t.Errorf("getObjectValue failed, err:%s", maskErr.Error())
		return
	}

	filter.Equal("name", strValue)
	filter.ValueMask(maskVal)
	cModelList, cModelErr = o1.BatchQuery(filter)
	if cModelErr != nil {
		t.Errorf("batch query compose failed, err:%s", cModelErr.Error())
		return
	}
	if len(cModelList) != 3 {
		t.Errorf("batch query compose failed")
		return
	}
}

func TestRemoteOnlineEntity(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	remoteProvider := provider.NewRemoteProvider(remoteOwner)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Entity{}, &OnlineEntity{}}
	mList, mErr := registerRemoteModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("register model failed. err:%s", mErr.Error())
		return
	}

	err = dropModel(o1, mList)
	if err != nil {
		t.Errorf("drop model failed. err:%s", err.Error())
		return
	}

	err = createModel(o1, mList)
	if err != nil {
		t.Errorf("create model failed. err:%s", err.Error())
		return
	}

	entityPtr := &Entity{EName: "test", EType: "test", EID: 10, Namespace: "test"}
	remoteEntityVal, remoteEntityErr := getObjectValue(entityPtr)
	if remoteEntityErr != nil {
		t.Errorf("getObjectValue failed. err:%s", remoteEntityErr.Error())
		return
	}

	entityModelVal, entityModelErr := remoteProvider.GetEntityModel(remoteEntityVal)
	if entityModelErr != nil {
		t.Errorf("get entity model failed, err:%s", entityModelErr.Error())
		return
	}

	entityModelVal, entityModelErr = o1.Insert(entityModelVal)
	if entityModelErr != nil {
		t.Errorf("insert entity model failed, err:%s", entityModelErr.Error())
		return
	}

	newRemoteEntityPtr := entityModelVal.Interface(true).(*remote.ObjectValue)
	if len(newRemoteEntityPtr.Fields) != 5 {
		t.Errorf("insert entity model failed")
		return
	}

	entityPtr.ID = 1
	onlineEntityPtr := &OnlineEntity{
		SessionID:   "abc",
		Entity:      entityPtr,
		RefreshTime: 10000,
		ExpireTime:  20000,
		Namespace:   "test",
	}
	remoteOnlineEntityVal, remoteOnlineEntityErr := getObjectValue(onlineEntityPtr)
	if remoteOnlineEntityErr != nil {
		t.Errorf("get remote online entity value failed")
		return
	}

	onlineEntityModelVal, onlineEntityModelErr := remoteProvider.GetEntityModel(remoteOnlineEntityVal)
	if onlineEntityModelErr != nil {
		t.Errorf("get online entity model failed, err:%s", onlineEntityModelErr.Error())
		return
	}
	onlineEntityModelVal, onlineEntityModelErr = o1.Insert(onlineEntityModelVal)
	if onlineEntityModelErr != nil {
		t.Errorf("insert online entity failed, err:%s", onlineEntityModelErr.Error())
		return
	}
	newOnlineEntityPtr := onlineEntityModelVal.Interface(true).(*remote.ObjectValue)
	if len(newOnlineEntityPtr.Fields) != 6 {
		t.Errorf("insert online entity failed")
		return
	}

	queryOnlineEntityPtr := &OnlineEntity{
		SessionID: onlineEntityPtr.SessionID,
		Entity:    &Entity{},
	}

	remoteQueryVal, remoteQueryErr := getObjectValue(queryOnlineEntityPtr)
	if remoteQueryErr != nil {
		t.Errorf("getObjectValue failed, error:%s", remoteQueryErr.Error())
		return
	}

	queryModelVal, queryModelErr := remoteProvider.GetEntityModel(remoteQueryVal)
	if queryModelErr != nil {
		t.Errorf("query online entity failed, error:%v", queryModelErr)
		return
	}

	resultVal, resultErr := o1.Query(queryModelVal)
	if resultErr != nil {
		t.Errorf("query online entity failed, error:%v", resultErr)
		return
	}

	qVal := resultVal.Interface(true).(*remote.ObjectValue)
	if len(qVal.Fields) != 6 {
		t.Errorf("query online entity failed")
	}
}
