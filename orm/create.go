package orm

import (
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) createSingle(vModel model.Model) (err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	createSQL, createErr := builder.BuildCreateTable()
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
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationSQL, relationErr := builder.BuildCreateRelationTable(vField, rModel)
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

		fType := field.GetType()
		relationModel, relationErr := s.modelProvider.GetTypeModel(fType)
		if relationErr != nil {
			err = relationErr
			log.Errorf("createSchema failed, s.modelProvider.GetTypeModel error:%s", err.Error())
			return
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

func (s *impl) Create(vModel model.Model) (err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer func() {
		err = s.finalTransaction(err)
	}()

	err = s.createSchema(vModel)
	if err != nil {
		log.Errorf("Create failed, s.createSchema error:%s", err.Error())
	}
	return
}
