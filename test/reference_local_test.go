package test

import (
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

const referenceLocalOwner = "referenceLocal"

const localLoop = 1

func TestReferenceLocal(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider(referenceLocalOwner, nil)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	simpleDef := &Simple{}
	referenceDef := &Reference{}
	composeDef := &Compose{}

	entityList := []any{simpleDef, referenceDef, composeDef}
	modelList, modelErr := registerLocalModel(localProvider, entityList)
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
	sModelList := []models.Model{}

	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []string{strValue, strValue}

	// insert
	for idx := 0; idx < localLoop; idx++ {
		sVal := &Reference{
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
		sValList = append(sValList, sVal)

		sModel, sErr := localProvider.GetEntityModel(sVal, true)
		if sErr != nil {
			err = sErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		sModelList = append(sModelList, sModel)
	}

	for idx := 0; idx < localLoop; idx++ {
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
	for idx := 0; idx < localLoop; idx++ {
		sVal := sValList[idx]
		sVal.Name = "hi"
		sModel, sErr := localProvider.GetEntityModel(sVal, true)
		if sErr != nil {
			err = sErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		sModelList[idx] = sModel
	}
	for idx := 0; idx < localLoop; idx++ {
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
	qModelList := []models.Model{}
	for idx := 0; idx < localLoop; idx++ {
		var fVal float32
		var ts time.Time
		var flag bool
		strArray := []string{}
		ptrStrArray := []string{}

		qVal := &Reference{
			ID:          sValList[idx].ID,
			FValue:      fVal,
			TimeStamp:   ts,
			Flag:        flag,
			IArray:      []int{},
			FArray:      []float32{},
			StrArray:    []string{},
			BArray:      []bool{},
			PtrArray:    &strArray,
			StrPtrArray: []string{},
			PtrStrArray: &ptrStrArray,
		}
		qValList = append(qValList, qVal)

		qModel, qErr := localProvider.GetEntityModel(qVal, true)
		if qErr != nil {
			err = qErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		qModelList = append(qModelList, qModel)
	}

	for idx := 0; idx < localLoop; idx++ {
		qModel, qErr := o1.Query(qModelList[idx])
		if qErr != nil {
			err = qErr
			t.Errorf("Query failed. err:%s", err.Error())
			return
		}

		qModelList[idx] = qModel
		qValList[idx] = qModel.Interface(true).(*Reference)
	}

	for idx := 0; idx < localLoop; idx++ {
		sVal := sValList[idx]
		qVal := qValList[idx]
		if !sVal.IsSame(qVal) {
			err = cd.NewError(cd.Unexpected, "compare value failed")
			t.Errorf("IsSame failed. err:%s", err.Error())
			return
		}
	}

	var fVal float32
	var ts2 time.Time
	var flag2 bool
	strArray2 := []string{}
	ptrStrArray := []string{}

	referenceModel, _ := localProvider.GetEntityModel(&Reference{}, true)
	filter, err := localProvider.GetModelFilter(referenceModel)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	err = filter.Equal("name", "hi")
	if err != nil {
		t.Errorf("filter.Equal failed, err:%s", err.Error())
		return
	}

	err = filter.ValueMask(&Reference{FValue: fVal, TimeStamp: ts2, Flag: flag2, PtrArray: &strArray2, PtrStrArray: &ptrStrArray})
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
		t.Errorf("batch query reference failed")
		return
	}

	// delete
	for idx := 0; idx < localLoop; idx++ {
		_, qErr := o1.Delete(bqModelList[idx])
		if qErr != nil {
			t.Errorf("Delete failed. err:%s", qErr.Error())
			return
		}
	}
}
