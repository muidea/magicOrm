package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) dropSingle(vModel model.Model) (err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	tableName := builder.GetTableName()

	existFlag, existErr := s.executor.CheckTableExist(tableName)
	if existErr != nil {
		err = existErr
		return
	}

	if existFlag {
		sql, err := builder.BuildDropTable()
		if err != nil {
			return err
		}

		_, _, err = s.executor.Execute(sql)
	}

	return
}

func (s *impl) dropRelation(vModel model.Model, field model.Field, rModel model.Model) (err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationTableName := builder.GetRelationTableName(field, rModel)

	existFlag, existErr := s.executor.CheckTableExist(relationTableName)
	if existErr != nil {
		err = existErr
		return
	}
	if existFlag {
		sql, err := builder.BuildDropRelationTable(relationTableName)
		if err != nil {
			return err
		}

		_, _, err = s.executor.Execute(sql)
	}

	return
}

func (s *impl) batchDropSchema(vModel model.Model) (err error) {
	err = s.dropSingle(vModel)
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

	err = s.batchDropSchema(vModel)
	return
}
