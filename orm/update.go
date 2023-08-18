package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) updateSingle(vModel model.Model) (err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builder.BuildUpdate()
	if sqlErr != nil {
		err = sqlErr
		return err
	}

	_, _, err = s.executor.Execute(sqlStr)

	return err
}

func (s *impl) updateRelation(vModel model.Model, vField model.Field) (err error) {
	fType := vField.GetType()
	if fType.IsBasic() {
		return
	}

	err = s.deleteRelation(vModel, vField, 0)
	if err != nil {
		return
	}

	err = s.insertRelation(vModel, vField)
	if err != nil {
		return
	}

	return
}

func (s *impl) Update(vModel model.Model) (ret model.Model, err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	for {
		err = s.updateSingle(vModel)
		if err != nil {
			break
		}

		for _, field := range vModel.GetFields() {
			err = s.updateRelation(vModel, field)
			if err != nil {
				break
			}
		}

		break
	}

	if err != nil {
		return
	}

	ret = vModel
	return
}
