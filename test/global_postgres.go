//go:build !mysql
// +build !mysql

package test

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

var config = orm.NewConfig("localhost:5432", "magicplatform_db", "postgres", "rootkit")

func registerLocalModel(provider provider.Provider, objList []any) (ret []model.Model, err *cd.Error) {
	for _, val := range objList {
		modelVal, modelErr := provider.RegisterModel(val)
		if modelErr != nil {
			err = modelErr
			return
		}

		ret = append(ret, modelVal)
	}

	return
}

func registerRemoteModel(provider provider.Provider, objList []any) (ret []model.Model, err *cd.Error) {
	for _, val := range objList {
		remoteObjectPtr, remoteObjectErr := helper.GetObject(val)
		if remoteObjectErr != nil {
			err = remoteObjectErr
			return
		}
		modelVal, modelErr := provider.RegisterModel(remoteObjectPtr)
		if modelErr != nil {
			err = modelErr
			return
		}

		ret = append(ret, modelVal)
	}

	return
}

func createModel(orm orm.Orm, modelList []model.Model) (err *cd.Error) {
	for _, val := range modelList {
		err = orm.Create(val)
		if err != nil {
			return
		}
	}

	return
}

func dropModel(orm orm.Orm, modelList []model.Model) (err *cd.Error) {
	for _, val := range modelList {
		err = orm.Drop(val)
		if err != nil {
			return
		}
	}

	return
}

func getObjectValue(val any) (ret *remote.ObjectValue, err *cd.Error) {
	objVal, objErr := helper.GetObjectValue(val)
	if objErr != nil {
		err = objErr
		return
	}

	data, dataErr := remote.EncodeObjectValue(objVal)
	if dataErr != nil {
		err = dataErr
		return
	}
	ret, err = remote.DecodeObjectValue(data)
	if err != nil {
		return
	}

	return
}
