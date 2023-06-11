package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) insertSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builder.BuildInsert()
	if sqlErr != nil {
		err = sqlErr
		return err
	}

	_, id, idErr := s.executor.Execute(sqlStr)
	if idErr != nil {
		err = idErr
		return
	}

	pk := modelInfo.GetPrimaryField()
	tVal, tErr := s.modelProvider.DecodeValue(id, pk.GetType())
	if tErr != nil {
		err = tErr
		return
	}

	err = pk.SetValue(tVal)
	if err != nil {
		return err
	}

	return
}

func (s *impl) insertRelation(modelInfo model.Model, fieldInfo model.Field) (err error) {
	fValue := fieldInfo.GetValue()
	fType := fieldInfo.GetType()
	if fType.IsBasic() || fValue.IsNil() /* || !s.modelProvider.IsAssigned(fValue, fType)*/ {
		return
	}

	fSliceValue, fSliceErr := s.modelProvider.ElemDependValue(fValue)
	if fSliceErr != nil {
		err = fSliceErr
		return
	}

	for _, fVal := range fSliceValue {
		relationInfo, relationErr := s.modelProvider.GetValueModel(fVal, fType)
		if relationErr != nil {
			err = relationErr
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			err = s.insertSingle(relationInfo)
			if err != nil {
				return
			}

			for _, subField := range relationInfo.GetFields() {
				err = s.insertRelation(relationInfo, subField)
				if err != nil {
					return
				}
			}
		}

		builder := builder.NewBuilder(modelInfo, s.modelProvider, s.specialPrefix)
		relationSQL, relationErr := builder.BuildInsertRelation(fieldInfo, relationInfo)
		if relationErr != nil {
			err = relationErr
			return err
		}

		_, _, err = s.executor.Execute(relationSQL)
		if err != nil {
			return
		}

		rVal, _ := s.modelProvider.GetEntityValue(relationInfo.Interface(true))
		fVal.Set(rVal.Get())
	}

	return
}

// Insert insert
func (s *impl) Insert(entityModel model.Model) (ret model.Model, err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}

	for {
		err = s.insertSingle(entityModel)
		if err != nil {
			break
		}

		for _, field := range entityModel.GetFields() {
			err = s.insertRelation(entityModel, field)
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

	if err != nil {
		return
	}

	ret = entityModel
	return
}
