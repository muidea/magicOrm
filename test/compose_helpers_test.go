package test

import (
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

const composeLocalOwner = "composeLocal"
const composeRemoteOwner = "composeRemote"

func prepareLocalData(localProvider provider.Provider, orm orm.Orm) (sPtr *Simple, rPtr *Reference, cPtr *Compose, err *cd.Error) {
	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	sVal := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}

	sModel, sErr := localProvider.GetEntityModel(sVal, true)
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
	strPtrArray := []string{strValue, strValue}
	rVal := &Reference{
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

	rModel, rErr := localProvider.GetEntityModel(rVal, true)
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
		Name:              strValue,
		Simple:            *sPtr,
		SimplePtr:         sPtr,
		SimpleArray:       []Simple{*sPtr, *sPtr},
		SimplePtrArray:    []*Simple{sPtr, sPtr},
		Reference:         *rPtr,
		ReferencePtr:      rPtr,
		ReferenceArray:    []Reference{*rPtr, *rPtr, *rPtr},
		ReferencePtrArray: refPtrArray,
	}
	cModel, cErr := localProvider.GetEntityModel(cVal, true)
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

func prepareRemoteData(remoteProvider provider.Provider, orm orm.Orm) (sPtr *Simple, rPtr *Reference, cPtr *Compose, err *cd.Error) {
	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	sVal := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}

	sObjectVal, _ := helper.GetObjectValue(sVal)
	sModel, sErr := remoteProvider.GetEntityModel(sObjectVal, true)
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
	strPtrArray := []string{strValue, strValue}
	rVal := &Reference{
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

	rObjectVal, _ := helper.GetObjectValue(rVal)
	rModel, rErr := remoteProvider.GetEntityModel(rObjectVal, true)
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
	ptrStrArray := []string{}

	rPtr = &Reference{FValue: fVal, TimeStamp: ts2, Flag: flag2, PtrArray: &strArray2, PtrStrArray: &ptrStrArray}
	rErr = helper.UpdateEntity(rObjectVal, rPtr)
	if rErr != nil {
		err = rErr
		return
	}

	refPtrArray := []*Reference{rPtr}
	cVal := &Compose{
		Name:              strValue,
		Simple:            *sPtr,
		SimplePtr:         sPtr,
		SimpleArray:       []Simple{*sPtr, *sPtr},
		SimplePtrArray:    []*Simple{sPtr, sPtr},
		Reference:         *rPtr,
		ReferencePtr:      rPtr,
		ReferenceArray:    []Reference{*rPtr, *rPtr, *rPtr},
		ReferencePtrArray: refPtrArray,
	}
	cObjectVal, _ := helper.GetObjectValue(cVal)
	cModel, cErr := remoteProvider.GetEntityModel(cObjectVal, true)
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
		SimplePtr:         &Simple{},
		SimpleArray:       []Simple{},
		SimplePtrArray:    []*Simple{},
		SimpleArrayPtr:    &[]Simple{},
		ReferencePtr:      &Reference{},
		ReferenceArray:    []Reference{},
		ReferencePtrArray: []*Reference{},
		ComposePtr:        &Compose{},
	}
	cErr = helper.UpdateEntity(cObjectVal, cPtr)
	if cErr != nil {
		err = cErr
		return
	}

	return
}
