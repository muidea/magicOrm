package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
)

func (s *orm) deleteSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildDelete()
	if err != nil {
		return err
	}
	num := s.executor.Delete(sql)
	if num != 1 {
		log.Printf("unexception delete, rowNum:%d", num)
		err = fmt.Errorf("delete %s failed", modelInfo.GetName())
	}

	return
}

func (s *orm) deleteRelation(modelInfo model.Model, fieldInfo model.Field) (err error) {
	fType := fieldInfo.GetType()
	fDepend := fType.GetDepend()

	if fDepend == nil {
		return
	}

	relationInfo, relationErr := s.modelProvider.GetTypeModel(fDepend)
	if relationErr != nil {
		err = relationErr
		return
	}

	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	rightSQL, relationSQL, relationErr := builder.BuildDeleteRelation(fieldInfo.GetName(), relationInfo)
	if relationErr != nil {
		err = relationErr
		return
	}

	if fDepend.Kind() != reflect.Ptr {
		s.executor.Delete(rightSQL)
	}

	s.executor.Delete(relationSQL)

	return
}

func (s *orm) Delete(obj interface{}) (err error) {
	modelInfo, modelErr := s.modelProvider.GetObjectModel(obj)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetObjectModel failed, err:%s", err.Error())
		return
	}

	err = s.deleteSingle(modelInfo)
	if err != nil {
		return
	}

	fields := modelInfo.GetDependField()
	for _, field := range fields {
		err = s.deleteRelation(modelInfo, field)
		if err != nil {
			return
		}
	}

	return
}
