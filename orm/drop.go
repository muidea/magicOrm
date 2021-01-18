package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) dropSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
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

		_, err = s.executor.Execute(sql)
	}

	return
}

func (s *impl) dropRelation(modelInfo model.Model, fieldName string, relationInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	tableName := builder.GetRelationTableName(fieldName, relationInfo)

	existFlag, existErr := s.executor.CheckTableExist(tableName)
	if existErr != nil {
		err = existErr
		return
	}
	if existFlag {
		sql, err := builder.BuildDropRelationSchema(fieldName, relationInfo)
		if err != nil {
			return err
		}

		_, err = s.executor.Execute(sql)
	}

	return
}

// Drop drop
func (s *impl) Drop(entityModel model.Model) (err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}

	for {
		err = s.dropSingle(entityModel)
		if err != nil {
			break
		}

		for _, field := range entityModel.GetFields() {
			fType := field.GetType()
			if fType.IsBasic() {
				continue
			}

			relationInfo, relationErr := s.modelProvider.GetTypeModel(fType)
			if relationErr != nil {
				err = relationErr
				break
			}

			elemType := fType.Elem()
			if !elemType.IsPtrType() {
				err = s.dropSingle(relationInfo)
				if err != nil {
					break
				}
			}

			err = s.dropRelation(entityModel, field.GetName(), relationInfo)
			if err != nil {
				break
			}
		}
		break
	}

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
