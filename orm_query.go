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
	for _, item := range *fields {
		fType := item.GetType()

		dependType := fType.Depend()
		if dependType != nil {
			continue
		}

		itemVal, itemErr := util.GetBasicTypeInitValue(fType.Value())
		if itemErr != nil {
			err = itemErr
			return
		}

		items = append(items, itemVal)
	}
	s.executor.GetField(items...)

	idx := 0
	for _, item := range *fields {
		fType := item.GetType()

		dependType := fType.Depend()
		if dependType != nil {
			continue
		}

		v := items[idx]
		err = item.SetValue(reflect.Indirect(reflect.ValueOf(v)))
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

	if util.IsStructType(fType.Value()) {
		if len(values) > 0 {
			fDepend := fType.Depend()
			fDependType := fDepend.Type()
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

			modelInfo.UpdateFieldValue(fieldInfo.GetName(), relationVal)
		}
	} else if util.IsSliceType(fType.Value()) {
		relationType := fType.Type()
		relationVal := reflect.New(relationType).Elem()
		for _, item := range values {
			fDepend := fType.Depend()
			fDependType := fDepend.Type()
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

			if !fDepend.IsPtr() {
				itemVal = reflect.Indirect(itemVal)
			}

			relationVal = reflect.Append(relationVal, itemVal)
		}
		modelInfo.UpdateFieldValue(fieldInfo.GetName(), relationVal)
	}

	return
}

func (s *orm) Query(obj interface{}) (err error) {
	modelInfo, structErr := s.modelProvider.GetObjectModel(obj)
	if structErr != nil {
		err = structErr
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
		fDepend := fType.Depend()

		if fDepend == nil {
			continue
		}

		infoVal, infoErr := s.modelProvider.GetTypeModel(fDepend.Type())
		if infoErr != nil {
			err = infoErr
			return
		}
		err = s.queryRelation(modelInfo, item, infoVal)
		if err != nil {
			return
		}
	}

	return
}
