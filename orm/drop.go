package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/database/context"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) dropSingle(hBuilder builder.Builder) (err *cd.Result) {
	dropResult, dropErr := hBuilder.BuildDropTable()
	if dropErr != nil {
		err = dropErr
		log.Errorf("dropSingle failed, hBuilder.BuildDropTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(dropResult.SQL(), dropResult.Args()...)
	if err != nil {
		log.Errorf("dropSingle failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) dropRelation(hBuilder builder.Builder, vField model.Field, rModel model.Model) (err *cd.Result) {
	relationResult, relationErr := hBuilder.BuildDropRelationTable(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("dropRelation failed, hBuilder.BuildDropRelationTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(relationResult.SQL(), relationResult.Args()...)
	if err != nil {
		log.Errorf("dropRelation failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) dropSchema(vModel model.Model) (err *cd.Result) {
	hContext := context.New(vModel, s.modelProvider, s.specialPrefix)
	hBuilder := builder.NewBuilder(vModel, hContext)
	err = s.dropSingle(hBuilder)
	if err != nil {
		log.Errorf("dropSchema failed, s.dropSingle error:%s", err.Error())
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
			log.Errorf("dropSchema failed, s.modelProvider.GetTypeModel error:%s", err.Error())
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			rContext := context.New(relationModel, s.modelProvider, s.specialPrefix)
			rBuilder := builder.NewBuilder(relationModel, rContext)
			err = s.dropSingle(rBuilder)
			if err != nil {
				log.Errorf("dropSchema failed, s.dropSingle error:%s", err.Error())
				return
			}
		}

		err = s.dropRelation(hBuilder, field, relationModel)
		if err != nil {
			log.Errorf("dropSchema failed, s.dropRelation error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) Drop(vModel model.Model) (err *cd.Result) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err = s.dropSchema(vModel)
	if err != nil {
		log.Errorf("Drop failed, s.dropSchema error:%s", err.Error())
	}
	return
}
