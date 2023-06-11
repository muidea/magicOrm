package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) dropSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider, s.specialPrefix)
	tableName := builder.GetTableName()

	existFlag, existErr := s.executor.CheckTableExist(tableName)
	if existErr != nil {
		err = existErr
		return
	}

	if existFlag {
		sql, err := builder.BuildDropSchema()
		if err != nil {
			return err
		}

		_, _, err = s.executor.Execute(sql)
	}

	return
}

func (s *impl) dropRelation(modelInfo model.Model, field model.Field, relationInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider, s.specialPrefix)
	relationSchema := builder.GetRelationTableName(field, relationInfo)

	existFlag, existErr := s.executor.CheckTableExist(relationSchema)
	if existErr != nil {
		err = existErr
		return
	}
	if existFlag {
		sql, err := builder.BuildDropRelationSchema(relationSchema)
		if err != nil {
			return err
		}

		_, _, err = s.executor.Execute(sql)
	}

	return
}

func (s *impl) batchDropSchema(modelInfo model.Model) (err error) {
	err = s.dropSingle(modelInfo)
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
			err = s.dropSingle(relationInfo)
			if err != nil {
				return
			}
		}

		err = s.dropRelation(modelInfo, field, relationInfo)
		if err != nil {
			break
		}
	}

	return
}

// Drop drop
func (s *impl) Drop(entityModel model.Model) (err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}

	err = s.batchDropSchema(entityModel)
	if err == nil {
		cErr := s.executor.CommitTransaction()
		if cErr != nil {
			err = cErr
		}
	} else {
		rErr := s.executor.RollbackTransaction()
		if rErr != nil {
			err = rErr
		}
	}

	return
}
