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
		pkv := pk.GetValue()
		err = pkv.Set(reflect.ValueOf(id))
	}

	return
}

func (s *orm) insertRelation(modelInfo model.Model, fieldInfo model.Field) (err error) {
	fType := fieldInfo.GetType()
	fDependModel, fDependErr := s.modelProvider.GetTypeModel(fType.GetType())
	if fDependErr != nil {
		err = fDependErr
		return
	}
	if fDependModel == nil {
		return
	}

	fDependValue, fDependErr := s.modelProvider.GetModelDependValue(fDependModel, fieldInfo.GetValue())
	if fDependErr != nil {
		err = fDependErr
		return
	}

	for _, fVal := range fDependValue {
		relationInfo, relationErr := s.modelProvider.GetValueModel(fVal)
		if relationErr != nil {
			log.Printf("GetValueModel faield, err:%s", relationErr.Error())
			err = relationErr
			return
		}

		if !fType.IsPtrType() {
			err = s.insertSingle(relationInfo)
			if err != nil {
				return
			}
		}

		builder := builder.NewBuilder(modelInfo, s.modelProvider)
		relationSQL, relationErr := builder.BuildInsertRelation(fieldInfo.GetName(), relationInfo)
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

	for _, field := range modelInfo.GetFields() {
		err = s.insertRelation(modelInfo, field)
		if err != nil {
			return
		}
	}

	return
}
