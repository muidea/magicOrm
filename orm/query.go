package orm

import (
	"fmt"
	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

type resultItems []any
type resultItemsList []resultItems

func (s *impl) innerQuery(vModel model.Model, filter model.Filter) (ret resultItemsList, err *cd.Result) {
	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builderVal.BuildQuery(filter)
	if sqlErr != nil {
		err = sqlErr
		log.Errorf("innerQuery failed, builder.BuildQuery error:%s", err.Error())
		return
	}

	err = s.executor.Query(sqlStr)
	if err != nil {
		log.Errorf("innerQuery failed, s.executor.Query error:%s", err.Error())
		return
	}

	defer s.executor.Finish()
	for s.executor.Next() {
		itemValues, itemErr := s.getModelFieldsScanDestPtr(vModel, builderVal)
		if itemErr != nil {
			err = itemErr
			if err.Fail() {
				log.Errorf("innerQuery failed, s.getModelFieldsScanDestPtr error:%s", err.Error())
			} else if err.Warn() {
				log.Warnf("innerQuery failed, s.getModelFieldsScanDestPtr error:%s", err.Error())
			}
			return
		}

		err = s.executor.GetField(itemValues...)
		if err != nil {
			log.Errorf("innerQuery failed, s.executor.GetField error:%s", err.Error())
			return
		}

		ret = append(ret, itemValues)
	}

	return
}

func (s *impl) innerAssign(vModel model.Model, queryVal resultItems, deepLevel int) (ret model.Model, err *cd.Result) {
	offset := 0
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if fValue.IsNil() {
			continue
		}

		if !fType.IsBasic() {
			err = s.assignModelField(field, vModel, deepLevel)
			if err != nil {
				if err.Fail() {
					log.Errorf("innerAssign failed, s.assignModelField error:%v", err.Error())
				} else if err.Warn() {
					log.Warnf("innerAssign failed, s.assignModelField error:%v", err.Error())
				}
				return
			}

			//offset++
			continue
		}

		err = s.assignBasicField(field, fType, queryVal[offset])
		if err != nil {
			if err.Fail() {
				log.Errorf("innerAssign failed, s.assignBasicField error:%v", err.Error())
			} else if err.Warn() {
				log.Warnf("innerAssign failed, s.assignBasicField error:%v", err.Error())
			}
			return
		}
		offset++
	}

	ret = vModel
	return
}

