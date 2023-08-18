package orm

import (
	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) innerInsert(vModel model.Model) (ret interface{}, err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builder.BuildInsert()
	if sqlErr != nil {
		err = sqlErr
		log.Errorf("build insert sql failed, err:%s", err.Error())
		return
	}

	_, id, idErr := s.executor.Execute(sqlStr)
	if idErr != nil {
		err = idErr
		return
	}

	ret = id
	return
}

func (s *impl) insertSingle(vModel model.Model) (ret model.Model, err error) {
	autoIncrementFlag := false
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		if !fType.IsBasic() {
			continue
		}

		fSpec := field.GetSpec()
		fValue := field.GetValue()
		if fValue.IsZero() {
			fValue = s.modelProvider.GetValue(fSpec.GetValueDeclare())
			if !fValue.IsNil() {
				field.SetValue(fValue)
			}
		}
		if fSpec.GetValueDeclare() == model.AutoIncrement {
			autoIncrementFlag = true
		}
	}

	insertVal, insertErr := s.innerInsert(vModel)
	if insertErr != nil {
		err = insertErr
		return
	}

	if insertVal != nil && autoIncrementFlag {
		pkField := vModel.GetPrimaryField()
		tVal, tErr := s.modelProvider.DecodeValue(insertVal, pkField.GetType())
		if tErr != nil {
			err = tErr
			return
		}

		err = pkField.SetValue(tVal)
		if err != nil {
			return
		}
	}

	ret = vModel
	return
}

func (s *impl) insertRelation(vModel model.Model, vField model.Field) (err error) {
	fValue := vField.GetValue()
	fType := vField.GetType()
	if fType.IsBasic() || fValue.IsNil() /* || !s.modelProvider.IsAssigned(fValue, fType)*/ {
		return
	}

	fSliceValue, fSliceErr := s.modelProvider.ElemDependValue(fValue)
	if fSliceErr != nil {
		err = fSliceErr
		return
	}

	for _, fVal := range fSliceValue {
		rModel, rErr := s.modelProvider.GetValueModel(fVal, fType)
		if rErr != nil {
			err = rErr
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			rModel, rErr = s.insertSingle(rModel)
			if rErr != nil {
				err = rErr
				return
			}

			for _, subField := range rModel.GetFields() {
				err = s.insertRelation(rModel, subField)
				if err != nil {
					return
				}
			}
		}

		builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
		relationSQL, relationErr := builder.BuildInsertRelation(vField, rModel)
		if relationErr != nil {
			err = relationErr
			return err
		}

		_, _, err = s.executor.Execute(relationSQL)
		if err != nil {
			return
		}

		rVal, _ := s.modelProvider.GetEntityValue(rModel.Interface(true))
		fVal.Set(rVal.Get())
	}

	return
}

// Insert insert
func (s *impl) Insert(vModel model.Model) (ret model.Model, err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	insertVal, insertErr := s.insertSingle(vModel)
	if insertErr != nil {
		err = insertErr
		return
	}

	for _, field := range insertVal.GetFields() {
		err = s.insertRelation(insertVal, field)
		if err != nil {
			break
		}
	}

	if err != nil {
		return
	}

	ret = insertVal
	return
}
