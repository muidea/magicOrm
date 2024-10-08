package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) dropSingle(hBuilder builder.Builder) (err *cd.Result) {
	dropSQL, dropErr := hBuilder.BuildDropTable()
	if dropErr != nil {
		err = dropErr
		log.Errorf("dropSingle failed, hBuilder.BuildDropTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(dropSQL)
	if err != nil {
		log.Errorf("dropSingle failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) dropRelation(hBuilder builder.Builder, vField model.Field, rModel model.Model) (err *cd.Result) {
	relationSQL, relationErr := hBuilder.BuildDropRelationTable(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("dropRelation failed, hBuilder.BuildDropRelationTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(relationSQL)
	if err != nil {
		log.Errorf("dropRelation failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) dropSchema(vModel model.Model) (err *cd.Result) {
	hBuilder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
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
			rBuilder := builder.NewBuilder(relationModel, s.modelProvider, s.specialPrefix)
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