func (s *impl) querySingle(vFilter model.Filter, deepLevel int) (ret model.Model, err *cd.Result) {
	vModel := vFilter.MaskModel()
	valueList, queryErr := s.innerQuery(vModel, vFilter)
	if queryErr != nil {
		err = queryErr
		if err.Fail() {
			log.Errorf("querySingle failed, innerQuery error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("querySingle failed, innerQuery error:%v", err.Error())
		}
		return
	}

	resultSize := len(valueList)
	if resultSize != 1 {
		err = cd.NewWarn(cd.Warned, fmt.Sprintf("matched %d values", resultSize))
		log.Warnf("querySingle failed, error:%v", err.Error())
		return
	}

	modelVal, modelErr := s.innerAssign(vModel, valueList[0], deepLevel)
	if modelErr != nil {
		err = modelErr
		if err.Warn() {
			log.Warnf("querySingle failed, s.innerAssign error:%v", err.Error())
		} else {
			log.Errorf("querySingle failed, s.innerAssign error:%v", err.Error())
		}
		return
	}

	ret = modelVal
	return
}

func (s *impl) assignModelField(vField model.Field, vModel model.Model, deepLevel int) (err *cd.Result) {
	vVal, vErr := s.queryRelation(vModel, vField, deepLevel)
	if vErr != nil {
		err = vErr
		if err.Fail() {
			log.Errorf("assignModelField failed, s.queryRelation error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("assignModelField failed, s.queryRelation error:%v", err.Error())
		}
		return
	}

	vField.SetValue(vVal)
	return
}

func (s *impl) assignBasicField(vField model.Field, fType model.Type, val interface{}) (err *cd.Result) {
	fVal, fErr := s.modelProvider.DecodeValue(val, fType)
	if fErr != nil {
		err = fErr
		log.Errorf("assignBasicField failed, s.modelProvider.DecodeValue error:%v", err.Error())
		return
	}

	vField.SetValue(fVal)
	return
}

func (s *impl) innerQueryRelationModel(id any, vModel model.Model, deepLevel int) (ret model.Model, err *cd.Result) {
	pkField := vModel.GetPrimaryField()
	fVal, fErr := s.modelProvider.DecodeValue(id, pkField.GetType())
	if fErr != nil {
		err = fErr
		log.Errorf("innerQueryRelationModel failed, s.modelProvider.DecodeValue error:%v", err.Error())
		return
	}
	vModel.SetFieldValue(pkField.GetName(), fVal)

	vFilter, vErr := s.getModelFilter(vModel, model.LiteView)
	if vErr != nil {
		err = vErr
		if err.Fail() {
			log.Errorf("innerQueryRelationModel failed, getModelFilter error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("innerQueryRelationModel failed, getModelFilter error:%v", err.Error())
		}
		return
	}

	queryVal, queryErr := s.querySingle(vFilter, deepLevel+1)
	if queryErr != nil {
		err = queryErr
		if err.Fail() {
			log.Errorf("innerQueryRelationModel failed, s.querySingle error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("innerQueryRelationModel failed, s.querySingle error:%v", err.Error())
		}
		return
	}

	ret = queryVal
	return
}

func (s *impl) innerQueryRelationSliceModel(ids []any, vModel model.Model, deepLevel int) (ret []model.Model, err *cd.Result) {
	sliceVal := []model.Model{}
	for _, id := range ids {
		svModel := vModel.Copy()
		pkField := svModel.GetPrimaryField()
		fVal, fErr := s.modelProvider.DecodeValue(id, pkField.GetType())
		if fErr != nil {
			err = fErr
			log.Errorf("innerQueryRelationSliceModel failed, s.modelProvider.DecodeValue error:%v", err.Error())
			return
		}
		svModel.SetFieldValue(pkField.GetName(), fVal)

		vFilter, vErr := s.getModelFilter(svModel, model.LiteView)
		if vErr != nil {
			err = vErr
			if err.Fail() {
				log.Errorf("innerQueryRelationSliceModel failed, getModelFilter error:%v", err.Error())
			} else if err.Warn() {
				log.Warnf("innerQueryRelationSliceModel failed, getModelFilter error:%v", err.Error())
			}
			return
		}

		queryVal, queryErr := s.querySingle(vFilter, deepLevel+1)
		if queryErr != nil {
			err = queryErr
			if err.Fail() {
				log.Errorf("innerQueryRelationSliceModel failed, s.querySingle error:%v", err.Error())
			} else if err.Warn() {
				log.Warnf("innerQueryRelationSliceModel failed, s.querySingle error:%v", err.Error())
			}
			return
		}

		if queryVal != nil {
			sliceVal = append(sliceVal, queryVal)
		}
	}

	ret = sliceVal
	return
}

func (s *impl) innerQueryRelationKeys(vModel model.Model, rModel model.Model, vField model.Field) (ret resultItems, err *cd.Result) {
	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationSQL, relationErr := builderVal.BuildQueryRelation(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("innerQueryRelationKeys failed, builder.BuildQueryRelation error:%v", err.Error())
		return
	}

	values := resultItems{}
	func() {
		err = s.executor.Query(relationSQL)
		if err != nil {
			log.Errorf("innerQueryRelationKeys failed, s.executor.Query error:%v", err.Error())
			return
		}
		defer s.executor.Finish()

		for s.executor.Next() {
			itemValue, itemErr := s.getModelPKFieldScanDestPtr(vModel, builderVal)
			if itemErr != nil {
				err = itemErr
				if err.Fail() {
					log.Errorf("innerQueryRelationKeys failed, s.getModelPKFieldScanDestPtr error:%v", err.Error())
				} else if err.Warn() {
					log.Warnf("innerQueryRelationKeys failed, s.getModelPKFieldScanDestPtr error:%v", err.Error())
				}
				return
			}

			err = s.executor.GetField(itemValue)
			if err != nil {
				log.Errorf("innerQueryRelationKeys failed, s.executor.GetField error:%v", err.Error())
				return
			}
			values = append(values, itemValue)
		}
	}()

	if err != nil {
		return
	}

	ret = values
	return
}

func (s *impl) querySingleRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err *cd.Result) {
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("querySingleRelation failed, s.modelProvider.GetTypeModel error:%v", err.Error())
		return
	}

	vType := vField.GetType()
	vValue, _ := vType.Interface(nil)
	valueList, valueErr := s.innerQueryRelationKeys(vModel, rModel, vField)
	if valueErr != nil {
		err = valueErr
		if err.Fail() {
			log.Errorf("querySingleRelation failed, s.innerQueryRelationKeys error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("querySingleRelation failed, s.innerQueryRelationKeys error:%v", err.Error())
		}
		return
	}
	valueSize := len(valueList)
	if valueSize == 0 {
		if vType.IsPtrType() {
			ret = vValue
			return
		}

		err = cd.NewWarn(cd.Warned, fmt.Sprintf("mismatch relation field:%s", vField.GetName()))
		return
	}

	rvModel, rvErr := s.innerQueryRelationModel(valueList[0], rModel, deepLevel)
	if rvErr != nil {
		err = rvErr
		if err.Fail() {
			log.Errorf("querySingleRelation failed, s.innerQueryRelationModel error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("querySingleRelation failed, s.innerQueryRelationModel error:%v", err.Error())
		}
		return
	}

	modelVal, modelErr := s.modelProvider.GetEntityValue(rvModel.Interface(vType.IsPtrType(), model.LiteView))
	if modelErr != nil {
		err = modelErr
		log.Errorf("querySingleRelation failed, s.modelProvider.GetEntityValue error:%v", err.Error())
		return
	}

	vValue.Set(modelVal.Get())
	ret = vValue
	return
}

func (s *impl) querySliceRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err *cd.Result) {
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("querySliceRelation failed, s.modelProvider.GetTypeModel error:%sv", err.Error())
		return
	}

	vType := vField.GetType()
	vValue, _ := vType.Interface(nil)
	valueList, valueErr := s.innerQueryRelationKeys(vModel, rModel, vField)
	if valueErr != nil {
		err = valueErr
		if err.Fail() {
			log.Errorf("querySliceRelation failed, s.innerQueryRelationKeys error:%sv", err.Error())
		} else if err.Warn() {
			log.Warnf("querySliceRelation failed, s.innerQueryRelationKeys error:%sv", err.Error())
		}
		return
	}
	valueSize := len(valueList)
	if valueSize == 0 {
		ret = vValue
		return
	}

	rModelList, rModelErr := s.innerQueryRelationSliceModel(valueList, rModel, deepLevel)
	if rModelErr != nil {
		err = rModelErr
		if err.Fail() {
			log.Errorf("querySliceRelation failed, s.innerQueryRelationSliceModel error:%sv", err.Error())
		} else if err.Warn() {
			log.Warnf("querySliceRelation failed, s.innerQueryRelationSliceModel error:%sv", err.Error())
		}
		return
	}

	elemType := vType.Elem()
	for _, sv := range rModelList {
		modelVal, modelErr := s.modelProvider.GetEntityValue(sv.Interface(elemType.IsPtrType(), model.LiteView))
		if modelErr != nil {
			err = modelErr
			log.Errorf("querySliceRelation failed, s.modelProvider.GetEntityValue error:%sv", err.Error())
			return
		}

		vValue, rErr = s.modelProvider.AppendSliceValue(vValue, modelVal)
		if rErr != nil {
			err = rErr
			log.Errorf("querySliceRelation failed, s.modelProvider.AppendSliceValue error:%sv", err.Error())
			return
		}
	}
	ret = vValue
	return
}

func (s *impl) queryRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err *cd.Result) {
	if deepLevel > maxDeepLevel {
		fieldType := vField.GetType()
		ret, _ = fieldType.Interface(nil)
		return
	}

	if vField.IsSlice() {
		ret, err = s.querySliceRelation(vModel, vField, deepLevel)
		if err != nil {
			if err.Fail() {
				log.Errorf("queryRelation failed, s.querySliceRelation error:%v", err.Error())
			} else if err.Warn() {
				log.Warnf("queryRelation failed, s.querySliceRelation error:%v", err.Error())
			}
		}
		return
	}

	ret, err = s.querySingleRelation(vModel, vField, deepLevel)
	if err != nil {
		if err.Fail() {
			log.Errorf("queryRelation failed, s.querySingleRelation error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("queryRelation failed, s.querySingleRelation error:%v", err.Error())
		}
	}
	return
}

func (s *impl) Query(vModel model.Model) (ret model.Model, err *cd.Result) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	vFilter, vErr := s.getModelFilter(vModel, 0)
	if vErr != nil {
		err = vErr
		if err.Fail() {
			log.Errorf("Query failed, s.getModelFilter error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("Query failed, s.getModelFilter error:%v", err.Error())
		}
		return
	}

	qModel, qErr := s.querySingle(vFilter, 0)
	if qErr != nil {
		err = qErr
		if err.Warn() {
			log.Warnf("Query failed, s.querySingle error:%v", err.Error())
		} else {
			log.Errorf("Query failed, s.querySingle error:%v", err.Error())
		}
		return
	}

	ret = qModel
	return
}
