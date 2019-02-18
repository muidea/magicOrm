package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

func (s *orm) querySingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildQuery()
	if err != nil {
		return err
	}

	s.executor.Query(sql)
	if !s.executor.Next() {
		return fmt.Errorf("no found object")
	}
	defer s.executor.Finish()

	items := []interface{}{}
	fields := modelInfo.GetFields()
	for _, item := range fields {
		fType := item.GetType()

		dependModel, dependErr := s.modelProvider.GetTypeModel(fType.GetType())
		if dependErr != nil {
			err = dependErr
			return
		}
		if dependModel != nil {
			continue
		}

		itemVal, itemErr := util.GetBasicTypeInitValue(fType.GetValue())
		if itemErr != nil {
			err = itemErr
			return
		}

		items = append(items, itemVal)
	}
	s.executor.GetField(items...)

	idx := 0
	for _, item := range fields {
		fType := item.GetType()
		fValue := item.GetValue()

		dependModel, dependErr := s.modelProvider.GetTypeModel(fType.GetType())
		if dependErr != nil {
			err = dependErr
			return
		}
		if dependModel != nil {
			continue
		}

		v := items[idx]
		err = fValue.Set(reflect.Indirect(reflect.ValueOf(v)))
		if err != nil {
			return err
		}

		idx++
	}

	return
}

func (s *orm) queryRelation(modelInfo model.Model, fieldInfo model.Field, relationInfo model.Model) (err error) {
	fValue := fieldInfo.GetValue()
	if fValue == nil || fValue.IsNil() {
		return
	}

	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	relationSQL, relationErr := builder.BuildQueryRelation(fieldInfo.GetName(), relationInfo)
	if relationErr != nil {
		err = relationErr
		return err
	}

	fType := fieldInfo.GetType()
	values := []int{}

	func() {
		s.executor.Query(relationSQL)
		defer s.executor.Finish()
		for s.executor.Next() {
			v := 0
			s.executor.GetField(&v)
			values = append(values, v)
		}
	}()

	if util.IsStructType(fType.GetValue()) {
		if len(values) > 0 {

			dependModel, dependErr := s.modelProvider.GetTypeModel(fType.GetType())
			if dependErr != nil {
				err = dependErr
				return
			}
			if dependModel != nil {
				continue
			}
			fDepend := fType.GetDepend()
			fDependType := fDepend
			if fDependType.Kind() == reflect.Ptr {
				fDependType = fDependType.Elem()
			}
			relationVal := reflect.New(fDependType)
			relationInfo, relationErr = s.modelProvider.GetValueModel(relationVal)
			if relationErr != nil {
				err = relationErr
				return
			}

			relationInfo.GetPrimaryField().SetValue(reflect.ValueOf(values[0]))
			err = s.querySingle(relationInfo)
			if err != nil {
				return
			}

			err = modelInfo.UpdateFieldValue(fieldInfo.GetName(), relationVal)
			if err != nil {
				return
			}
		}
	} else if util.IsSliceType(fType.GetValue()) {
		relationType := fType.GetType()
		relationVal := reflect.New(relationType).Elem()
		for _, item := range values {
			fDepend := fType.GetDepend()
			fDependType := fDepend
			if fDependType.Kind() == reflect.Ptr {
				fDependType = fDependType.Elem()
			}
			itemVal := reflect.New(fDependType)
			itemInfo, itemErr := s.modelProvider.GetValueModel(itemVal)
			if itemErr != nil {
				log.Printf("GetValueModel faield, err:%s", itemErr.Error())
				err = itemErr
				return
			}

			itemInfo.GetPrimaryField().SetValue(reflect.ValueOf(item))
			err = s.querySingle(itemInfo)
			if err != nil {
				return
			}

			if fDepend.Kind() != reflect.Ptr {
				itemVal = reflect.Indirect(itemVal)
			}

			relationVal = reflect.Append(relationVal, itemVal)
		}
		err = modelInfo.UpdateFieldValue(fieldInfo.GetName(), relationVal)
		if err != nil {
			return
		}
	}

	return
}

func (s *orm) Query(obj interface{}) (err error) {
	modelInfo, modelErr := s.modelProvider.GetObjectModel(obj)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetObjectModel failed, err:%s", err.Error())
		return
	}

	err = s.querySingle(modelInfo)
	if err != nil {
		return
	}

	fields := modelInfo.GetDependField()
	for _, item := range fields {
		fType := item.GetType()

		relationInfo, relationErr := s.modelProvider.GetTypeModel(fType.GetType())
		if relationErr != nil {
			err = relationErr
			return
		}
		if relationInfo != nil {
			continue
		}
		err = s.queryRelation(modelInfo, item, relationInfo)
		if err != nil {
			return
		}
	}

	return
}
