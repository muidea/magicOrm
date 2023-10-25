package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) createSingle(vModel model.Model) (err error) {
	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	createSQL, createErr := builderVal.BuildCreateTable()
	if createErr != nil {
		err = createErr
		log.Errorf("createSingle failed, builder.BuildCreateTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(createSQL)
	if err != nil {
		log.Errorf("createSingle failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) createRelation(vModel model.Model, vField model.Field, rModel model.Model) (err error) {
	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationSQL, relationErr := builderVal.BuildCreateRelationTable(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("createRelation failed, builder.BuildCreateRelationTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(relationSQL)
	if err != nil {
		log.Errorf("createRelation failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) createSchema(vModel model.Model) (err error) {
	err = s.createSingle(vModel)
	if err != nil {
		log.Errorf("createSchema failed, s.createSingle error:%s", err.Error())
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
				log.Errorf("createSchema failed, s.modelProvider.GetTypeModel error:%s", err.Error())
				return
			}
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			err = s.createSingle(relationModel)
			if err != nil {
				log.Errorf("createSchema failed, s.createSingle error:%s", err.Error())
				return
			}
		}

		err = s.createRelation(vModel, field, relationModel)
		if err != nil {
			log.Errorf("createSchema failed, s.createRelation error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) Create(vModel model.Model) (re *cd.Result) {
	if vModel == nil {
		re = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}
	err := s.executor.BeginTransaction()
	if err != nil {
		re = cd.NewError(cd.UnExpected, err.Error())
		return
	}

	defer s.finalTransaction(err)

	err = s.createSchema(vModel)
	if err != nil {
		re = cd.NewError(cd.UnExpected, err.Error())
		log.Errorf("Create failed, s.createSchema error:%s", err.Error())
	}
	return
}
