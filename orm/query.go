package orm

import (
	"fmt"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *Orm) querySingle(vModel model.Model, vVal model.Value, filter model.Filter) (err error) {
	func() {
		builder := builder.NewBuilder(vModel, s.modelProvider)
		sqlStr, sqlErr := builder.BuildQuery(filter)
		if sqlErr != nil {
			err = sqlErr
			log.Errorf("build query failed, err:%s", err.Error())
			return
		}

		err = s.executor.Query(sqlStr)
		if err != nil {
			return
		}

		defer s.executor.Finish()
		if !s.executor.Next() {
			err = fmt.Errorf("query %s failed, no found object", vModel.GetName())
			return
		}

		items, itemErr := s.getModelItems(vModel, builder)
		if itemErr != nil {
			err = itemErr
			return
		}

		err = s.executor.GetField(items...)
		if err != nil {
			return
		}

		for idx, item := range vModel.GetFields() {
			fType := item.GetType()
			if item.GetValue().IsNil() || !fType.IsBasic() {
				continue
			}

			vVal := fType.Interface(s.stripSlashes(fType, items[idx]))
			err = item.SetValue(vVal)
			if err != nil {
				return
			}
		}
	}()
	if err != nil {
		return
	}

	for _, item := range vModel.GetFields() {
		fType := item.GetType()
		if item.GetValue().IsNil() || fType.IsBasic() {
			continue
		}

		itemVal, itemErr := s.queryRelation(vModel, item)
		if itemErr != nil {
			err = itemErr
			log.Errorf("queryRelation failed, modelName:%s, fieldName:%s, err:%s", vModel.GetName(), item.GetName(), err.Error())
			return
		}

		itemErr = item.SetValue(itemVal)
		if itemErr != nil {
			err = itemErr
			log.Errorf("UpdateFieldValue failed, modelName:%s, fieldName:%s, err:%s", vModel.GetName(), item.GetName(), err.Error())
			return
		}
	}

	vVal.Set(vModel.Interface().Get())

	return
}

func (s *Orm) queryRelationSingle(id int, vModel model.Model) (ret model.Value, err error) {
	relationModel := vModel.Copy()
	relationVal, relationErr := s.modelProvider.GetEntityValue(id)
	if relationErr != nil {
		err = fmt.Errorf("GetEntityValue failed, err:%s", relationErr)
		return
	}

	pkField := relationModel.GetPrimaryField()
	pkField.SetValue(relationVal)
	relationFilter, relationErr := s.getFieldFilter(pkField)
	if relationErr != nil {
		err = fmt.Errorf("GetEntityValue failed, err:%s", relationErr)
		return
	}

	err = s.querySingle(relationModel, relationVal, relationFilter)
	if err != nil {
		log.Errorf("querySingle for struct failed, err:%s", err.Error())
		return
	}

	ret = relationModel.Interface()
	return
}

func (s *Orm) queryRelationSlice(ids []int, vModel model.Model, sliceVal model.Value) (ret model.Value, err error) {
	for _, item := range ids {
		relationModel := vModel.Copy()
		relationVal, relationErr := s.modelProvider.GetEntityValue(item)
		if relationErr != nil {
			err = fmt.Errorf("GetEntityValue failed, err:%s", relationErr)
			return
		}

		pkField := relationModel.GetPrimaryField()
		pkField.SetValue(relationVal)
		relationFilter, relationErr := s.getFieldFilter(pkField)
		if relationErr != nil {
			err = fmt.Errorf("GetEntityValue failed, err:%s", relationErr)
			return
		}

		err = s.querySingle(relationModel, relationVal, relationFilter)
		if err != nil {
			log.Errorf("querySingle for slice failed, err:%s", err.Error())
			return
		}

		itemVal := relationModel.Interface()
		sliceVal, err = s.modelProvider.AppendSliceValue(sliceVal, itemVal)
		if err != nil {
			log.Errorf("append slice value failed, err:%s", err.Error())
			return
		}
	}

	ret = sliceVal
	return
}

func (s *Orm) queryRelation(modelInfo model.Model, fieldInfo model.Field) (ret model.Value, err error) {
	fieldType := fieldInfo.GetType()
	fieldModel, fieldErr := s.modelProvider.GetTypeModel(fieldType)
	if fieldErr != nil {
		err = fieldErr
		log.Errorf("GetTypeModel failed, type:%s, err:%s", fieldType.GetName(), err.Error())
		return
	}

	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	relationSQL, relationErr := builder.BuildQueryRelation(fieldInfo.GetName(), fieldModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildQueryRelation failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
		return
	}

	var values []int
	func() {
		err = s.executor.Query(relationSQL)
		if err != nil {
			return
		}

		defer s.executor.Finish()
		for s.executor.Next() {
			v := 0
			err = s.executor.GetField(&v)
			if err != nil {
				return
			}
			values = append(values, v)
		}
	}()
	if err != nil || len(values) == 0 {
		return
	}

	if util.IsStructType(fieldType.GetValue()) {
		singleVal, singleErr := s.queryRelationSingle(values[0], fieldModel)
		if singleErr != nil {
			err = singleErr
			log.Errorf("queryRelationSingle failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
			return
		}

		if fieldType.IsPtrType() {
			singleVal = singleVal.Addr()
		}
		ret = singleVal
	} else if util.IsSliceType(fieldType.GetValue()) {
		sliceVal := fieldType.Interface(nil)
		sliceVal, sliceErr := s.queryRelationSlice(values, fieldModel, sliceVal)
		if sliceErr != nil {
			err = sliceErr
			log.Errorf("queryRelationSlice failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
			return
		}

		if fieldType.IsPtrType() {
			sliceVal = sliceVal.Addr()
		}
		ret = sliceVal
	}

	return
}

// Query query
func (s *Orm) Query(entity interface{}) (err error) {
	entityModel, entityErr := s.modelProvider.GetEntityModel(entity)
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	entityVal, entityErr := s.modelProvider.GetEntityValue(entity)
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetEntityValue failed, err:%s", err.Error())
		return
	}

	entityFilter, entityErr := s.getModelFilter(entityModel)
	if entityErr != nil {
		err = entityErr
		log.Errorf("getFilter failed, err:%s", err.Error())
		return
	}

	err = s.querySingle(entityModel, entityVal, entityFilter)
	if err != nil {
		log.Errorf("querySingle failed, modelName:%s, err:%s", entityModel.GetName(), err.Error())
		return
	}

	return
}
