package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) createSingle(vModel model.Model) (err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	tableName := builder.GetTableName()

	existFlag, existErr := s.executor.CheckTableExist(tableName)
	if existErr != nil {
		err = existErr
		return
	}

	if !existFlag {
		// no exist
		sql, err := builder.BuildCreateSchema()
		if err != nil {
			return err
		}

		_, _, err = s.executor.Execute(sql)
	}

	return
}

func (s *impl) createRelation(vModel model.Model, vField model.Field, rModel model.Model) (err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationSchema := builder.GetRelationTableName(vField, rModel)

	existFlag, existErr := s.executor.CheckTableExist(relationSchema)
	if existErr != nil {
		err = existErr
		return
	}
	if !existFlag {
		// no exist
		sql, err := builder.BuildCreateRelationSchema(relationSchema)
		if err != nil {
			return err
		}

		_, _, err = s.executor.Execute(sql)
	}

	return
}

func (s *impl) batchCreateSchema(vModel model.Model) (err error) {
	err = s.createSingle(vModel)
	if err != nil {
		return
	}

	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		if fType.IsBasic() {
			continue
		}

		relationModel, relationErr := s.modelProvider.GetTypeModel(fType)
		if relationErr != nil {
			err = relationErr
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			err = s.createSingle(relationModel)
			if err != nil {
				return
			}
		}

		err = s.createRelation(vModel, field, relationModel)
		if err != nil {
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
	defer s.finalTransaction(err)

	err = s.batchCreateSchema(vModel)
	return
}
