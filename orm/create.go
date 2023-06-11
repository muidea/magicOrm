package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) createSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider, s.specialPrefix)
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

func (s *impl) createRelation(modelInfo model.Model, field model.Field, relationInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider, s.specialPrefix)
	relationSchema := builder.GetRelationTableName(field, relationInfo)

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

func (s *impl) batchCreateSchema(modelInfo model.Model) (err error) {
	err = s.createSingle(modelInfo)
	if err != nil {
		return
	}

	for _, field := range modelInfo.GetFields() {
		fType := field.GetType()
		if fType.IsBasic() {
			continue
		}

		relationInfo, relationErr := s.modelProvider.GetTypeModel(fType)
		if relationErr != nil {
			err = relationErr
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			err = s.createSingle(relationInfo)
			if err != nil {
				return
			}
		}

		err = s.createRelation(modelInfo, field, relationInfo)
		if err != nil {
			return
		}
	}

	return
}

func (s *impl) Create(entityModel model.Model) (err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	err = s.batchCreateSchema(entityModel)
	return
}
