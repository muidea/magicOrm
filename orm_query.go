package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/local"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

func (s *orm) querySingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo)
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
	for _, val := range *fields {
		fType := val.GetType()

		dependType, _ := fType.Depend()
		if dependType != nil {
			continue
		}

		v := util.GetBasicTypeInitValue(fType.Value())
		items = append(items, v)
	}
	s.executor.GetField(items...)

	idx := 0
	for _, val := range *fields {
		fType := val.GetType()

		dependType, _ := fType.Depend()
		if dependType != nil {
			continue
		}

		v := items[idx]
		err = val.SetValue(reflect.Indirect(reflect.ValueOf(v)))
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

	builder := builder.NewBuilder(modelInfo)
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
			fDepend, _ := fType.Depend()
			relationVal := reflect.New(fDepend)
			relationInfo, relationErr = local.GetValueModel(relationVal, s.modelInfoCache)
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
		relationVal, _ := fValue.Get()
		relationType := relationVal.Type()
		if fType.IsPtr() {
			relationType = relationType.Elem()
		}

		relationVal = reflect.New(relationType).Elem()
		for _, val := range values {
			fDepend, fDependPtr := fType.Depend()
			itemVal := reflect.New(fDepend)
			itemInfo, itemErr := local.GetValueModel(itemVal, s.modelInfoCache)
			if itemErr != nil {
				log.Printf("GetValueModel faield, err:%s", itemErr.Error())
				err = itemErr
				return
			}

			itemInfo.GetPrimaryField().SetValue(reflect.ValueOf(val))
			err = s.querySingle(itemInfo)
			if err != nil {
				return
			}

			if !fDependPtr {
				itemVal = reflect.Indirect(itemVal)
			}

			relationVal = reflect.Append(relationVal, itemVal)
		}
		modelInfo.UpdateFieldValue(fieldInfo.GetName(), relationVal)
	}

	return
}

func (s *orm) Query(obj interface{}) (err error) {
	modelInfo, structErr := local.GetObjectModel(obj, s.modelInfoCache)
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
	for _, val := range fields {
		fType := val.GetType()
		fDepend, _ := fType.Depend()

		if fDepend == nil {
			continue
		}

		infoVal, infoErr := local.GetTypeModel(fDepend, s.modelInfoCache)
		if infoErr != nil {
			err = infoErr
			return
		}
		err = s.queryRelation(modelInfo, val, infoVal)
		if err != nil {
			return
		}
	}

	return
}
