package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
)

const composeLocalOwner = "composeLocal"
const composeRemoteOwner = "composeRemote"

func prepareLocalData(localProvider provider.Provider, orm orm.Orm) (sPtr *Simple, rPtr *Reference, cPtr *Compose, err error) {
	ts, _ := time.Parse("2006-01-02 15:04:05", "2018-01-02 15:04:05")
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
	sPtr = sModel.Interface(true).(*Simple)

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
	rPtr = rModel.Interface(true).(*Reference)

	refPtrArray := []*Reference{rPtr}
	cVal := &Compose{
		Name:           strValue,
		Simple:         *sPtr,
		PtrSimple:      sPtr,
		SimpleArray:    []Simple{*sPtr, *sPtr},
		SimplePtrArray: []*Simple{sPtr, sPtr},
		Reference:      *rPtr,
		PtrReference:   rPtr,
		RefArray:       []Reference{*rPtr, *rPtr, *rPtr},
		RefPtrArray:    refPtrArray,
		PtrRefArray:    &refPtrArray,
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
	cPtr = cModel.Interface(true).(*Compose)

	return
}

func prepareRemoteData(remoteProvider provider.Provider, orm orm.Orm) (sPtr *Simple, rPtr *Reference, cPtr *Compose, err error) {
	ts, _ := time.Parse("2006-01-02 15:04:05", "2018-01-02 15:04:05")
	sVal := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}

	sObjectVal, _ := remote.GetObjectValue(sVal)
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
	sObjectVal = sModel.Interface(true).(*remote.ObjectValue)
	sPtr = &Simple{}
	sErr = provider.UpdateEntity(sObjectVal, sPtr)
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

	rObjectVal, _ := remote.GetObjectValue(rVal)
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
	rObjectVal = rModel.Interface(true).(*remote.ObjectValue)
	var fVal float32
	var ts2 time.Time
	var flag2 bool
	strArray2 := []string{}
	ptrStrArray := []*string{}

	rPtr = &Reference{FValue: &fVal, TimeStamp: &ts2, Flag: &flag2, PtrArray: &strArray2, PtrStrArray: &ptrStrArray}
	rErr = provider.UpdateEntity(rObjectVal, rPtr)
	if rErr != nil {
		err = rErr
		return
	}

	refPtrArray := []*Reference{rPtr}
	cVal := &Compose{
		Name:           strValue,
		Simple:         *sPtr,
		PtrSimple:      sPtr,
		SimpleArray:    []Simple{*sPtr, *sPtr},
		SimplePtrArray: []*Simple{sPtr, sPtr},
		Reference:      *rPtr,
		PtrReference:   rPtr,
		RefArray:       []Reference{*rPtr, *rPtr, *rPtr},
		RefPtrArray:    refPtrArray,
		PtrRefArray:    &refPtrArray,
	}
	cObjectVal, _ := remote.GetObjectValue(cVal)
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
	cObjectVal = cModel.Interface(true).(*remote.ObjectValue)
	cPtr = &Compose{
		PtrSimple:      &Simple{},
		SimpleArray:    []Simple{},
		SimplePtrArray: []*Simple{},
		PtrSimpleArray: &[]Simple{},
		PtrReference:   &Reference{},
		RefArray:       []Reference{},
		RefPtrArray:    []*Reference{},
		PtrRefArray:    &[]*Reference{},
		PtrCompose:     &Compose{},
	}
	cErr = provider.UpdateEntity(cObjectVal, cPtr)
	if cErr != nil {
		err = cErr
		return
	}

	return
}

