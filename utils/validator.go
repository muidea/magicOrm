package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"

	"github.com/muidea/magicOrm/models"
)

type valueValidator struct {
	registry map[models.Key]models.ValidatorFunc
}

func NewValueValidator() models.ValueValidator {
	sv := &valueValidator{registry: make(map[models.Key]models.ValidatorFunc)}
	sv.loadBuiltins()
	return sv
}

func (sv *valueValidator) Register(k models.Key, fn models.ValidatorFunc) { sv.registry[k] = fn }

func (sv *valueValidator) ValidateValue(val any, directives []models.Directive) error {
	for _, d := range directives {
		k := models.Key(d.Key())
		if fn, ok := sv.registry[k]; ok {
			if err := fn(val, d.Args()); err != nil {
				return &models.ConstraintViolation{Key: k, Cause: err}
			}
		}
	}
	return nil
}

func (sv *valueValidator) loadBuiltins() {
	// req
	sv.Register(models.KeyRequired, func(v any, _ []string) error {
		if IsReallyZeroValue(v) {
			return fmt.Errorf("required")
		}
		return nil
	})
	// min & max
	compare := func(v any, arg string, isMin bool) error {
		limit, parseErr := strconv.ParseFloat(arg, 64)
		if parseErr != nil {
			return fmt.Errorf("invalid numeric constraint")
		}
		rv := indirectConstraintValue(v)
		var val float64
		switch rv.Kind() {
		case reflect.String, reflect.Slice, reflect.Map:
			val = float64(rv.Len())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val = float64(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val = float64(rv.Uint())
		case reflect.Float32, reflect.Float64:
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
	sv.Register(models.KeyMin, func(v any, a []string) error {
		if len(a) != 1 {
			return fmt.Errorf("min requires one argument")
		}
		return compare(v, a[0], true)
	})
	sv.Register(models.KeyMax, func(v any, a []string) error {
		if len(a) != 1 {
			return fmt.Errorf("max requires one argument")
		}
		return compare(v, a[0], false)
	})

	// range=min:max
	sv.Register(models.KeyRange, func(v any, a []string) error {
		if len(a) != 2 {
			return fmt.Errorf("range requires two arguments")
		}
		min, e1 := strconv.ParseFloat(a[0], 64)
		max, e2 := strconv.ParseFloat(a[1], 64)
		if e1 != nil || e2 != nil || min > max {
			return fmt.Errorf("invalid range")
		}
		rv := indirectConstraintValue(v)
		var f float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f = float64(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			f = float64(rv.Uint())
		case reflect.Float32, reflect.Float64:
			f = rv.Float()
		default:
			return nil
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
		if len(a) != 1 {
			return fmt.Errorf("re requires one argument")
		}
		m, regexErr := regexp.MatchString(a[0], fmt.Sprintf("%v", v))
		if regexErr != nil {
			return regexErr
		}
		if !m {
			return fmt.Errorf("invalid format")
		}
		return nil
	})
}

func indirectConstraintValue(v any) reflect.Value {
	rv := reflect.ValueOf(v)
	for rv.IsValid() && (rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface) {
		if rv.IsNil() {
			return reflect.Value{}
		}
		rv = rv.Elem()
	}
	return rv
}
