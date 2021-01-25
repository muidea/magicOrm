package test

import (
	"fmt"
	orm "github.com/muidea/magicOrm"
	"github.com/muidea/magicOrm/model"
	"testing"
	"time"
)

const simpleOwner = "simple"

func TestLocal(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	o1, err := orm.NewOrm(simpleOwner)
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	provider := orm.GetProvider(simpleOwner)

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

	ts, _ := time.Parse("2006-01-02 15:04:05", "2018-01-02 15:04:05")
	sVal := Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}
	sValList := []*Simple{}
	sModelList := []model.Model{}

	// insert
	for idx := 0; idx < 100; idx++ {
		sVal.I32 = int32(idx)
		sValList = append(sValList, &sVal)

		sModel, sErr := provider.GetEntityModel(&sVal)
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
		sValList[idx] = vModel.Interface(true).(*Simple)
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
		sValList[idx] = vModel.Interface(true).(*Simple)
	}

	// query
	qValList := []*Simple{}
	qModelList := []model.Model{}
	for idx := 0; idx < 100; idx++ {
		qVal := &Simple{ID: sValList[idx].ID}
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
		qValList[idx] = qModel.Interface(true).(*Simple)
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

	bqValList := []*Simple{}
	bqModel, bqErr := provider.GetEntityModel(&bqValList)
	if bqErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", bqErr.Error())
		return
	}

	filter := orm.GetFilter(localOwner)
	filter.Equal("Name", "hi")
	filter.ValueMask(&Simple{})
	bqModelList, bqModelErr := o1.BatchQuery(bqModel, filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != 100 {
		t.Errorf("batch query simple failed")
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
