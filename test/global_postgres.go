//go:build !mysql
// +build !mysql

package test

import (
	"os"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

const (
	defaultPostgresTestServer   = "localhost:5432"
	defaultPostgresTestDatabase = "testdb"
	defaultPostgresTestUser     = "postgres"
	defaultPostgresTestPassword = "rootkit"
)

var config = orm.NewConfig(
	postgresTestEnv("MAGICORM_POSTGRES_SERVER", defaultPostgresTestServer),
	postgresTestEnv("MAGICORM_POSTGRES_DATABASE", defaultPostgresTestDatabase),
	postgresTestEnv("MAGICORM_POSTGRES_USER", defaultPostgresTestUser),
	postgresTestEnv("MAGICORM_POSTGRES_PASSWORD", defaultPostgresTestPassword),
)

func postgresTestEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func registerLocalModel(provider provider.Provider, objList []any) (ret []models.Model, err *cd.Error) {
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

func registerRemoteModel(provider provider.Provider, objList []any) (ret []models.Model, err *cd.Error) {
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

func createModel(orm orm.Orm, modelList []models.Model) (err *cd.Error) {
	for _, val := range modelList {
		err = orm.Create(val)
		if err != nil {
			return
		}
	}

	return
}

func dropModel(orm orm.Orm, modelList []models.Model) (err *cd.Error) {
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
