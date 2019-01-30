package orm

import (
	"log"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
)

func (s *orm) updateSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildUpdate()
	if err != nil {
		return err
	}

	s.executor.Update(sql)

	return err
}

func (s *orm) updateRelation(modelInfo model.Model, fieldInfo model.Field) (err error) {
	err = s.deleteRelation(modelInfo, fieldInfo)
	if err != nil {
		return
	}

	err = s.insertRelation(modelInfo, fieldInfo)
	if err != nil {
		return
	}

	return
}

func (s *orm) Update(obj interface{}) (err error) {
	modelInfo, modelErr := s.modelProvider.GetObjectModel(obj)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetObjectModel failed, err:%s", err.Error())
		return
	}

	err = s.updateSingle(modelInfo)
	if err != nil {
		return
	}

	fields := modelInfo.GetDependField()
	for _, field := range fields {
		err = s.updateRelation(modelInfo, field)
		if err != nil {
			return
		}
	}

	return
}
