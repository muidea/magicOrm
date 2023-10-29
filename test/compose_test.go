package test

import (
	"fmt"
	cd "github.com/muidea/magicCommon/def"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

const composeLocalOwner = "composeLocal"
const composeRemoteOwner = "composeRemote"

func prepareLocalData(localProvider provider.Provider, orm orm.Orm) (sPtr *Simple, rPtr *Reference, cPtr *Compose, err *cd.Result) {
	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	sVal := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}

	sModel, sErr := localProvider.GetEntityModel(sVal)
	if sErr != nil {
		err = sErr
		return
	}

	sModel, sErr = orm.Insert(sModel)
	if sErr != nil {
		err = sErr
		return
	}
	sPtr = sModel.Interface(true, 0).(*Simple)

	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []*string{&strValue, &strValue}
	rVal := &Reference{
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

	rModel, rErr := localProvider.GetEntityModel(rVal)
	if rErr != nil {
		err = rErr
		return
	}

	rModel, rErr = orm.Insert(rModel)
	if rErr != nil {
		err = rErr
		return
	}
	rPtr = rModel.Interface(true, 0).(*Reference)

	refPtrArray := []*Reference{rPtr}
	cVal := &Compose{
		Name:         strValue,
		H1:           *sPtr,
		R3:           sPtr,
		H2:           []Simple{*sPtr, *sPtr},
		R4:           []*Simple{sPtr, sPtr},
		Reference:    *rPtr,
		PtrReference: rPtr,
		RefArray:     []Reference{*rPtr, *rPtr, *rPtr},
		RefPtrArray:  refPtrArray,
		PtrRefArray:  refPtrArray,
	}
	cModel, cErr := localProvider.GetEntityModel(cVal)
	if cErr != nil {
		err = cErr
		return
	}

	cModel, cErr = orm.Insert(cModel)
	if cErr != nil {
		err = cErr
		return
	}
	cPtr = cModel.Interface(true, 0).(*Compose)

	return
}

func prepareRemoteData(remoteProvider provider.Provider, orm orm.Orm) (sPtr *Simple, rPtr *Reference, cPtr *Compose, err *cd.Result) {
	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	sVal := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}

	sObjectVal, _ := helper.GetObjectValue(sVal)
	sModel, sErr := remoteProvider.GetEntityModel(sObjectVal)
	if sErr != nil {
		err = sErr
		return
	}

	sModel, sErr = orm.Insert(sModel)
	if sErr != nil {
		err = sErr
		return
	}
	sObjectVal = sModel.Interface(true, 0).(*remote.ObjectValue)
	sPtr = &Simple{}
	sErr = helper.UpdateEntity(sObjectVal, sPtr)
	if sErr != nil {
		err = sErr
		return
	}

	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []*string{&strValue, &strValue}
	rVal := &Reference{
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

	rObjectVal, _ := helper.GetObjectValue(rVal)
	rModel, rErr := remoteProvider.GetEntityModel(rObjectVal)
	if rErr != nil {
		err = rErr
		return
	}

	rModel, rErr = orm.Insert(rModel)
	if rErr != nil {
		err = rErr
		return
	}
	rObjectVal = rModel.Interface(true, 0).(*remote.ObjectValue)
	var fVal float32
	var ts2 time.Time
	var flag2 bool
	strArray2 := []string{}
	ptrStrArray := []*string{}

	rPtr = &Reference{FValue: &fVal, TimeStamp: &ts2, Flag: &flag2, PtrArray: &strArray2, PtrStrArray: &ptrStrArray}
	rErr = helper.UpdateEntity(rObjectVal, rPtr)
	if rErr != nil {
		err = rErr
		return
	}

	refPtrArray := []*Reference{rPtr}
	cVal := &Compose{
		Name:         strValue,
		H1:           *sPtr,
		R3:           sPtr,
		H2:           []Simple{*sPtr, *sPtr},
		R4:           []*Simple{sPtr, sPtr},
		Reference:    *rPtr,
		PtrReference: rPtr,
		RefArray:     []Reference{*rPtr, *rPtr, *rPtr},
		RefPtrArray:  refPtrArray,
		PtrRefArray:  refPtrArray,
	}
	cObjectVal, _ := helper.GetObjectValue(cVal)
	cModel, cErr := remoteProvider.GetEntityModel(cObjectVal)
	if cErr != nil {
		err = cErr
		return
	}

	cModel, cErr = orm.Insert(cModel)
	if cErr != nil {
		err = cErr
		return
	}
	cObjectVal = cModel.Interface(true, 0).(*remote.ObjectValue)
	cPtr = &Compose{
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
	cErr = helper.UpdateEntity(cObjectVal, cPtr)
	if cErr != nil {
		err = cErr
		return
	}

	return
}

func TestComposeLocal(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider(composeLocalOwner)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	simpleDef := &Simple{}
	referenceDef := &Reference{}
	composeDef := &Compose{}

	entityList := []interface{}{simpleDef, referenceDef, composeDef}
	modelList, modelErr := registerModel(localProvider, entityList)
	if modelErr != nil {
		err = modelErr
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	err = dropModel(o1, modelList)
	if err != nil {
		t.Errorf("drop model failed. err:%s", err.Error())
		return
	}

	err = createModel(o1, modelList)
	if err != nil {
		t.Errorf("create model failed. err:%s", err.Error())
		return
	}

	sPtr, rPtr, cPtr, pErr := prepareLocalData(localProvider, o1)
	if pErr != nil {
		t.Errorf("prepareLocalData failed. err:%s", pErr.Error())
		return
	}

	strValue := "test code"
	// insert
	refPtrArray := []*Reference{rPtr}
	composePtr := &Compose{
		Name:         strValue,
		H1:           *sPtr,
		R3:           sPtr,
		H2:           []Simple{*sPtr, *sPtr},
		R4:           []*Simple{sPtr, sPtr},
		Reference:    *rPtr,
		PtrReference: rPtr,
		RefArray:     []Reference{*rPtr, *rPtr, *rPtr},
		RefPtrArray:  refPtrArray,
		PtrRefArray:  refPtrArray,
		PtrCompose:   cPtr,
	}

	composeModel, composeErr := localProvider.GetEntityModel(composePtr)
	if composeErr != nil {
		err = composeErr
		t.Errorf("GetEntityModel failed. err:%s", err.Error())
		return
	}

	composeModel, composeErr = o1.Insert(composeModel)
	if composeErr != nil {
		err = composeErr
		t.Errorf("Insert failed. err:%s", err.Error())
		return
	}

	// update
	composePtr = composeModel.Interface(true, 0).(*Compose)
	composePtr.Name = "hi"
	composeModel, composeErr = localProvider.GetEntityModel(composePtr)
	if composeErr != nil {
		err = composeErr
		t.Errorf("GetEntityModel failed. err:%s", err.Error())
		return
	}

	composeModel, composeErr = o1.Update(composeModel)
	if composeErr != nil {
		err = composeErr
		t.Errorf("Update failed. err:%s", err.Error())
		return
	}

	composePtr = composeModel.Interface(true, 0).(*Compose)

	// query
	queryVal := &Compose{
		ID:           composePtr.ID,
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

	queryModel, queryErr := localProvider.GetEntityModel(queryVal)
	if queryErr != nil {
		err = queryErr
		t.Errorf("GetEntityModel failed. err:%s", err.Error())
		return
	}

	queryModel, queryErr = o1.Query(queryModel)
	if queryErr != nil {
		err = queryErr
		t.Errorf("Query failed, err:%s", err.Error())
		return
	}
	queryVal = queryModel.Interface(true, 0).(*Compose)

	if !composePtr.IsSame(queryVal) {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("compare value failed"))
		t.Errorf("IsSame failed. err:%s", err.Error())
		return
	}

	cModel, _ := localProvider.GetEntityModel(&Compose{})
	filter, err := localProvider.GetModelFilter(cModel)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	err = filter.Equal("name", "hi")
	if err != nil {
		t.Errorf("filter.Equal failed, err:%s", err.Error())
		return
	}
	err = filter.ValueMask(&Compose{
		R3:           &Simple{},
		H2:           []Simple{},
		R4:           []*Simple{},
		PR4:          &[]Simple{},
		PtrReference: &Reference{},
		RefArray:     []Reference{},
		RefPtrArray:  []*Reference{},
		PtrRefArray:  []*Reference{},
		PtrCompose:   &Compose{},
	})
	if err != nil {
		t.Errorf("filter.ValueMask failed, err:%s", err.Error())
		return
	}

	bqModelList, bqModelErr := o1.BatchQuery(filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != 1 {
		t.Errorf("batch query compose failed")
		return
	}

	// delete
	_, qErr := o1.Delete(bqModelList[0])
	if qErr != nil {
		err = qErr
		return
	}
}

func TestComposeRemote(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()
	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider(composeRemoteOwner)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	simpleDef, _ := helper.GetObject(&Simple{})
	referenceDef, _ := helper.GetObject(&Reference{})
	composeDef, _ := helper.GetObject(&Compose{})

	entityList := []interface{}{simpleDef, referenceDef, composeDef}
	modelList, modelErr := registerModel(remoteProvider, entityList)
	if modelErr != nil {
		err = modelErr
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	err = dropModel(o1, modelList)
	if err != nil {
		t.Errorf("drop model failed. err:%s", err.Error())
		return
	}

	err = createModel(o1, modelList)
	if err != nil {
		t.Errorf("create model failed. err:%s", err.Error())
		return
	}

	sPtr, rPtr, cPtr, pErr := prepareRemoteData(remoteProvider, o1)
	if pErr != nil {
		t.Errorf("prepareRemoteData failed. err:%s", pErr.Error())
		return
	}

	strValue := "test code"
	// insert
	refPtrArray := []*Reference{rPtr}
	composePtr := &Compose{
		Name:         strValue,
		H1:           *sPtr,
		R3:           sPtr,
		H2:           []Simple{*sPtr, *sPtr},
		R4:           []*Simple{sPtr, sPtr},
		Reference:    *rPtr,
		PtrReference: rPtr,
		RefArray:     []Reference{*rPtr, *rPtr, *rPtr},
		RefPtrArray:  refPtrArray,
		PtrRefArray:  refPtrArray,
		PtrCompose:   cPtr,
	}
	composeObjectValue, composeObjectErr := helper.GetObjectValue(composePtr)
	if composeObjectErr != nil {
		err = composeObjectErr
		t.Errorf("GetObjectValue failed. err:%s", err.Error())
		return
	}

	composeModel, composeErr := remoteProvider.GetEntityModel(composeObjectValue)
	if composeErr != nil {
		err = composeErr
		t.Errorf("GetEntityModel failed. err:%s", err.Error())
		return
	}

	composeModel, composeErr = o1.Insert(composeModel)
	if composeErr != nil {
		err = composeErr
		t.Errorf("Insert failed. err:%s", err.Error())
		return
	}

	composeObjectValue = composeModel.Interface(true, 0).(*remote.ObjectValue)
	err = helper.UpdateEntity(composeObjectValue, composePtr)
	if err != nil {
		t.Errorf("UpdateEntity failed. err:%s", err.Error())
		return
	}

	composePtr.Name = "hi"
	composeObjectValue, composeObjectErr = helper.GetObjectValue(composePtr)
	if composeObjectErr != nil {
		err = composeObjectErr
		t.Errorf("GetObjectValue failed. err:%s", err.Error())
		return
	}

	composeModel, composeErr = remoteProvider.GetEntityModel(composeObjectValue)
	if composeErr != nil {
		err = composeErr
		t.Errorf("GetEntityModel failed. err:%s", err.Error())
		return
	}

	vModel, vErr := o1.Update(composeModel)
	if vErr != nil {
		err = vErr
		t.Errorf("Update failed. err:%s", err.Error())
		return
	}

	composeObjectValue = vModel.Interface(true, 0).(*remote.ObjectValue)
	err = helper.UpdateEntity(composeObjectValue, composePtr)
	if err != nil {
		t.Errorf("UpdateEntity failed. err:%s", err.Error())
		return
	}

	// query
	queryComposeVal := &Compose{
		ID:           composePtr.ID,
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

	queryObjectValue, queryObjectErr := helper.GetObjectValue(queryComposeVal)
	if queryObjectErr != nil {
		err = queryObjectErr
		t.Errorf("GetObjectValue failed. err:%s", err.Error())
		return
	}

	queryModel, queryErr := remoteProvider.GetEntityModel(queryObjectValue)
	if queryErr != nil {
		err = queryErr
		t.Errorf("GetEntityModel failed. err:%s", err.Error())
		return
	}

	queryModel, queryErr = o1.Query(queryModel)
	if queryErr != nil {
		err = queryErr
		t.Errorf("Query failed. err:%s", err.Error())
		return
	}

	queryObjectVal := queryModel.Interface(true, 0).(*remote.ObjectValue)
	err = helper.UpdateEntity(queryObjectVal, queryComposeVal)
	if err != nil {
		t.Errorf("UpdateEntity failed. err:%s", err.Error())
		return
	}

	if !composePtr.IsSame(queryComposeVal) {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("compare value failed"))
		t.Errorf("IsSame failed. err:%s", err.Error())
		return
	}

	composePtr = &Compose{}
	composeObjectPtr, composeObjectErr := helper.GetObject(composePtr)
	if composeObjectErr != nil {
		t.Errorf("GetObject failed, err:%s", composeObjectErr.Error())
		return
	}

	filter, err := remoteProvider.GetModelFilter(composeObjectPtr)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	err = filter.Equal("name", "hi")
	if err != nil {
		t.Errorf("filter.Equal failed, err:%s", err.Error())
		return
	}

	maskComposePtr := &Compose{
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

	maskObjectValuePtr, maskObjectValueErr := helper.GetObjectValue(maskComposePtr)
	if maskObjectValueErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", maskObjectValueErr.Error())
		return
	}

	err = filter.ValueMask(maskObjectValuePtr)
	if err != nil {
		t.Errorf("filter.ValueMask failed, err:%s", err.Error())
		return
	}

	bqModelList, bqModelErr := o1.BatchQuery(filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != 1 {
		t.Errorf("batch query compose failed")
		return
	}

	// delete
	_, qErr := o1.Delete(bqModelList[0])
	if qErr != nil {
		err = qErr
		t.Errorf("o1.Delete failed, err:%s", qErr.Error())
		return
	}
}
