package test

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

const localOwner = "local"

func TestLocalSimple(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider(localOwner)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Simple{}}
	_, err = registerModel(localProvider, objList)
	if err != nil {
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	s1 := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}

	s1Model, s1Err := localProvider.GetEntityModel(s1)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}

	err = o1.Drop(s1Model)
	if err != nil {
		t.Errorf("drop simple schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(s1Model)
	if err != nil {
		t.Errorf("create simple schema failed, err:%s", err.Error())
		return
	}

	s1Model, err = o1.Insert(s1Model)
	if err != nil {
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}
	s1 = s1Model.Interface(true, 0).(*Simple)

	s1.Name = "hello"
	s1Model, s1Err = localProvider.GetEntityModel(s1)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	s1Model, err = o1.Update(s1Model)
	if err != nil {
		t.Errorf("update simple failed, err:%s", err.Error())
		return
	}
	s1 = s1Model.Interface(true, 0).(*Simple)

	s2 := Simple{ID: s1.ID}
	s2Model, s2Err := localProvider.GetEntityModel(s2)
	if s2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s2Err.Error())
		return
	}

	s2Model, err = o1.Query(s2Model)
	if err != nil {
		t.Errorf("query simple failed, err:%s", err.Error())
		return
	}
	s2 = s2Model.Interface(false, 0).(Simple)

	if !s1.IsSame(&s2) {
		t.Errorf("Query simple failed.")
	}
}

func TestLocalReference(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider(localOwner)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Reference{}}
	_, err = registerModel(localProvider, objList)
	if err != nil {
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []*string{&strValue, &strValue}
	s1 := &Reference{
		Name:        strValue,
		FValue:      &fValue,
		F64:         23.456,
		TimeStamp:   &ts,
		Flag:        &flag,
		IArray:      iArray,
		FArray:      fArray,
		StrArray:    strArray,
		BArray:      bArray,
		PtrArray:    &strArray,
		StrPtrArray: strPtrArray,
		PtrStrArray: &strPtrArray,
	}

	s1Model, s1Err := localProvider.GetEntityModel(s1)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	err = o1.Drop(s1Model)
	if err != nil {
		t.Errorf("drop simple schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(s1Model)
	if err != nil {
		t.Errorf("create simple schema failed, err:%s", err.Error())
		return
	}

	s1Model, err = o1.Insert(s1Model)
	if err != nil {
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}
	s1 = s1Model.Interface(true, 0).(*Reference)

	s1.Name = "hello"
	s1Model, s1Err = localProvider.GetEntityModel(s1)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}

	s1Model, err = o1.Update(s1Model)
	if err != nil {
		t.Errorf("update simple failed, err:%s", err.Error())
		return
	}
	s1 = s1Model.Interface(true, 0).(*Reference)

	fValue2 := float32(0.0)
	var ts2 time.Time
	var bVal bool
	var strArray2 []string
	var ptrArray2 []*string
	s2 := Reference{
		ID:          s1.ID,
		FValue:      &fValue2,
		TimeStamp:   &ts2,
		Flag:        &bVal,
		IArray:      []int{},
		FArray:      []float32{},
		StrArray:    []string{},
		BArray:      []bool{},
		PtrArray:    &strArray2,
		StrPtrArray: []*string{},
		PtrStrArray: &ptrArray2,
	}

	s2Model, s2Err := localProvider.GetEntityModel(s2)
	if s2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s2Err.Error())
		return
	}

	s2Model, err = o1.Query(s2Model)
	if err != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	s2 = s2Model.Interface(false, 0).(Reference)

	if !s1.IsSame(&s2) {
		t.Errorf("Query reference failed.")
		return
	}

	s2Model, err = o1.Insert(s2Model)
	if err != nil {
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	s2 = s2Model.Interface(false, 0).(Reference)
	if s1.IsSame(&s2) {
		t.Errorf("Query reference failed.")
		return
	}

	s4 := Reference{
		ID: s1.ID,
	}
	s4Model, s4Err := localProvider.GetEntityModel(s4)
	if s4Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s4Err.Error())
		return
	}

	s4Model, err = o1.Query(s4Model)
	if err != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	s4 = s4Model.Interface(false, 0).(Reference)
	if s4.Name != s2.Name {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	if s4.FValue != nil || s4.TimeStamp != nil || s4.Flag != nil || s4.PtrStrArray != nil || s4.PtrArray != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
}

