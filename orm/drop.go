package orm

import (
	"fmt"

	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) dropSingle(vModel model.Model) (err error) {
	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	dropSQL, dropErr := builderVal.BuildDropTable()
	if dropErr != nil {
		err = dropErr
		log.Errorf("dropSingle failed, builder.BuildDropTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(dropSQL)
	if err != nil {
		log.Errorf("dropSingle failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) dropRelation(vModel model.Model, vField model.Field, rModel model.Model) (err error) {
	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationSQL, relationErr := builderVal.BuildDropRelationTable(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("dropRelation failed, builder.BuildDropRelationTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(relationSQL)
	if err != nil {
		log.Errorf("dropRelation failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) dropSchema(vModel model.Model) (err error) {
	err = s.dropSingle(vModel)
	if err != nil {
		log.Errorf("dropSchema failed, s.dropSingle error:%s", err.Error())
		return
	}

	for _, field := range vModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		var relationModel model.Model
		fType := field.GetType()
		if vModel.GetPkgKey() == fType.GetPkgKey() {
			relationModel = vModel
		} else {
			relationModel, err = s.modelProvider.GetTypeModel(fType)
			if err != nil {
				log.Errorf("dropSchema failed, s.modelProvider.GetTypeModel error:%s", err.Error())
				return
			}
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			err = s.dropSingle(relationModel)
			if err != nil {
				log.Errorf("dropSchema failed, s.dropSingle error:%s", err.Error())
				return
			}
		}

		err = s.dropRelation(vModel, field, relationModel)
		if err != nil {
			log.Errorf("dropSchema failed, s.dropRelation error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) Drop(vModel model.Model) (err error) {
	if vModel == nil {
		err = fmt.Errorf("illegal model value")
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	err = s.dropSchema(vModel)
	if err != nil {
		log.Errorf("Drop failed, s.dropSchema error:%s", err.Error())
	}
	return
}
