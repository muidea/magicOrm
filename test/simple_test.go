//go:build mixed || all
// +build mixed all

package test

import (
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

const simpleLocalOwner = "simpleLocal"
const simpleRemoteOwner = "simpleRemote"

func TestSimpleLocal(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	loopSize := 10
	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider(simpleLocalOwner)

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

	//ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	sValList := []*Simple{}
	sModelList := []model.Model{}

	// insert
	for idx := 0; idx < loopSize; idx++ {
		sVal := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, Flag: true}
		sVal.I32 = int32(idx)
		sValList = append(sValList, sVal)

		sModel, sErr := localProvider.GetEntityModel(sVal)
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

	// update
	for idx := 0; idx < loopSize; idx++ {
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

	// query
	qValList := []*Simple{}
	qModelList := []model.Model{}
	for idx := 0; idx < loopSize; idx++ {
		qVal := &Simple{ID: sValList[idx].ID}
		qValList = append(qValList, qVal)

		qModel, qErr := localProvider.GetEntityModel(qVal)
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
			err = cd.NewResult(cd.UnExpected, "compare value failed")
			t.Errorf("IsSame failed. err:%s", err.Error())
			return
		}
	}

	simpleModel, _ := localProvider.GetEntityModel(&Simple{})
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

	// delete
	for idx := 0; idx < loopSize; idx++ {
		_, qErr := o1.Delete(bqModelList[idx])
		if qErr != nil {
			t.Errorf("Delete failed. err:%s", qErr.Error())
			return
		}
	}
}

func TestSimpleRemote(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider(simpleRemoteOwner)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	simpleDef, _ := helper.GetObject(&Simple{})
	referenceDef, _ := helper.GetObject(&Reference{})
	composeDef, _ := helper.GetObject(&Compose{})

	entityList := []any{simpleDef, referenceDef, composeDef}
	modelList, modelErr := registerLocalModel(remoteProvider, entityList)
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

	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	sValList := []*Simple{}
	sObjectValList := []*remote.ObjectValue{}
	sModelList := []model.Model{}

	// insert
	for idx := 0; idx < 100; idx++ {
		sVal := Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}
		sVal.I32 = int32(idx)
		sValList = append(sValList, &sVal)

		sObjectVal, sObjectErr := helper.GetObjectValue(&sVal)
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
		err = helper.UpdateEntity(sObjectVal, sVal)
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
		sObjectVal, sObjectErr := helper.GetObjectValue(sVal)
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
		err = helper.UpdateEntity(sObjectVal, sVal)
		if err != nil {
			t.Errorf("UpdateEntity failed. err:%s", err.Error())
			return
		}
		sValList[idx] = sVal
		sModelList[idx] = vModel
		sObjectValList[idx] = sObjectVal
	}

	// query
	qValList := []*Simple{}
	qObjectValList := []*remote.ObjectValue{}
	qModelList := []model.Model{}
	for idx := 0; idx < 100; idx++ {
		qVal := &Simple{ID: sValList[idx].ID}
		qValList = append(qValList, qVal)

		qObjectVal, qObjectErr := helper.GetObjectValue(qVal)
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
		err = helper.UpdateEntity(qObjectVal, qVal)
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
			err = cd.NewResult(cd.UnExpected, "compare value failed")
			t.Errorf("IsSame failed. err:%s", err.Error())
			return
		}
	}

	objectPtr, objectErr := helper.GetObject(&Simple{})
	if objectErr != nil {
		t.Errorf("GetObject failed, error:%s", objectErr.Error())
	}
	filter, err := remoteProvider.GetModelFilter(objectPtr)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	filter.Equal("name", "hi")
	bqModelList, bqModelErr := o1.BatchQuery(filter)
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
			t.Errorf("Delete failed. err:%s", qErr.Error())
			return
		}
	}
}