func TestLocalCompose(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider(localOwner)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Simple{}, &Reference{}, &Compose{}}
	mList, mErr := registerModel(localProvider, objList)
	if mErr != nil {
		err = mErr
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

	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	s1 := Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}
	s1Model, s1Err := localProvider.GetEntityModel(s1)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	s1Model, err = o1.Insert(s1Model)
	if err != nil {
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}
	s1 = s1Model.Interface(false, 0).(Simple)

	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []*string{&strValue, &strValue}
	r1 := Reference{
		Name:        strValue,
		FValue:      &fValue,
		F64:         23.456,
		TimeStamp:   &ts,
		Flag:        &flag,
		IArray:      iArray,
		FArray:      fArray,
		StrArray:    strArray,
		BArray:      bArray,
		PtrArray:    &strArray,
		StrPtrArray: strPtrArray,
		PtrStrArray: &strPtrArray,
	}

	r1Model, r1Err := localProvider.GetEntityModel(r1)
	if r1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", r1Err.Error())
		return
	}

	r1Model, err = o1.Insert(r1Model)
	if err != nil {
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}

	r1 = r1Model.Interface(false, 0).(Reference)

	refPtrArray := []*Reference{&r1}
	c1 := &Compose{
		Name:         strValue,
		H1:           s1,
		R3:           &s1,
		H2:           []Simple{s1, s1},
		R4:           []*Simple{&s1, &s1},
		Reference:    r1,
		PtrReference: &r1,
		RefArray:     []Reference{r1, r1, r1},
		RefPtrArray:  refPtrArray,
		PtrRefArray:  refPtrArray,
	}
	c1Model, c1Err := localProvider.GetEntityModel(c1)
	if c1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c1Err.Error())
		return
	}
	c1Model, err = o1.Insert(c1Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	c1 = c1Model.Interface(true, 0).(*Compose)

	c2 := Compose{
		Name:         strValue,
		H1:           s1,
		R3:           &s1,
		H2:           []Simple{s1, s1},
		R4:           []*Simple{&s1, &s1},
		Reference:    r1,
		PtrReference: &r1,
		RefArray:     []Reference{r1, r1, r1},
		RefPtrArray:  refPtrArray,
		PtrRefArray:  refPtrArray,
		PtrCompose:   c1,
	}
	c2Model, c2Err := localProvider.GetEntityModel(c2)
	if c2Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", c2Err.Error())
		return
	}

	c2Model, err = o1.Insert(c2Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	c2 = c2Model.Interface(false, 0).(Compose)

	c3 := Compose{
		ID:           c2.ID,
		R3:           &Simple{},
		H2:           []Simple{},
		R4:           []*Simple{},
		PR4:          &[]Simple{},
		PtrReference: &Reference{},
		RefArray:     []Reference{},
		RefPtrArray:  []*Reference{},
		PtrRefArray:  []*Reference{},
		PtrCompose:   &Compose{},
	}
	c3Model, c3Err := localProvider.GetEntityModel(c3)
	if c3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c3Err.Error())
		return
	}

	c3Model, err = o1.Query(c3Model)
	if err != nil {
		t.Errorf("query compose failed, err:%s", err.Error())
		return
	}
	c3 = c3Model.Interface(false, 0).(Compose)

	if c3.IsSame(c1) {
		t.Error("query compose failed")
		return
	}
	if !c3.IsSame(&c2) {
		t.Error("query compose failed")
		return
	}
}

func TestLocalQuery(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider(localOwner)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Simple{}, &Reference{}, &Compose{}}
	mList, mErr := registerModel(localProvider, objList)
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

	ts, _ := time.ParseInLocation(util.CSTLayout, "2018-01-02 15:04:05", time.Local)
	s1 := Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}
	s1Model, s1Err := localProvider.GetEntityModel(s1)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}

	s1Model, err = o1.Insert(s1Model)
	if err != nil {
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}
	s1 = s1Model.Interface(false, 0).(Simple)

	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []*string{&strValue, &strValue}
	r1 := Reference{
		Name:        strValue,
		FValue:      &fValue,
		F64:         23.456,
		TimeStamp:   &ts,
		Flag:        &flag,
		IArray:      iArray,
		FArray:      fArray,
		StrArray:    strArray,
		BArray:      bArray,
		PtrArray:    &strArray,
		StrPtrArray: strPtrArray,
		PtrStrArray: &strPtrArray,
	}
	r1Model, r1Err := localProvider.GetEntityModel(r1)
	if r1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", r1Err.Error())
		return
	}
	r1Model, err = o1.Insert(r1Model)
	if err != nil {
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	r1 = r1Model.Interface(false, 0).(Reference)

	refPtrArray := []*Reference{&r1}
	c1 := Compose{
		Name:         strValue,
		H1:           s1,
		R3:           &s1,
		H2:           []Simple{s1, s1},
		R4:           []*Simple{&s1, &s1},
		Reference:    r1,
		PtrReference: &r1,
		RefArray:     []Reference{r1, r1, r1},
		RefPtrArray:  refPtrArray,
		PtrRefArray:  refPtrArray,
	}
	c1Model, c1Err := localProvider.GetEntityModel(c1)
	if c1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c1Err.Error())
		return
	}

	c1Model, err = o1.Insert(c1Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	c1 = c1Model.Interface(false, 0).(Compose)

	strValue = "123"
	c2 := Compose{
		Name:         strValue,
		H1:           s1,
		R3:           &s1,
		H2:           []Simple{s1, s1},
		R4:           []*Simple{&s1, &s1},
		Reference:    r1,
		PtrReference: &r1,
		RefArray:     []Reference{r1, r1, r1},
		RefPtrArray:  refPtrArray,
		PtrRefArray:  refPtrArray,
		PtrCompose:   &c1,
	}
	c2Model, c2Err := localProvider.GetEntityModel(c2)
	if c2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c2Err.Error())
		return
	}

	c2Model, err = o1.Insert(c2Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	c2 = c2Model.Interface(false, 0).(Compose)

	c3 := c2
	c3Model, c3Err := localProvider.GetEntityModel(c3)
	if c3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c3Err.Error())
		return
	}
	c3Model, err = o1.Insert(c3Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	c3 = c3Model.Interface(false, 0).(Compose)

	c4 := c2
	c4Model, c4Err := localProvider.GetEntityModel(c4)
	if c4Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", c4Err.Error())
		return
	}
	c4Model, err = o1.Insert(c4Model)
	if err != nil {
		t.Errorf("insert compose failed, err:%s", err.Error())
		return
	}
	c4 = c4Model.Interface(false, 0).(Compose)

	cModel, _ := localProvider.GetEntityModel(&Compose{})
	filter, err := localProvider.GetModelFilter(cModel, 0)
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
	filter.ValueMask(&Compose{R3: &Simple{}})
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
