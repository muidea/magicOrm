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

func (s *impl) innerQuery(vModel model.Model, filter model.Filter) (ret resultItemsList, re *cd.Result) {
	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builderVal.BuildQuery(filter)
	if sqlErr != nil {
		re = cd.NewError(cd.UnExpected, sqlErr.Error())
		log.Errorf("innerQuery failed, builder.BuildQuery error:%s", sqlErr.Error())
		return
	}

	err := s.executor.Query(sqlStr)
	if err != nil {
		re = cd.NewError(cd.UnExpected, err.Error())
		log.Errorf("innerQuery failed, s.executor.Query error:%s", err.Error())
		return
	}
	defer s.executor.Finish()
	for s.executor.Next() {
		itemValues, itemErr := s.getModelFieldsScanDestPtr(vModel, builderVal)
		if itemErr != nil {
			re = itemErr
			if itemErr.Fail() {
				log.Errorf("innerQuery failed, s.getModelFieldsScanDestPtr error:%s", itemErr.Error())
			} else if itemErr.Warn() {
				log.Warnf("innerQuery failed, s.getModelFieldsScanDestPtr error:%s", itemErr.Error())
			}
			return
		}

		err = s.executor.GetField(itemValues...)
		if err != nil {
			re = cd.NewError(cd.UnExpected, err.Error())
			log.Errorf("innerQuery failed, s.executor.GetField error:%s", err.Error())
			return
		}

		ret = append(ret, itemValues)
	}

	return
}

