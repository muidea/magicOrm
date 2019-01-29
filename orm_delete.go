package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/local"
	"muidea.com/magicOrm/model"
)

func (s *orm) deleteSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo)
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
	fDepend := fType.Depend()

	if fDepend == nil {
		return
	}

	infoVal, infoErr := local.GetTypeModel(fDepend, s.modelInfoCache)
	if infoErr != nil {
		err = infoErr
		return
	}

	builder := builder.NewBuilder(modelInfo)
	rightSQL, relationSQL, err := builder.BuildDeleteRelation(fieldInfo.GetName(), infoVal)
	if err != nil {
		return err
	}

	if fDepend.Kind() != reflect.Ptr {
		s.executor.Delete(rightSQL)
	}

	s.executor.Delete(relationSQL)

	return
}

func (s *orm) Delete(obj interface{}) (err error) {
	modelInfo, structErr := local.GetObjectModel(obj, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectModel failed, err:%s", err.Error())
		return
	}

	err = s.deleteSingle(modelInfo)
	if err != nil {
		return
	}

	fields := modelInfo.GetDependField()
	for _, val := range fields {
		err = s.deleteRelation(modelInfo, val)
		if err != nil {
			return
		}
	}

	return
}
