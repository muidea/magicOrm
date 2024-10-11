package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) createSingle(hBuilder builder.Builder) (err *cd.Result) {
	createResult, createErr := hBuilder.BuildCreateTable()
	if createErr != nil {
		err = createErr
		log.Errorf("createSingle failed, hBuilder.BuildCreateTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(createResult.SQL(), createResult.Args()...)
	if err != nil {
		log.Errorf("createSingle failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) createRelation(hBuilder builder.Builder, vField model.Field, rModel model.Model) (err *cd.Result) {
	relationResult, relationErr := hBuilder.BuildCreateRelationTable(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("createRelation failed, hBuilder.BuildCreateRelationTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(relationResult.SQL(), relationResult.Args()...)
	if err != nil {
		log.Errorf("createRelation failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) createSchema(vModel model.Model) (err *cd.Result) {
	hContext := codec.New(s.modelProvider, s.specialPrefix)
	hBuilder := builder.NewBuilder(vModel, hContext)
	err = s.createSingle(hBuilder)
	if err != nil {
		log.Errorf("createSchema failed, s.createSingle error:%s", err.Error())
		return
	}

	for _, field := range vModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		fType := field.GetType()
		relationModel, relationErr := s.modelProvider.GetTypeModel(fType)
		if relationErr != nil {
			err = relationErr
			log.Errorf("createSchema failed, s.modelProvider.GetTypeModel error:%s", err.Error())
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			rContext := codec.New(s.modelProvider, s.specialPrefix)
			rBuilder := builder.NewBuilder(relationModel, rContext)
			err = s.createSingle(rBuilder)
			if err != nil {
				log.Errorf("createSchema failed, s.createSingle error:%s", err.Error())
				return
			}
		}

		err = s.createRelation(hBuilder, field, relationModel)
		if err != nil {
			log.Errorf("createSchema failed, s.createRelation error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) Create(vModel model.Model) (err *cd.Result) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err = s.createSchema(vModel)
	if err != nil {
		log.Errorf("Create failed, s.createSchema error:%s", err.Error())
	}
	return
}
