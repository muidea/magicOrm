package local

import (
	"reflect"
	"strings"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

const (
	constraintsTag = "constraint"
	ormTag         = "orm"
	viewTag        = "view"
)

type SpecImpl struct {
	fieldName  string
	primaryKey bool

	constraints  models.Constraints
	valueDeclare models.ValueDeclare
	viewDeclare  []models.ViewDeclare
}

var emptySpec = SpecImpl{primaryKey: false, valueDeclare: models.Customer}

// NewSpec name[key][auto]
func NewSpec(tag reflect.StructTag) (ret *SpecImpl, err *cd.Error) {
	ormSpec := tag.Get(ormTag)
	ret, err = getOrmSpec(ormSpec)
	if err != nil {
		return
	}
	constraints := tag.Get(constraintsTag)
	if constraints != "" {
		ret.constraints = utils.ParseConstraints(constraints)
	}

	viewSpec := tag.Get(viewTag)
	ret.viewDeclare = getViewItems(viewSpec)
	return
}

func getOrmSpec(spec string) (ret *SpecImpl, err *cd.Error) {
	items := strings.Split(strings.TrimSpace(spec), " ")
	ret = &SpecImpl{primaryKey: false, valueDeclare: models.Customer}
	ret.fieldName = items[0]
	for idx := 1; idx < len(items); idx++ {
		switch items[idx] {
		case models.AutoIncrement:
			ret.valueDeclare = models.AutoIncrement
		case models.UUID:
			ret.valueDeclare = models.UUID
		case models.Snowflake:
			ret.valueDeclare = models.Snowflake
		case models.DateTime:
			ret.valueDeclare = models.DateTime
		case models.KeyTag:
			ret.primaryKey = true
		}
	}

	return
}

func getViewItems(spec string) (ret []models.ViewDeclare) {
	ret = []models.ViewDeclare{}
	items := strings.Split(spec, ",")
	for _, sv := range items {
		switch sv {
		case models.DetailView:
			ret = append(ret, models.DetailView)
		case models.LiteView:
			ret = append(ret, models.LiteView)
		}
	}
	return
}

// GetFieldName Name
func (s *SpecImpl) GetFieldName() string {
	return s.fieldName
}

func (s *SpecImpl) IsPrimaryKey() bool {
	return s.primaryKey
}

func (s *SpecImpl) GetValueDeclare() models.ValueDeclare {
	return s.valueDeclare
}

func (s *SpecImpl) GetConstraints() models.Constraints {
	return s.constraints
}

func (s *SpecImpl) EnableView(viewSpec models.ViewDeclare) bool {
	if viewSpec == models.MetaView {
		return true
	}

	for _, val := range s.viewDeclare {
		if val == viewSpec {
			return true
		}
	}

	return false
}

func (s *SpecImpl) GetDefaultValue() any {
	return nil
}
