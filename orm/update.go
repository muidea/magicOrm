package orm

import (
	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) updateSingle(hBuilder builder.Builder) (err *cd.Result) {
	sqlStr, sqlErr := hBuilder.BuildUpdate()
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

func (s *impl) updateRelation(hBuilder builder.Builder, vField model.Field) (err *cd.Result) {
	err = s.deleteRelation(hBuilder, vField, 0)
	if err != nil {
		log.Errorf("updateRelation failed, s.deleteRelation error:%s", err.Error())
		return
	}

	err = s.insertRelation(hBuilder, vField)
	if err != nil {
		log.Errorf("updateRelation failed, s.insertRelation error:%s", err.Error())
	}
	return
}

func (s *impl) updateModel(vModel model.Model) (ret model.Model, err *cd.Result) {
	hBuilder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	err = s.updateSingle(hBuilder)
	if err != nil {
		log.Errorf("Update failed, s.updateSingle error:%s", err.Error())
		return
	}

	for _, field := range vModel.GetFields() {
		if field.IsBasic() || !field.GetValue().IsValid() {
			continue
		}

		err = s.updateRelation(hBuilder, field)
		if err != nil {
			log.Errorf("Update failed, s.updateRelation error:%s", err.Error())
			return
		}
	}

	ret = vModel
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

	ret, err = s.updateModel(vModel)
	if err != nil {
		log.Errorf("Update failed, s.updateSingle error:%s", err.Error())
		return
	}
	return
}
