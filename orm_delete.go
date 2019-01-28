package orm

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/builder"
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

func (s *orm) deleteRelation(modelInfo model.Model, fieldInfo model.FieldInfo) (err error) {
	fType := fieldInfo.GetFieldType()
	fDepend, fDependPtr := fType.Depend()

	if fDepend == nil {
		return
	}

	infoVal, infoErr := model.GetStructInfo(fDepend, s.modelInfoCache)
	if infoErr != nil {
		err = infoErr
		return
	}

	builder := builder.NewBuilder(modelInfo)
	rightSQL, relationSQL, err := builder.BuildDeleteRelation(fieldInfo.GetFieldName(), infoVal)
	if err != nil {
		return err
	}

	if !fDependPtr {
		s.executor.Delete(rightSQL)
	}

	s.executor.Delete(relationSQL)

	return
}

func (s *orm) Delete(obj interface{}) (err error) {
	modelInfo, structErr := model.GetObjectStructInfo(obj, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectStructInfo failed, err:%s", err.Error())
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
