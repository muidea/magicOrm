package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"

	"github.com/muidea/magicOrm/models"
)

type ValidatorFunc func(val any, args []string) error

type ValueValidator struct {
	registry map[models.Key]ValidatorFunc
}

func NewValueValidator() *ValueValidator {
	sv := &ValueValidator{registry: make(map[models.Key]ValidatorFunc)}
	sv.loadBuiltins()
	return sv
}

func (sv *ValueValidator) Register(k models.Key, fn ValidatorFunc) { sv.registry[k] = fn }

func (sv *ValueValidator) ValidateValue(val any, directives []models.Directive) error {
	for _, d := range directives {
		k := models.Key(d.Key())
		if fn, ok := sv.registry[k]; ok {
			if err := fn(val, d.Args()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (sv *ValueValidator) loadBuiltins() {
	// req
	sv.Register(models.KeyRequired, func(v any, _ []string) error {
		if IsReallyZeroValue(v) {
			return fmt.Errorf("required")
		}
		return nil
	})
	// min & max
	compare := func(v any, arg string, isMin bool) error {
		limit, _ := strconv.ParseFloat(arg, 64)
		rv := reflect.ValueOf(v)
		var val float64
		switch rv.Kind() {
		case reflect.String, reflect.Slice, reflect.Map:
			val = float64(rv.Len())
		case reflect.Int, reflect.Int64:
			val = float64(rv.Int())
		case reflect.Float64:
			val = rv.Float()
		default:
			return nil
		}
		if isMin && val < limit {
			return fmt.Errorf("too small/short")
		}
		if !isMin && val > limit {
			return fmt.Errorf("too large/long")
		}
		return nil
	}
	sv.Register(models.KeyMin, func(v any, a []string) error { return compare(v, a[0], true) })
	sv.Register(models.KeyMax, func(v any, a []string) error { return compare(v, a[0], false) })

	// range=min:max
	sv.Register(models.KeyRange, func(v any, a []string) error {
		min, _ := strconv.ParseFloat(a[0], 64)
		max, _ := strconv.ParseFloat(a[1], 64)
		rv := reflect.ValueOf(v)
		if rv.Kind() < reflect.Int || rv.Kind() > reflect.Float64 {
			return nil
		}
		f := 0.0
		if rv.Kind() >= reflect.Float32 {
			f = rv.Float()
		} else {
			f = float64(rv.Int())
		}
		if f < min || f > max {
			return fmt.Errorf("out of range [%.1f, %.1f]", min, max)
		}
		return nil
	})

	// in=A:B
	sv.Register(models.KeyIn, func(v any, a []string) error {
		s := fmt.Sprintf("%v", v)
		if slices.Contains(a, s) {
			return nil
		}
		return fmt.Errorf("must be one of %v", a)
	})

	// re=pattern
	sv.Register(models.KeyRegexp, func(v any, a []string) error {
		m, _ := regexp.MatchString(a[0], fmt.Sprintf("%v", v))
		if !m {
			return fmt.Errorf("invalid format")
		}
		return nil
	})
}
