package orm

import (
	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) updateSingle(vModel model.Model) (err error) {
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

func (s *impl) updateRelation(vModel model.Model, vField model.Field) (err error) {
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

func (s *impl) Update(vModel model.Model) (ret model.Model, re *cd.Result) {
	if vModel == nil {
		re = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err := s.executor.BeginTransaction()
	if err != nil {
		re = cd.NewError(cd.UnExpected, err.Error())
		return
	}
	defer s.finalTransaction(err)

	err = s.updateSingle(vModel)
	if err != nil {
		re = cd.NewError(cd.UnExpected, err.Error())
		log.Errorf("Update failed, s.updateSingle error:%s", re.Error())
		return
	}

	for _, field := range vModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		err = s.updateRelation(vModel, field)
		if err != nil {
			re = cd.NewError(cd.UnExpected, err.Error())
			log.Errorf("Update failed, s.updateRelation error:%s", re.Error())
			return
		}
	}

	ret = vModel
	return
}
