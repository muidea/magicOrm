package test

import (
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

const simpleLocalOwner = "simpleLocal"

func TestSimpleLocal(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	loopSize := 10
	localProvider := provider.NewLocalProvider(simpleLocalOwner, nil)

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

	sValList := []*Simple{}
	sModelList := []models.Model{}

	for idx := 0; idx < loopSize; idx++ {
		sVal := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, Flag: true}
		sVal.I32 = int32(idx)
		sValList = append(sValList, sVal)

		sModel, sErr := localProvider.GetEntityModel(sVal, true)
		if sErr != nil {
			err = sErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		sModelList = append(sModelList, sModel)
	}

	for idx := 0; idx < loopSize; idx++ {
		vModel, vErr := o1.Insert(sModelList[idx])
		if vErr != nil {
			err = vErr
			t.Errorf("Insert failed. err:%s", err.Error())
			return
		}

		sModelList[idx] = vModel
		sValList[idx] = vModel.Interface(true).(*Simple)
	}

	for idx := 0; idx < loopSize; idx++ {
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
	for idx := 0; idx < loopSize; idx++ {
		vModel, vErr := o1.Update(sModelList[idx])
		if vErr != nil {
			err = vErr
			t.Errorf("Update failed. err:%s", err.Error())
			return
		}

		sModelList[idx] = vModel
		sValList[idx] = vModel.Interface(true).(*Simple)
	}

	qValList := []*Simple{}
	qModelList := []models.Model{}
	for idx := 0; idx < loopSize; idx++ {
		qVal := &Simple{ID: sValList[idx].ID}
		qValList = append(qValList, qVal)

		qModel, qErr := localProvider.GetEntityModel(qVal, true)
		if qErr != nil {
			err = qErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}

		qModelList = append(qModelList, qModel)
	}

	for idx := 0; idx < loopSize; idx++ {
		qModel, qErr := o1.Query(qModelList[idx])
		if qErr != nil {
			err = qErr
			t.Errorf("Query failed. err:%s", err.Error())
			return
		}

		qModelList[idx] = qModel
		qValList[idx] = qModel.Interface(true).(*Simple)
	}

	for idx := 0; idx < loopSize; idx++ {
		sVal := sValList[idx]
		qVal := qValList[idx]
		if !sVal.IsSame(qVal) {
			err = cd.NewError(cd.Unexpected, "compare value failed")
			t.Errorf("IsSame failed. err:%s", err.Error())
			return
		}
	}

	simpleModel, _ := localProvider.GetEntityModel(&Simple{}, true)
	filter, err := localProvider.GetModelFilter(simpleModel)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	filter.Equal("name", "hi")
	filter.ValueMask(&Simple{})
	bqModelList, bqModelErr := o1.BatchQuery(filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != loopSize {
		t.Errorf("batch query simple failed")
		return
	}

	for idx := 0; idx < loopSize; idx++ {
		_, qErr := o1.Delete(bqModelList[idx])
		if qErr != nil {
			t.Errorf("Delete failed. err:%s", qErr.Error())
			return
		}
	}
}
