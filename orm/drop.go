package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) dropSingle(vModel model.Model) (err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	dropSQL, dropErr := builder.BuildDropTable()
	if dropErr != nil {
		err = dropErr
		return
	}

	_, _, err = s.executor.Execute(dropSQL)
	return
}

func (s *impl) dropRelation(vModel model.Model, vField model.Field, rModel model.Model) (err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationSQL, relationErr := builder.BuildDropRelationTable(vField, rModel)
	if relationErr != nil {
		err = relationErr
		return
	}

	_, _, err = s.executor.Execute(relationSQL)
	return
}

func (s *impl) dropSchema(vModel model.Model) (err error) {
	err = s.dropSingle(vModel)
	if err != nil {
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
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			err = s.dropSingle(relationModel)
			if err != nil {
				return
			}
		}

		err = s.dropRelation(vModel, field, relationModel)
		if err != nil {
			break
		}
	}

	return
}

func (s *impl) Drop(vModel model.Model) (err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	err = s.dropSchema(vModel)
	return
}
