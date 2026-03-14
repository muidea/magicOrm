package test

import (
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestComposeRemote(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()
	remoteProvider := provider.NewRemoteProvider(composeRemoteOwner, nil)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	entityList := []any{&Simple{}, &Reference{}, &Compose{}}
	modelList, modelErr := registerRemoteModel(remoteProvider, entityList)
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
	refPtrArray := []*Reference{rPtr}
	composePtr := &Compose{
		Name:              strValue,
		Simple:            *sPtr,
		SimplePtr:         sPtr,
		SimpleArray:       []Simple{*sPtr, *sPtr},
		SimplePtrArray:    []*Simple{sPtr, sPtr},
		Reference:         *rPtr,
		ReferencePtr:      rPtr,
		ReferenceArray:    []Reference{*rPtr, *rPtr, *rPtr},
		ReferencePtrArray: refPtrArray,
		ComposePtr:        cPtr,
	}
	composeObjectValue, composeObjectErr := helper.GetObjectValue(composePtr)
	if composeObjectErr != nil {
		err = composeObjectErr
		t.Errorf("GetObjectValue failed. err:%s", err.Error())
		return
	}

	composeModel, composeErr := remoteProvider.GetEntityModel(composeObjectValue, true)
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

	composeObjectValue = composeModel.Interface(true).(*remote.ObjectValue)
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

	composeModel, composeErr = remoteProvider.GetEntityModel(composeObjectValue, true)
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

	composeObjectValue = vModel.Interface(true).(*remote.ObjectValue)
	err = helper.UpdateEntity(composeObjectValue, composePtr)
	if err != nil {
		t.Errorf("UpdateEntity failed. err:%s", err.Error())
		return
	}

	queryComposeVal := &Compose{
		ID:                composePtr.ID,
		SimplePtr:         &Simple{},
		SimpleArray:       []Simple{},
		SimplePtrArray:    []*Simple{},
		SimpleArrayPtr:    &[]Simple{},
		ReferencePtr:      &Reference{},
		ReferenceArray:    []Reference{},
		ReferencePtrArray: []*Reference{},
		ComposePtr:        &Compose{},
	}

	queryObjectValue, queryObjectErr := helper.GetObjectValue(queryComposeVal)
	if queryObjectErr != nil {
		err = queryObjectErr
		t.Errorf("GetObjectValue failed. err:%s", err.Error())
		return
	}

	queryModel, queryErr := remoteProvider.GetEntityModel(queryObjectValue, true)
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

	queryObjectVal := queryModel.Interface(true).(*remote.ObjectValue)
	err = helper.UpdateEntity(queryObjectVal, queryComposeVal)
	if err != nil {
		t.Errorf("UpdateEntity failed. err:%s", err.Error())
		return
	}

	if !composePtr.IsSame(queryComposeVal) {
		err = cd.NewError(cd.Unexpected, "compare value failed")
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
		SimplePtr:         &Simple{},
		SimpleArray:       []Simple{},
		SimplePtrArray:    []*Simple{},
		SimpleArrayPtr:    &[]Simple{},
		ReferencePtr:      &Reference{},
		ReferenceArray:    []Reference{},
		ReferencePtrArray: []*Reference{},
		ComposePtr:        &Compose{},
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

	_, qErr := o1.Delete(bqModelList[0])
	if qErr != nil {
		err = qErr
		t.Errorf("o1.Delete failed, err:%s", qErr.Error())
		return
	}
}
