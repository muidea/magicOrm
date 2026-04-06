package test

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestLocalQuery(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider(localOwner, nil)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Simple{}, &Reference{}, &Compose{}}
	mList, mErr := registerLocalModel(localProvider, objList)
	if mErr != nil {
		t.Errorf("register model failed. err:%s", err.Error())
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

	ts, _ := time.ParseInLocation(util.CSTLayout, "2018-01-02 15:04:05", time.Local)
	s1 := Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}
	s1Model, s1Err := localProvider.GetEntityModel(s1, true)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}

	s1Model, err = o1.Insert(s1Model)
	if err != nil {
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}
	s1 = s1Model.Interface(false).(Simple)

	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []string{strValue, strValue}
	r1 := Reference{
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
	r1Model, r1Err := localProvider.GetEntityModel(r1, true)
	if r1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", r1Err.Error())
		return
	}
	r1Model, err = o1.Insert(r1Model)
	if err != nil {
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	r1 = r1Model.Interface(false).(Reference)

	refPtrArray := []*Reference{&r1}
	c1 := Compose{
		Name:              strValue,
		Simple:            s1,
		SimplePtr:         &s1,
		SimpleArray:       []Simple{s1, s1},
		SimplePtrArray:    []*Simple{&s1, &s1},
		Reference:         r1,
		ReferencePtr:      &r1,
		ReferenceArray:    []Reference{r1, r1, r1},
		ReferencePtrArray: refPtrArray,
	}
	c1Model, c1Err := localProvider.GetEntityModel(c1, true)
	if c1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c1Err.Error())
		return
	}

	c1Model, err = o1.Insert(c1Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	c1 = c1Model.Interface(false).(Compose)

	strValue = "123"
	c2 := Compose{
		Name:              strValue,
		Simple:            s1,
		SimplePtr:         &s1,
		SimpleArray:       []Simple{s1, s1},
		SimplePtrArray:    []*Simple{&s1, &s1},
		Reference:         r1,
		ReferencePtr:      &r1,
		ReferenceArray:    []Reference{r1, r1, r1},
		ReferencePtrArray: refPtrArray,
		ComposePtr:        &c1,
	}
	c2Model, c2Err := localProvider.GetEntityModel(c2, true)
	if c2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c2Err.Error())
		return
	}

	c2Model, err = o1.Insert(c2Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	c2 = c2Model.Interface(false).(Compose)

	c3 := c2
	c3Model, c3Err := localProvider.GetEntityModel(c3, true)
	if c3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c3Err.Error())
		return
	}
	c3Model, err = o1.Insert(c3Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	_ = c3Model.Interface(false).(Compose)

	c4 := c2
	c4Model, c4Err := localProvider.GetEntityModel(c4, true)
	if c4Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c4Err.Error())
		return
	}
	c4Model, err = o1.Insert(c4Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	_ = c4Model.Interface(false).(Compose)

	cModel, _ := localProvider.GetEntityModel(&Compose{}, true)
	filter, err := localProvider.GetModelFilter(cModel)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}
	cModelList, cModelErr := o1.BatchQuery(filter)
	if cModelErr != nil {
		t.Errorf("batch query compose failed, err:%s", cModelErr.Error())
		return
	}
	if len(cModelList) != 4 {
		t.Errorf("batch query compose failed")
		return
	}

	filter.Equal("name", c2.Name)
	filter.ValueMask(&Compose{SimplePtr: &Simple{}})
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

func TestLocalOnlineEntity(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider(localOwner, nil)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Entity{}, &OnlineEntity{}}
	mList, mErr := registerLocalModel(localProvider, objList)
	if mErr != nil {
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

	entityPtr := &Entity{EName: "test", EType: "test", EID: 10, Namespace: "test"}
	entityModelVal, entityModelErr := localProvider.GetEntityModel(entityPtr, true)
	if entityModelErr != nil {
		t.Errorf("get entity model failed, err:%s", entityModelErr.Error())
		return
	}

	entityModelVal, entityModelErr = o1.Insert(entityModelVal)
	if entityModelErr != nil {
		t.Errorf("insert entity model failed, err:%s", entityModelErr.Error())
		return
	}

	newEntityPtr := entityModelVal.Interface(true).(*Entity)
	if newEntityPtr.Namespace != "" || newEntityPtr.EType != "test" || newEntityPtr.EName != "test" {
		t.Errorf("insert entity model failed")
		return
	}

	onlineEntityPtr := &OnlineEntity{
		SessionID:   "abc",
		Entity:      newEntityPtr,
		RefreshTime: 10000,
		ExpireTime:  20000,
		Namespace:   "test",
	}

	onlineEntityModelVal, onlineEntityModelErr := localProvider.GetEntityModel(onlineEntityPtr, true)
	if onlineEntityModelErr != nil {
		t.Errorf("get online entity model failed, err:%s", onlineEntityModelErr.Error())
		return
	}
	onlineEntityModelVal, onlineEntityModelErr = o1.Insert(onlineEntityModelVal)
	if onlineEntityModelErr != nil {
		t.Errorf("insert online entity failed, err:%s", onlineEntityModelErr.Error())
		return
	}
	newOnlineEntityPtr := onlineEntityModelVal.Interface(true).(*OnlineEntity)
	if newOnlineEntityPtr.SessionID != onlineEntityPtr.SessionID || newOnlineEntityPtr.RefreshTime != onlineEntityPtr.RefreshTime || newOnlineEntityPtr.ExpireTime != onlineEntityPtr.ExpireTime {
		t.Errorf("insert online entity failed")
		return
	}

	queryOnlineEntityPtr := &OnlineEntity{
		SessionID: onlineEntityPtr.SessionID,
		Entity:    &Entity{},
	}

	queryModelVal, queryModelErr := localProvider.GetEntityModel(queryOnlineEntityPtr, true)
	if queryModelErr != nil {
		t.Errorf("query online entity failed, error:%v", queryModelErr)
		return
	}

	resultVal, resultErr := o1.Query(queryModelVal)
	if resultErr != nil {
		t.Errorf("query online entity failed, error:%v", resultErr)
		return
	}

	qVal := resultVal.Interface(true).(*OnlineEntity)
	if qVal.ID != newOnlineEntityPtr.ID {
		t.Errorf("query online entity failed")
		return
	}
}
