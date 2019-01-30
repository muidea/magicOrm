package orm

import (
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
)

func (s *orm) insertSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildInsert()
	if err != nil {
		return err
	}

	id := s.executor.Insert(sql)
	pk := modelInfo.GetPrimaryField()
	if pk != nil {
		pk.SetValue(reflect.ValueOf(id))
	}

	return
}

func (s *orm) insertRelation(modelInfo model.Model, fieldInfo model.Field) (err error) {
	fType := fieldInfo.GetType()
	fDepend := fType.Depend()

	fValue := fieldInfo.GetValue()
	if fValue == nil {
		return
	}

	fDependValue, fDependErr := fValue.Depend()
	if fDependErr != nil {
		err = fDependErr
		return
	}

	for _, fVal := range fDependValue {
		infoVal, infoErr := s.modelProvider.GetValueModel(fVal)
		if infoErr != nil {
			log.Printf("GetValueModel faield, err:%s", infoErr.Error())
			err = infoErr
			return
		}

		if !fDepend.IsPtr() {
			err = s.insertSingle(infoVal)
			if err != nil {
				return
			}
		}

		builder := builder.NewBuilder(modelInfo, s.modelProvider)
		relationSQL, relationErr := builder.BuildInsertRelation(fieldInfo.GetName(), infoVal)
		if relationErr != nil {
			err = relationErr
			return err
		}

		s.executor.Insert(relationSQL)
	}

	return
}

func (s *orm) Insert(obj interface{}) (err error) {
	modelInfo, modelErr := s.modelProvider.GetObjectModel(obj)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetObjectModel failed, err:%s", err.Error())
		return
	}

	err = s.insertSingle(modelInfo)
	if err != nil {
		log.Printf("insertSingle failed, name:%s, err:%s", modelInfo.GetName(), err.Error())
		return
	}

	fields := modelInfo.GetDependField()
	for _, val := range fields {
		err = s.insertRelation(modelInfo, val)
		if err != nil {
			return
		}
	}

	return
}
