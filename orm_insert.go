package orm

import (
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
)

func (s *orm) insertSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo)
	sql, err := builder.BuildInsert()
	if err != nil {
		return err
	}

	id := s.executor.Insert(sql)
	pk := modelInfo.GetPrimaryField()
	if pk != nil {
		pk.SetFieldValue(reflect.ValueOf(id))
	}

	return
}

func (s *orm) insertRelation(modelInfo model.Model, fieldInfo model.FieldInfo) (err error) {
	fType := fieldInfo.GetFieldType()
	_, fDependPtr := fType.Depend()

	fValue := fieldInfo.GetFieldValue()
	if fValue == nil {
		return
	}

	fDependValue, fDependErr := fValue.GetDepend()
	if fDependErr != nil {
		err = fDependErr
		return
	}

	for _, fVal := range fDependValue {
		infoVal, infoErr := model.GetStructValue(fVal, s.modelInfoCache)
		if infoErr != nil {
			log.Printf("GetStructValue faield, err:%s", infoErr.Error())
			err = infoErr
			return
		}

		if !fDependPtr {
			err = s.insertSingle(infoVal)
			if err != nil {
				return
			}
		}

		builder := builder.NewBuilder(modelInfo)
		relationSQL, relationErr := builder.BuildInsertRelation(fieldInfo.GetFieldName(), infoVal)
		if relationErr != nil {
			err = relationErr
			return err
		}

		s.executor.Insert(relationSQL)
	}

	return
}

func (s *orm) Insert(obj interface{}) (err error) {
	modelInfo, structErr := model.GetObjectStructInfo(obj, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectStructInfo failed, err:%s", err.Error())
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
