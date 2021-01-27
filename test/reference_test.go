package test

import (
	"fmt"
	orm "github.com/muidea/magicOrm"
	"github.com/muidea/magicOrm/model"
	ormProvider "github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
	"testing"
	"time"
)

const referenceLocalOwner = "referenceLocal"
const referenceRemoteOwner = "referenceRemote"

func TestReferenceLocal(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	o1, err := orm.NewOrm(referenceLocalOwner)
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	provider := orm.GetProvider(referenceLocalOwner)

	simpleDef := &Simple{}
	referenceDef := &Reference{}
	composeDef := &Compose{}

	entityList := []interface{}{simpleDef, referenceDef, composeDef}
	modelList, modelErr := registerModel(provider, entityList)
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

	sValList := []*Reference{}
	sModelList := []model.Model{}

	ts, _ := time.Parse("2006-01-02 15:04:05", "2018-01-02 15:04:05")
	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []*string{&strValue, &strValue}

	// insert
	for idx := 0; idx < 100; idx++ {
		sVal := &Reference{
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
		sValList = append(sValList, sVal)

		sModel, sErr := provider.GetEntityModel(sVal)
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
		sValList[idx] = vModel.Interface(true).(*Reference)
	}

	// update
	for idx := 0; idx < 100; idx++ {
		sVal := sValList[idx]
		sVal.Name = "hi"
		sModel, sErr := provider.GetEntityModel(sVal)
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
		sValList[idx] = vModel.Interface(true).(*Reference)
	}

	// query
	qValList := []*Reference{}
	qModelList := []model.Model{}
	for idx := 0; idx < 100; idx++ {
		var fVal float32
		var ts time.Time
		var flag bool
		strArray := []string{}
		ptrStrArray := []*string{}

		qVal := &Reference{ID: sValList[idx].ID, FValue: &fVal, TimeStamp: &ts, Flag: &flag, PtrArray: &strArray, PtrStrArray: &ptrStrArray}
		qValList = append(qValList, qVal)

		qModel, qErr := provider.GetEntityModel(qVal)
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
		qValList[idx] = qModel.Interface(true).(*Reference)
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

	bqValList := []*Reference{}
	bqModel, bqErr := provider.GetEntityModel(&bqValList)
	if bqErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", bqErr.Error())
		return
	}

	var fVal float32
	var ts2 time.Time
	var flag2 bool
	strArray2 := []string{}
	ptrStrArray := []*string{}

	filter := orm.GetFilter(referenceLocalOwner)
	filter.Equal("Name", "hi")
	filter.ValueMask(&Reference{FValue: &fVal, TimeStamp: &ts2, Flag: &flag2, PtrArray: &strArray2, PtrStrArray: &ptrStrArray})
	bqModelList, bqModelErr := o1.BatchQuery(bqModel, filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != 100 {
		t.Errorf("batch query reference failed")
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

func TestReferenceRemote(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	o1, err := orm.NewOrm(referenceRemoteOwner)
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	provider := orm.GetProvider(referenceRemoteOwner)

	simpleDef, _ := remote.GetObject(&Simple{})
	referenceDef, _ := remote.GetObject(&Reference{})
	composeDef, _ := remote.GetObject(&Compose{})

	entityList := []interface{}{simpleDef, referenceDef, composeDef}
	modelList, modelErr := registerModel(provider, entityList)
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

	sValList := []*Reference{}
	sObjectValList := []*remote.ObjectValue{}
	sModelList := []model.Model{}

	ts, _ := time.Parse("2006-01-02 15:04:05", "2018-01-02 15:04:05")
	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []*string{&strValue, &strValue}

	// insert
	for idx := 0; idx < 100; idx++ {
		sVal := &Reference{
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
		sValList = append(sValList, sVal)

		sObjectVal, sObjectErr := remote.GetObjectValue(sVal)
		if sObjectErr != nil {
			err = sObjectErr
			t.Errorf("GetObjectValue failed. err:%s", err.Error())
			return
		}
		sObjectValList = append(sObjectValList, sObjectVal)

		sModel, sErr := provider.GetEntityModel(sObjectVal)
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
		err = ormProvider.UpdateEntity(sObjectVal, sVal)
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

		sModel, sErr := provider.GetEntityModel(sObjectVal)
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
		err = ormProvider.UpdateEntity(sObjectVal, sVal)
		if err != nil {
			t.Errorf("UpdateEntity failed. err:%s", err.Error())
			return
		}
		sValList[idx] = sVal
		sModelList[idx] = vModel
		sObjectValList[idx] = sObjectVal
	}

	// query
	qValList := []*Reference{}
	qObjectValList := []*remote.ObjectValue{}
	qModelList := []model.Model{}
	for idx := 0; idx < 100; idx++ {
		var fVal float32
		var ts time.Time
		var flag bool
		strArray := []string{}
		ptrStrArray := []*string{}

		qVal := &Reference{ID: sValList[idx].ID, FValue: &fVal, TimeStamp: &ts, Flag: &flag, PtrArray: &strArray, PtrStrArray: &ptrStrArray}
		qValList = append(qValList, qVal)

		qObjectVal, qObjectErr := remote.GetObjectValue(qVal)
		if qObjectErr != nil {
			err = qObjectErr
			t.Errorf("GetObjectValue failed. err:%s", err.Error())
			return
		}
		qObjectValList = append(qObjectValList, qObjectVal)

		qModel, qErr := provider.GetEntityModel(qObjectVal)
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
		err = ormProvider.UpdateEntity(qObjectVal, qVal)
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

	bqValList := []*Reference{}
	bqSliceObject, bqSliceErr := remote.GetSliceObjectValue(&bqValList)
	if bqSliceErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", bqSliceErr.Error())
		return
	}
	bqModel, bqErr := provider.GetEntityModel(bqSliceObject)
	if bqErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", bqErr.Error())
		return
	}

	var fVal float32
	var ts2 time.Time
	var flag2 bool
	strArray2 := []string{}
	ptrStrArray := []*string{}

	filter := orm.GetFilter(referenceLocalOwner)
	filter.Equal("Name", "hi")
	filter.ValueMask(&Reference{FValue: &fVal, TimeStamp: &ts2, Flag: &flag2, PtrArray: &strArray2, PtrStrArray: &ptrStrArray})
	bqModelList, bqModelErr := o1.BatchQuery(bqModel, filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != 100 {
		t.Errorf("batch query reference failed")
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
