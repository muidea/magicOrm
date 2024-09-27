package orm

import (
	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) updateSingle(vModel model.Model) (err *cd.Result) {
	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builderVal.BuildUpdate()
	if sqlErr != nil {
		err = sqlErr
		log.Errorf("updateSingle failed, builderVal.BuildUpdate error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(sqlStr)
	if err != nil {
		log.Errorf("updateSingle failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) updateRelation(vModel model.Model, vField model.Field) (err *cd.Result) {
	err = s.deleteRelation(vModel, vField, 0)
	if err != nil {
		log.Errorf("updateRelation failed, s.deleteRelation error:%s", err.Error())
		return
	}

	err = s.insertRelation(vModel, vField)
	if err != nil {
		log.Errorf("updateRelation failed, s.insertRelation error:%s", err.Error())
	}
	return
}

func (s *impl) Update(vModel model.Model) (ret model.Model, err *cd.Result) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	err = s.updateSingle(vModel)
	if err != nil {
		log.Errorf("Update failed, s.updateSingle error:%s", err.Error())
		return
	}

	for _, field := range vModel.GetFields() {
		if field.IsBasic() || !field.GetValue().IsValid() {
			continue
		}

		err = s.updateRelation(vModel, field)
		if err != nil {
			log.Errorf("Update failed, s.updateRelation error:%s", err.Error())
			return
		}
	}

	ret = vModel
	return
}
