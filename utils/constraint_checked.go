package utils

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/muidea/magicOrm/models"
)

// ConstraintTerms splits directives without splitting commas in regex groups.
func ConstraintTerms(value string) ([]string, error) {
	var terms []string
	start := 0
	var stack []rune
	escaped := false
	quoted := false
	for i, c := range value {
		if escaped {
			escaped = false
			continue
		}
		if !strings.HasPrefix(strings.TrimSpace(value[start:]), "re=") {
			if c == ',' {
				terms = append(terms, strings.TrimSpace(value[start:i]))
				start = i + 1
			}
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}

		if c == '"' && strings.HasPrefix(strings.TrimSpace(value[start:]), `re="`) {
			quoted = !quoted
			continue
		}
		if quoted {
			continue
		}
		if len(stack) > 0 && stack[len(stack)-1] == ']' {
			if c == ']' {
				stack = stack[:len(stack)-1]
			}
			continue
		}
		switch c {
		case '[':
			stack = append(stack, ']')
		case '{':
			stack = append(stack, '}')
		case '(':
			stack = append(stack, ')')
		case ']', '}', ')':
			if len(stack) == 0 || stack[len(stack)-1] != c {
				return nil, fmt.Errorf("unbalanced constraint expression")
			}
			stack = stack[:len(stack)-1]
		case ',':
			if len(stack) == 0 {
				// A final regexp directive owns literal commas too. Split a
				// subsequent directive only when its name is recognized.
				if strings.HasPrefix(strings.TrimSpace(value[start:i]), "re=") {
					rest := strings.TrimSpace(value[i+1:])
					head, _, _ := strings.Cut(rest, ",")
					next, _, _ := strings.Cut(head, "=")
					next = strings.TrimSpace(next)
					switch next {
					case "req", "ro", "wo", "min", "max", "range", "in", "re":
					default:
						continue
					}
				}
				terms = append(terms, strings.TrimSpace(value[start:i]))
				start = i + 1
			}
		}
	}
	if len(stack) != 0 || escaped || quoted {
		return nil, fmt.Errorf("unbalanced constraint expression")
	}
	if tail := strings.TrimSpace(value[start:]); tail != "" {
		terms = append(terms, tail)
	}
	return terms, nil
}

// ParseConstraintsChecked validates the built-in portable constraint language.
// ParseConstraints remains available to consumers registering custom directives.
func ParseConstraintsChecked(value string) (models.Constraints, error) {
	terms, err := ConstraintTerms(value)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, term := range terms {
		key, arg, hasArgs := strings.Cut(term, "=")
		key = strings.TrimSpace(key)
		if seen[key] {
			return nil, fmt.Errorf("duplicate constraint %s", key)
		}
		seen[key] = true
		switch models.Key(key) {
		case models.KeyRequired, models.KeyReadOnly, models.KeyWriteOnly:
			if hasArgs {
				return nil, fmt.Errorf("%s takes no arguments", key)
			}
		case models.KeyMin, models.KeyMax, models.KeyRange:
			count := 1
			if models.Key(key) == models.KeyRange {
				count = 2
			}
			args := strings.Split(arg, ":")
			if !hasArgs || len(args) != count {
				return nil, fmt.Errorf("invalid %s arguments", key)
			}
			nums := make([]float64, count)
			for i, a := range args {
				n, e := strconv.ParseFloat(strings.TrimSpace(a), 64)
				if e != nil || math.IsNaN(n) || math.IsInf(n, 0) {
					return nil, fmt.Errorf("invalid %s number", key)
				}
				nums[i] = n
			}
			if count == 2 && nums[0] > nums[1] {
				return nil, fmt.Errorf("range minimum exceeds maximum")
			}
		case models.KeyIn:
			if !hasArgs || arg == "" {
				return nil, fmt.Errorf("enumeration requires values")
			}
			values := map[string]bool{}
			for _, a := range strings.Split(arg, ":") {
				a = strings.TrimSpace(a)
				if a == "" || values[a] {
					return nil, fmt.Errorf("empty or duplicate enumeration value")
				}
				values[a] = true
			}
		case models.KeyRegexp:
			if strings.HasPrefix(arg, `"`) {
				var e error
				arg, e = strconv.Unquote(arg)
				if e != nil {
					return nil, fmt.Errorf("invalid quoted regular expression: %w", e)
				}
			}
			if !hasArgs || arg == "" {
				return nil, fmt.Errorf("regular expression is required")
			}
			if _, e := regexp.Compile(arg); e != nil {
				return nil, fmt.Errorf("invalid regular expression: %w", e)
			}
		default:
			return nil, fmt.Errorf("unknown constraint %s", key)
		}
	}
	if seen["ro"] && seen["wo"] {
		return nil, fmt.Errorf("read-only and write-only conflict")
	}
	if seen["range"] && (seen["min"] || seen["max"]) {
		return nil, fmt.Errorf("range conflicts with min/max")
	}
	// Parse each directive separately: the permissive parser's regex protection
	// cannot recognize every comma inside a valid expression.
	result := constraintsImpl{}
	for _, term := range terms {
		key, arg, hasArgs := strings.Cut(term, "=")
		key = strings.TrimSpace(key)
		d := directiveImpl{key: models.Key(key), hasArgs: hasArgs}
		if hasArgs {
			if d.key == models.KeyRegexp {
				if strings.HasPrefix(arg, `"`) {
					arg, _ = strconv.Unquote(arg)
				}
				d.args = []string{arg}
			} else {
				for _, a := range strings.Split(arg, ":") {
					d.args = append(d.args, strings.TrimSpace(a))
				}
			}
		}
		result[d.key] = d
	}
	if lo, ok := result[models.KeyMin]; ok {
		if hi, ok := result[models.KeyMax]; ok {
			l, _ := strconv.ParseFloat(lo.args[0], 64)
			h, _ := strconv.ParseFloat(hi.args[0], 64)
			if l > h {
				return nil, fmt.Errorf("minimum exceeds maximum")
			}
		}
	}
	return &result, nil
}
