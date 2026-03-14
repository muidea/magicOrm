package test

import (
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestComposeLocal(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider(composeLocalOwner, nil)

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

	sPtr, rPtr, cPtr, pErr := prepareLocalData(localProvider, o1)
	if pErr != nil {
		t.Errorf("prepareLocalData failed. err:%s", pErr.Error())
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

	composeModel, composeErr := localProvider.GetEntityModel(composePtr, true)
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

	composePtr = composeModel.Interface(true).(*Compose)
	composePtr.Name = "hi"
	composeModel, composeErr = localProvider.GetEntityModel(composePtr, true)
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

	composePtr = composeModel.Interface(true).(*Compose)

	queryVal := &Compose{
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

	queryModel, queryErr := localProvider.GetEntityModel(queryVal, true)
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
	queryVal = queryModel.Interface(true).(*Compose)

	if !composePtr.IsSame(queryVal) {
		err = cd.NewError(cd.Unexpected, "compare value failed")
		t.Errorf("IsSame failed. err:%s", err.Error())
		return
	}

	cModel, _ := localProvider.GetEntityModel(&Compose{}, true)
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
		SimplePtr:         &Simple{},
		SimpleArray:       []Simple{},
		SimplePtrArray:    []*Simple{},
		SimpleArrayPtr:    &[]Simple{},
		ReferencePtr:      &Reference{},
		ReferenceArray:    []Reference{},
		ReferencePtrArray: []*Reference{},
		ComposePtr:        &Compose{},
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

	_, qErr := o1.Delete(bqModelList[0])
	if qErr != nil {
		t.Errorf("Delete failed, err:%s", qErr.Error())
		return
	}
}