func (s *impl) innerAssign(vModel model.Model, queryVal resultItems, deepLevel int) (ret model.Model, re *cd.Result) {
	offset := 0
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if fValue.IsNil() {
			continue
		}

		if !fType.IsBasic() {
			err := s.assignModelField(field, vModel, deepLevel)
			if err != nil {
				re = err
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

		err := s.assignBasicField(field, fType, queryVal[offset])
		if err != nil {
			re = err
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

func (s *impl) querySingle(vFilter model.Filter, deepLevel int) (ret model.Model, re *cd.Result) {
	vModel := vFilter.MaskModel()
	valueList, queryErr := s.innerQuery(vModel, vFilter)
	if queryErr != nil {
		re = queryErr
		if queryErr.Fail() {
			log.Errorf("querySingle failed, innerQuery error:%v", queryErr.Error())
		} else if queryErr.Warn() {
			log.Warnf("querySingle failed, innerQuery error:%v", queryErr.Error())
		}
		return
	}

	resultSize := len(valueList)
	if resultSize != 1 {
		re = cd.NewWarn(cd.Warned, fmt.Sprintf("matched %d values", resultSize))
		log.Warnf("querySingle failed, error:%v", re.Error())
		return
	}

	modelVal, modelErr := s.innerAssign(vModel, valueList[0], deepLevel)
	if modelErr != nil {
		re = modelErr
		if modelErr.Fail() {
			log.Errorf("querySingle failed, s.innerAssign error:%v", modelErr.Error())
		} else if modelErr.Warn() {
			log.Warnf("querySingle failed, s.innerAssign error:%v", modelErr.Error())
		}
		return
	}

	ret = modelVal
	return
}

func (s *impl) assignModelField(vField model.Field, vModel model.Model, deepLevel int) (re *cd.Result) {
	vVal, vErr := s.queryRelation(vModel, vField, deepLevel)
	if vErr != nil {
		re = vErr
		if vErr.Fail() {
			log.Errorf("assignModelField failed, s.queryRelation error:%v", vErr.Error())
		} else if vErr.Warn() {
			log.Warnf("assignModelField failed, s.queryRelation error:%v", vErr.Error())
		}
		return
	}

	vField.SetValue(vVal)
	return
}

func (s *impl) assignBasicField(vField model.Field, fType model.Type, val interface{}) (re *cd.Result) {
	fVal, fErr := s.modelProvider.DecodeValue(val, fType)
	if fErr != nil {
		re = cd.NewError(cd.UnExpected, fErr.Error())
		log.Errorf("assignBasicField failed, s.modelProvider.DecodeValue error:%v", fErr.Error())
		return
	}

	vField.SetValue(fVal)
	return
}

func (s *impl) innerQueryRelationModel(id any, vModel model.Model, deepLevel int) (ret model.Model, re *cd.Result) {
	pkField := vModel.GetPrimaryField()
	fVal, fErr := s.modelProvider.DecodeValue(id, pkField.GetType())
	if fErr != nil {
		re = cd.NewError(cd.UnExpected, fErr.Error())
		log.Errorf("innerQueryRelationModel failed, s.modelProvider.DecodeValue error:%v", fErr.Error())
		return
	}
	_ = vModel.SetFieldValue(pkField.GetName(), fVal)

	vFilter, vErr := s.getModelFilter(vModel)
	if vErr != nil {
		re = vErr
		if vErr.Fail() {
			log.Errorf("innerQueryRelationModel failed, getModelFilter error:%v", vErr.Error())
		} else if vErr.Warn() {
			log.Warnf("innerQueryRelationModel failed, getModelFilter error:%v", vErr.Error())
		}
		return
	}

	queryVal, queryErr := s.querySingle(vFilter, deepLevel+1)
	if queryErr != nil {
		re = queryErr
		if queryErr.Fail() {
			log.Errorf("innerQueryRelationModel failed, s.querySingle error:%v", queryErr.Error())
		} else if queryErr.Warn() {
			log.Warnf("innerQueryRelationModel failed, s.querySingle error:%v", queryErr.Error())
		}
		return
	}

	ret = queryVal
	return
}

func (s *impl) innerQueryRelationSliceModel(ids []any, vModel model.Model, deepLevel int) (ret []model.Model, re *cd.Result) {
	sliceVal := []model.Model{}
	for _, id := range ids {
		svModel := vModel.Copy()
		pkField := svModel.GetPrimaryField()
		fVal, fErr := s.modelProvider.DecodeValue(id, pkField.GetType())
		if fErr != nil {
			re = cd.NewError(cd.UnExpected, fErr.Error())
			log.Errorf("innerQueryRelationSliceModel failed, s.modelProvider.DecodeValue error:%v", fErr.Error())
			return
		}
		_ = svModel.SetFieldValue(pkField.GetName(), fVal)

		vFilter, vErr := s.getModelFilter(svModel)
		if vErr != nil {
			re = vErr
			if vErr.Fail() {
				log.Errorf("innerQueryRelationSliceModel failed, getModelFilter error:%v", vErr.Error())
			} else if vErr.Warn() {
				log.Warnf("innerQueryRelationSliceModel failed, getModelFilter error:%v", vErr.Error())
			}
			return
		}

		queryVal, queryErr := s.querySingle(vFilter, deepLevel+1)
		if queryErr != nil {
			re = queryErr
			if queryErr.Fail() {
				log.Errorf("innerQueryRelationSliceModel failed, s.querySingle error:%v", queryErr.Error())
			} else if queryErr.Warn() {
				log.Warnf("innerQueryRelationSliceModel failed, s.querySingle error:%v", queryErr.Error())
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

func (s *impl) innerQueryRelationKeys(vModel model.Model, rModel model.Model, vField model.Field) (ret resultItems, re *cd.Result) {
	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationSQL, relationErr := builderVal.BuildQueryRelation(vField, rModel)
	if relationErr != nil {
		re = cd.NewError(cd.UnExpected, relationErr.Error())
		log.Errorf("innerQueryRelationKeys failed, builder.BuildQueryRelation error:%v", relationErr.Error())
		return
	}

	values := resultItems{}
	func() {
		err := s.executor.Query(relationSQL)
		if err != nil {
			re = cd.NewError(cd.UnExpected, err.Error())
			log.Errorf("innerQueryRelationKeys failed, s.executor.Query error:%v", err.Error())
			return
		}

		defer s.executor.Finish()
		for s.executor.Next() {
			itemValue, itemErr := s.getModelPKFieldScanDestPtr(vModel, builderVal)
			if itemErr != nil {
				re = itemErr
				if itemErr.Fail() {
					log.Errorf("innerQueryRelationKeys failed, s.getModelPKFieldScanDestPtr error:%v", itemErr.Error())
				} else if itemErr.Warn() {
					log.Warnf("innerQueryRelationKeys failed, s.getModelPKFieldScanDestPtr error:%v", itemErr.Error())
				}
				return
			}

			err = s.executor.GetField(itemValue)
			if err != nil {
				re = cd.NewError(cd.UnExpected, err.Error())
				log.Errorf("innerQueryRelationKeys failed, s.executor.GetField error:%v", err.Error())
				return
			}
			values = append(values, itemValue)
		}
	}()

	if re != nil {
		return
	}

	ret = values
	return
}

func (s *impl) querySingleRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, re *cd.Result) {
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		re = cd.NewError(cd.UnExpected, rErr.Error())
		log.Errorf("querySingleRelation failed, s.modelProvider.GetTypeModel error:%v", rErr.Error())
		return
	}

	vType := vField.GetType()
	vValue, _ := vType.Interface(nil)
	valueList, valueErr := s.innerQueryRelationKeys(vModel, rModel, vField)
	if valueErr != nil {
		re = valueErr
		if valueErr.Fail() {
			log.Errorf("querySingleRelation failed, s.innerQueryRelationKeys error:%v", valueErr.Error())
		} else if valueErr.Warn() {
			log.Warnf("querySingleRelation failed, s.innerQueryRelationKeys error:%v", valueErr.Error())
		}
		return
	}
	valueSize := len(valueList)
	if valueSize == 0 {
		if vType.IsPtrType() {
			ret = vValue
			return
		}

		re = cd.NewWarn(cd.Warned, fmt.Sprintf("mismatch relation field:%s", vField.GetName()))
		return
	}

	rvModel, rvErr := s.innerQueryRelationModel(valueList[0], rModel, deepLevel)
	if rvErr != nil {
		re = rvErr
		if rvErr.Fail() {
			log.Errorf("querySingleRelation failed, s.innerQueryRelationModel error:%v", rvErr.Error())
		}
		if rvErr.Warn() {
			log.Warnf("querySingleRelation failed, s.innerQueryRelationModel error:%v", rvErr.Error())
		}
		return
	}

	modelVal, modelErr := s.modelProvider.GetEntityValue(rvModel.Interface(vType.IsPtrType()))
	if modelErr != nil {
		re = cd.NewError(cd.UnExpected, modelErr.Error())
		log.Errorf("querySingleRelation failed, s.modelProvider.GetEntityValue error:%v", modelErr.Error())
		return
	}

	vValue.Set(modelVal.Get())
	ret = vValue
	return
}

func (s *impl) querySliceRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, re *cd.Result) {
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		re = cd.NewError(cd.UnExpected, rErr.Error())
		log.Errorf("querySliceRelation failed, s.modelProvider.GetTypeModel error:%sv", rErr.Error())
		return
	}

	vType := vField.GetType()
	vValue, _ := vType.Interface(nil)
	valueList, valueErr := s.innerQueryRelationKeys(vModel, rModel, vField)
	if valueErr != nil {
		re = valueErr
		if valueErr.Fail() {
			log.Errorf("querySliceRelation failed, s.innerQueryRelationKeys error:%sv", valueErr.Error())
		} else if valueErr.Warn() {
			log.Warnf("querySliceRelation failed, s.innerQueryRelationKeys error:%sv", valueErr.Error())
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
		re = rModelErr
		if rModelErr.Fail() {
			log.Errorf("querySliceRelation failed, s.innerQueryRelationSliceModel error:%sv", rModelErr.Error())
		} else if rModelErr.Warn() {
			log.Warnf("querySliceRelation failed, s.innerQueryRelationSliceModel error:%sv", rModelErr.Error())
		}
		return
	}

	elemType := vType.Elem()
	for _, sv := range rModelList {
		modelVal, modelErr := s.modelProvider.GetEntityValue(sv.Interface(elemType.IsPtrType()))
		if modelErr != nil {
			re = cd.NewError(cd.UnExpected, modelErr.Error())
			log.Errorf("querySliceRelation failed, s.modelProvider.GetEntityValue error:%sv", modelErr.Error())
			return
		}

		vValue, rErr = s.modelProvider.AppendSliceValue(vValue, modelVal)
		if rErr != nil {
			re = cd.NewError(cd.UnExpected, rErr.Error())
			log.Errorf("querySliceRelation failed, s.modelProvider.AppendSliceValue error:%sv", rErr.Error())
			return
		}
	}
	ret = vValue
	return
}

func (s *impl) queryRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, re *cd.Result) {
	if deepLevel > maxDeepLevel {
		fieldType := vField.GetType()
		ret, _ = fieldType.Interface(nil)
		return
	}

	if vField.IsSlice() {
		ret, re = s.querySliceRelation(vModel, vField, deepLevel)
		if re != nil {
			if re.Fail() {
				log.Errorf("queryRelation failed, s.querySliceRelation error:%v", re.Error())
			} else if re.Warn() {
				log.Warnf("queryRelation failed, s.querySliceRelation error:%v", re.Error())
			}
		}
		return
	}

	ret, re = s.querySingleRelation(vModel, vField, deepLevel)
	if re != nil {
		if re.Fail() {
			log.Errorf("queryRelation failed, s.querySingleRelation error:%v", re.Error())
		} else if re.Warn() {
			log.Warnf("queryRelation failed, s.querySingleRelation error:%v", re.Error())
		}
	}
	return
}

func (s *impl) Query(vModel model.Model) (ret model.Model, re *cd.Result) {
	if vModel == nil {
		re = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	vFilter, vErr := s.getModelFilter(vModel)
	if vErr != nil {
		re = vErr
		if vErr.Fail() {
			log.Errorf("Query failed, s.getModelFilter error:%v", vErr.Error())
		} else if vErr.Warn() {
			log.Warnf("Query failed, s.getModelFilter error:%v", vErr.Error())
		}
		return
	}

	qModel, qErr := s.querySingle(vFilter, 0)
	if qErr != nil {
		re = qErr
		if qErr.Fail() {
			log.Errorf("Query failed, s.querySingle error:%v", re.Error())
		} else if qErr.Warn() {
			log.Warnf("Query failed, s.querySingle error:%v", re.Error())
		}
		return
	}

	ret = qModel
	return
}