func TestComposeLocal(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialize()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit")
	localProvider := provider.NewLocalProvider(composeLocalOwner)

	o1, err := orm.NewOrm(localProvider, config)
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

	sValList := []*Compose{}
	sModelList := []model.Model{}

	sPtr, rPtr, cPtr, pErr := prepareLocalData(localProvider, o1)
	if pErr != nil {
		t.Errorf("prepareLocalData failed. err:%s", pErr.Error())
		return
	}

	strValue := "test code"
	// insert
	for idx := 0; idx < 100; idx++ {
		refPtrArray := []*Reference{rPtr}
		sVal := &Compose{
			Name:           strValue,
			Simple:         *sPtr,
			PtrSimple:      sPtr,
			SimpleArray:    []Simple{*sPtr, *sPtr},
			SimplePtrArray: []*Simple{sPtr, sPtr},
			Reference:      *rPtr,
			PtrReference:   rPtr,
			RefArray:       []Reference{*rPtr, *rPtr, *rPtr},
			RefPtrArray:    refPtrArray,
			PtrRefArray:    &refPtrArray,
			PtrCompose:     cPtr,
		}
		sValList = append(sValList, sVal)

		sModel, sErr := localProvider.GetEntityModel(sVal)
		if sErr != nil {
			err = sErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		sModelList = append(sModelList, sModel)
	}

	for idx := 0; idx < 100; idx++ {
		vModel, vErr := o1.Insert(sModelList[idx])
		if vErr != nil {
			err = vErr
			t.Errorf("Insert failed. err:%s", err.Error())
			return
		}

		sModelList[idx] = vModel
		sValList[idx] = vModel.Interface(true).(*Compose)
	}

	// update
	for idx := 0; idx < 100; idx++ {
		sVal := sValList[idx]
		sVal.Name = "hi"
		sModel, sErr := localProvider.GetEntityModel(sVal)
		if sErr != nil {
			err = sErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		sModelList[idx] = sModel
	}
	for idx := 0; idx < 100; idx++ {
		vModel, vErr := o1.Update(sModelList[idx])
		if vErr != nil {
			err = vErr
			t.Errorf("Update failed. err:%s", err.Error())
			return
		}

		sModelList[idx] = vModel
		sValList[idx] = vModel.Interface(true).(*Compose)
	}

	// query
	qValList := []*Compose{}
	qModelList := []model.Model{}
	for idx := 0; idx < 100; idx++ {
		qVal := &Compose{
			ID:             sValList[idx].ID,
			PtrSimple:      &Simple{},
			SimpleArray:    []Simple{},
			SimplePtrArray: []*Simple{},
			PtrSimpleArray: &[]Simple{},
			PtrReference:   &Reference{},
			RefArray:       []Reference{},
			RefPtrArray:    []*Reference{},
			PtrRefArray:    &[]*Reference{},
			PtrCompose:     &Compose{},
		}
		qValList = append(qValList, qVal)

		qModel, qErr := localProvider.GetEntityModel(qVal)
		if qErr != nil {
			err = qErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		qModelList = append(qModelList, qModel)
	}

	for idx := 0; idx < 100; idx++ {
		qModel, qErr := o1.Query(qModelList[idx])
		if qErr != nil {
			err = qErr
			t.Errorf("Query failed. err:%s", err.Error())
			return
		}

		qModelList[idx] = qModel
		qValList[idx] = qModel.Interface(true).(*Compose)
	}

	for idx := 0; idx < 100; idx++ {
		sVal := sValList[idx]
		qVal := qValList[idx]
		if !sVal.IsSame(qVal) {
			err = fmt.Errorf("compare value failed")
			t.Errorf("IsSame failed. err:%s", err.Error())
			return
		}
	}

	bqValList := []*Compose{}
	bqModel, bqErr := localProvider.GetEntityModel(&bqValList)
	if bqErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", bqErr.Error())
		return
	}

	filter := orm.GetFilter(localProvider)
	filter.Equal("name", "hi")
	filter.ValueMask(&Compose{
		PtrSimple:      &Simple{},
		SimpleArray:    []Simple{},
		SimplePtrArray: []*Simple{},
		PtrSimpleArray: &[]Simple{},
		PtrReference:   &Reference{},
		RefArray:       []Reference{},
		RefPtrArray:    []*Reference{},
		PtrRefArray:    &[]*Reference{},
		PtrCompose:     &Compose{},
	})
	bqModelList, bqModelErr := o1.BatchQuery(bqModel, filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != 100 {
		t.Errorf("batch query compose failed")
		return
	}

	// delete
	for idx := 0; idx < 100; idx++ {
		_, qErr := o1.Delete(bqModelList[idx])
		if qErr != nil {
			err = qErr
			return
		}
	}
}

func TestComposeRemote(t *testing.T) {

	orm.Initialize()
	defer orm.Uninitialize()
	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit")
	remoteProvider := provider.NewRemoteProvider(composeRemoteOwner)

	o1, err := orm.NewOrm(remoteProvider, config)
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	simpleDef, _ := remote.GetObject(&Simple{})
	referenceDef, _ := remote.GetObject(&Reference{})
	composeDef, _ := remote.GetObject(&Compose{})

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

	sValList := []*Compose{}
	sObjectValList := []*remote.ObjectValue{}
	sModelList := []model.Model{}

	strValue := "test code"
	// insert
	for idx := 0; idx < 100; idx++ {
		refPtrArray := []*Reference{rPtr}
		sVal := &Compose{
			Name:           strValue,
			Simple:         *sPtr,
			PtrSimple:      sPtr,
			SimpleArray:    []Simple{*sPtr, *sPtr},
			SimplePtrArray: []*Simple{sPtr, sPtr},
			Reference:      *rPtr,
			PtrReference:   rPtr,
			RefArray:       []Reference{*rPtr, *rPtr, *rPtr},
			RefPtrArray:    refPtrArray,
			PtrRefArray:    &refPtrArray,
			PtrCompose:     cPtr,
		}
		sValList = append(sValList, sVal)

		sObjectVal, sObjectErr := remote.GetObjectValue(sVal)
		if sObjectErr != nil {
			err = sObjectErr
			t.Errorf("GetObjectValue failed. err:%s", err.Error())
			return
		}
		sObjectValList = append(sObjectValList, sObjectVal)

		sModel, sErr := remoteProvider.GetEntityModel(sObjectVal)
		if sErr != nil {
			err = sErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		sModelList = append(sModelList, sModel)
	}

	for idx := 0; idx < 100; idx++ {
		vModel, vErr := o1.Insert(sModelList[idx])
		if vErr != nil {
			err = vErr
			t.Errorf("Insert failed. err:%s", err.Error())
			return
		}

		sObjectVal := vModel.Interface(true).(*remote.ObjectValue)
		sVal := sValList[idx]
		err = provider.UpdateEntity(sObjectVal, sVal)
		if err != nil {
			t.Errorf("UpdateEntity failed. err:%s", err.Error())
			return
		}
		sValList[idx] = sVal
		sModelList[idx] = vModel
		sObjectValList[idx] = sObjectVal
	}

	// update
	for idx := 0; idx < 100; idx++ {
		sVal := sValList[idx]
		sVal.Name = "hi"
		sObjectVal, sObjectErr := remote.GetObjectValue(sVal)
		if sObjectErr != nil {
			err = sObjectErr
			t.Errorf("GetObjectValue failed. err:%s", err.Error())
			return
		}
		sObjectValList[idx] = sObjectVal

		sModel, sErr := remoteProvider.GetEntityModel(sObjectVal)
		if sErr != nil {
			err = sErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		sModelList[idx] = sModel
	}
	for idx := 0; idx < 100; idx++ {
		vModel, vErr := o1.Update(sModelList[idx])
		if vErr != nil {
			err = vErr
			t.Errorf("Update failed. err:%s", err.Error())
			return
		}

		sObjectVal := vModel.Interface(true).(*remote.ObjectValue)
		sVal := sValList[idx]
		err = provider.UpdateEntity(sObjectVal, sVal)
		if err != nil {
			t.Errorf("UpdateEntity failed. err:%s", err.Error())
			return
		}
		sValList[idx] = sVal
		sModelList[idx] = vModel
		sObjectValList[idx] = sObjectVal
	}

	// query
	qValList := []*Compose{}
	qObjectValList := []*remote.ObjectValue{}
	qModelList := []model.Model{}
	for idx := 0; idx < 100; idx++ {
		qVal := &Compose{
			ID:             sValList[idx].ID,
			PtrSimple:      &Simple{},
			SimpleArray:    []Simple{},
			SimplePtrArray: []*Simple{},
			PtrSimpleArray: &[]Simple{},
			PtrReference:   &Reference{},
			RefArray:       []Reference{},
			RefPtrArray:    []*Reference{},
			PtrRefArray:    &[]*Reference{},
			PtrCompose:     &Compose{},
		}
		qValList = append(qValList, qVal)

		qObjectVal, qObjectErr := remote.GetObjectValue(qVal)
		if qObjectErr != nil {
			err = qObjectErr
			t.Errorf("GetObjectValue failed. err:%s", err.Error())
			return
		}
		qObjectValList = append(qObjectValList, qObjectVal)

		qModel, qErr := remoteProvider.GetEntityModel(qObjectVal)
		if qErr != nil {
			err = qErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		qModelList = append(qModelList, qModel)
	}

	for idx := 0; idx < 100; idx++ {
		qModel, qErr := o1.Query(qModelList[idx])
		if qErr != nil {
			err = qErr
			t.Errorf("Query failed. err:%s", err.Error())
			return
		}

		qObjectVal := qModel.Interface(true).(*remote.ObjectValue)
		qVal := qValList[idx]
		err = provider.UpdateEntity(qObjectVal, qVal)
		if err != nil {
			t.Errorf("UpdateEntity failed. err:%s", err.Error())
			return
		}
		qValList[idx] = qVal
		qModelList[idx] = qModel
		qObjectValList[idx] = qObjectVal
	}

	for idx := 0; idx < 100; idx++ {
		sVal := sValList[idx]
		qVal := qValList[idx]
		if !sVal.IsSame(qVal) {
			err = fmt.Errorf("compare value failed")
			t.Errorf("IsSame failed. err:%s", err.Error())
			return
		}
	}

	bqValList := []*Compose{}
	bqSliceObject, bqSliceErr := remote.GetSliceObjectValue(&bqValList)
	if bqSliceErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", bqSliceErr.Error())
		return
	}
	bqModel, bqErr := remoteProvider.GetEntityModel(bqSliceObject)
	if bqErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", bqErr.Error())
		return
	}

	filter := orm.GetFilter(remoteProvider)
	filter.Equal("name", "hi")
	filter.ValueMask(&Compose{
		PtrSimple:      &Simple{},
		SimpleArray:    []Simple{},
		SimplePtrArray: []*Simple{},
		PtrSimpleArray: &[]Simple{},
		PtrReference:   &Reference{},
		RefArray:       []Reference{},
		RefPtrArray:    []*Reference{},
		PtrRefArray:    &[]*Reference{},
		PtrCompose:     &Compose{},
	})
	bqModelList, bqModelErr := o1.BatchQuery(bqModel, filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != 100 {
		t.Errorf("batch query compose failed")
		return
	}

	// delete
	for idx := 0; idx < 100; idx++ {
		_, qErr := o1.Delete(bqModelList[idx])
		if qErr != nil {
			err = qErr
			return
		}
	}
}
