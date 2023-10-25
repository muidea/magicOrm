package test

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

const referenceLocalOwner = "referenceLocal"
const referenceRemoteOwner = "referenceRemote"

const loop = 1

func TestReferenceLocal(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider(referenceLocalOwner)

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
		t.Errorf("register model failed. err:%s", modelErr.Error())
		return
	}

	model2Err := dropModel(o1, modelList)
	if model2Err != nil {
		t.Errorf("drop model failed. err:%s", model2Err.Error())
		return
	}

	model2Err = createModel(o1, modelList)
	if model2Err != nil {
		t.Errorf("create model failed. err:%s", model2Err.Error())
		return
	}

	sValList := []*Reference{}
	sModelList := []model.Model{}

	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []*string{&strValue, &strValue}

	// insert
	for idx := 0; idx < loop; idx++ {
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

		sModel, sErr := localProvider.GetEntityModel(sVal)
		if sErr != nil {
			t.Errorf("GetEntityModel failed. err:%s", sErr.Error())
			return
		}

		sModelList = append(sModelList, sModel)
	}

	for idx := 0; idx < loop; idx++ {
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
	for idx := 0; idx < loop; idx++ {
		sVal := sValList[idx]
		sVal.Name = "hi"
		sModel, sErr := localProvider.GetEntityModel(sVal)
		if sErr != nil {
			t.Errorf("GetEntityModel failed. err:%s", sErr.Error())
			return
		}

		sModelList[idx] = sModel
	}
	for idx := 0; idx < loop; idx++ {
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
	for idx := 0; idx < loop; idx++ {
		var fVal float32
		var ts time.Time
		var flag bool
		strArray := []string{}
		ptrStrArray := []*string{}

		qVal := &Reference{
			ID:          sValList[idx].ID,
			FValue:      &fVal,
			TimeStamp:   &ts,
			Flag:        &flag,
			IArray:      []int{},
			FArray:      []float32{},
			StrArray:    []string{},
			BArray:      []bool{},
			PtrArray:    &strArray,
			StrPtrArray: []*string{},
			PtrStrArray: &ptrStrArray,
		}
		qValList = append(qValList, qVal)

		qModel, qErr := localProvider.GetEntityModel(qVal)
		if qErr != nil {
			t.Errorf("GetEntityModel failed. err:%s", qErr.Error())
			return
		}

		qModelList = append(qModelList, qModel)
	}

	for idx := 0; idx < loop; idx++ {
		qModel, qErr := o1.Query(qModelList[idx])
		if qErr != nil {
			err = qErr
			t.Errorf("Query failed. err:%s", err.Error())
			return
		}

		qModelList[idx] = qModel
		qValList[idx] = qModel.Interface(true).(*Reference)
	}

	for idx := 0; idx < loop; idx++ {
		sVal := sValList[idx]
		qVal := qValList[idx]
		if !sVal.IsSame(qVal) {
			t.Errorf("IsSame failed. err:%s", "compare value failed")
			return
		}
	}

	var fVal float32
	var ts2 time.Time
	var flag2 bool
	strArray2 := []string{}
	ptrStrArray := []*string{}

	referenceModel, _ := localProvider.GetEntityModel(&Reference{})
	filter, fErr := localProvider.GetModelFilter(referenceModel)
	if fErr != nil {
		t.Errorf("GetEntityFilter failed, err:%s", fErr.Error())
		return
	}

	fErr = filter.Equal("name", "hi")
	if fErr != nil {
		t.Errorf("filter.Equal failed, err:%s", fErr.Error())
		return
	}

	fErr = filter.ValueMask(&Reference{FValue: &fVal, TimeStamp: &ts2, Flag: &flag2, PtrArray: &strArray2, PtrStrArray: &ptrStrArray})
	if fErr != nil {
		t.Errorf("filter.ValueMask failed, err:%s", fErr.Error())
		return
	}

	bqModelList, bqModelErr := o1.BatchQuery(filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != 1 {
		t.Errorf("batch query reference failed")
		return
	}

	// delete
	for idx := 0; idx < loop; idx++ {
		_, qErr := o1.Delete(bqModelList[idx])
		if qErr != nil {
			err = qErr
			return
		}
	}
}

func TestReferenceRemote(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider(referenceRemoteOwner)

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
		t.Errorf("register model failed. err:%s", modelErr.Error())
		return
	}

	model2Err := dropModel(o1, modelList)
	if model2Err != nil {
		t.Errorf("drop model failed. err:%s", model2Err.Error())
		return
	}

	model2Err = createModel(o1, modelList)
	if model2Err != nil {
		t.Errorf("create model failed. err:%s", model2Err.Error())
		return
	}

	sValList := []*Reference{}
	sObjectValList := []*remote.ObjectValue{}
	sModelList := []model.Model{}

	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []*string{&strValue, &strValue}

	// insert
	for idx := 0; idx < loop; idx++ {
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

		sObjectVal, sObjectErr := helper.GetObjectValue(sVal)
		if sObjectErr != nil {
			t.Errorf("GetObjectValue failed. err:%s", sObjectErr.Error())
			return
		}
		sObjectValList = append(sObjectValList, sObjectVal)

		sModel, sErr := remoteProvider.GetEntityModel(sObjectVal)
		if sErr != nil {
			t.Errorf("GetEntityModel failed. err:%s", sErr.Error())
			return
		}

		sModelList = append(sModelList, sModel)
	}

	for idx := 0; idx < loop; idx++ {
		vModel, vErr := o1.Insert(sModelList[idx])
		if vErr != nil {
			err = vErr
			t.Errorf("Insert failed. err:%s", err.Error())
			return
		}

		sObjectVal := vModel.Interface(true).(*remote.ObjectValue)
		sVal := sValList[idx]
		eErr := helper.UpdateEntity(sObjectVal, sVal)
		if eErr != nil {
			t.Errorf("UpdateEntity failed. err:%s", eErr.Error())
			return
		}
		sValList[idx] = sVal
		sModelList[idx] = vModel
		sObjectValList[idx] = sObjectVal
	}

	// update
	for idx := 0; idx < loop; idx++ {
		sVal := sValList[idx]
		sVal.Name = "hi"
		sObjectVal, sObjectErr := helper.GetObjectValue(sVal)
		if sObjectErr != nil {
			t.Errorf("GetObjectValue failed. err:%s", sObjectErr.Error())
			return
		}
		sObjectValList[idx] = sObjectVal

		sModel, sErr := remoteProvider.GetEntityModel(sObjectVal)
		if sErr != nil {
			t.Errorf("GetEntityModel failed. err:%s", sErr.Error())
			return
		}

		sModelList[idx] = sModel
	}
	for idx := 0; idx < loop; idx++ {
		vModel, vErr := o1.Update(sModelList[idx])
		if vErr != nil {
			err = vErr
			t.Errorf("Update failed. err:%s", err.Error())
			return
		}

		sObjectVal := vModel.Interface(true).(*remote.ObjectValue)
		sVal := sValList[idx]
		eErr := helper.UpdateEntity(sObjectVal, sVal)
		if eErr != nil {
			t.Errorf("UpdateEntity failed. err:%s", eErr.Error())
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
	for idx := 0; idx < loop; idx++ {
		var fVal float32
		var ts time.Time
		var flag bool
		strArray := []string{}
		ptrStrArray := []*string{}

		qVal := &Reference{
			ID:          sValList[idx].ID,
			FValue:      &fVal,
			TimeStamp:   &ts,
			Flag:        &flag,
			IArray:      []int{},
			FArray:      []float32{},
			StrArray:    []string{},
			BArray:      []bool{},
			PtrArray:    &strArray,
			StrPtrArray: []*string{},
			PtrStrArray: &ptrStrArray,
		}
		qValList = append(qValList, qVal)

		qObjectVal, qObjectErr := helper.GetObjectValue(qVal)
		if qObjectErr != nil {
			t.Errorf("GetObjectValue failed. err:%s", qObjectErr.Error())
			return
		}
		qObjectValList = append(qObjectValList, qObjectVal)

		qModel, qErr := remoteProvider.GetEntityModel(qObjectVal)
		if qErr != nil {
			t.Errorf("GetEntityModel failed. err:%s", qErr.Error())
			return
		}

		qModelList = append(qModelList, qModel)
	}

	for idx := 0; idx < loop; idx++ {
		qModel, qErr := o1.Query(qModelList[idx])
		if qErr != nil {
			err = qErr
			t.Errorf("Query failed. err:%s", err.Error())
			return
		}

		qObjectVal := qModel.Interface(true).(*remote.ObjectValue)
		qVal := qValList[idx]
		eErr := helper.UpdateEntity(qObjectVal, qVal)
		if eErr != nil {
			t.Errorf("UpdateEntity failed. err:%s", eErr.Error())
			return
		}
		qValList[idx] = qVal
		qModelList[idx] = qModel
		qObjectValList[idx] = qObjectVal
	}

	for idx := 0; idx < loop; idx++ {
		sVal := sValList[idx]
		qVal := qValList[idx]
		if !sVal.IsSame(qVal) {
			t.Errorf("IsSame failed. err:%s", "compare value failed")
			return
		}
	}

	bqValList := []*Reference{}

	var fVal float32
	var ts2 time.Time
	var flag2 bool
	strArray2 := []string{}
	ptrStrArray := []*string{}

	referenceModel, _ := helper.GetObject(&bqValList)

	filter, fErr := remoteProvider.GetModelFilter(referenceModel)
	if fErr != nil {
		t.Errorf("GetEntityFilter failed, err:%s", fErr.Error())
		return
	}

	fErr = filter.Equal("name", "hi")
	if fErr != nil {
		t.Errorf("filter.Equal failed, err:%s", fErr.Error())
		return
	}

	maskVal, maskErr := helper.GetObjectValue(&Reference{FValue: &fVal, TimeStamp: &ts2, Flag: &flag2, PtrArray: &strArray2, PtrStrArray: &ptrStrArray})
	if maskErr != nil {
		t.Errorf("helper.GetObjectValue failed, err:%s", maskErr.Error())
		return
	}

	fErr = filter.ValueMask(maskVal)
	if fErr != nil {
		t.Errorf("filter.ValueMask failed, err:%s", fErr.Error())
		return
	}

	fErr = filter.Like("strArray", "Abc")
	if fErr != nil {
		t.Errorf("filter.Like failed, err:%s", fErr.Error())
		return
	}

	bqModelList, bqModelErr := o1.BatchQuery(filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != loop {
		t.Errorf("batch query reference failed")
		return
	}

	// delete
	for idx := 0; idx < loop; idx++ {
		_, qErr := o1.Delete(bqModelList[idx])
		if qErr != nil {
			err = qErr
			return
		}
	}
}
