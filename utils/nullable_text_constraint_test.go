package utils

import "testing"

func TestNullableTextConstraintsValidateValuesAndKeepRequired(t *testing.T) {
	valid, invalid, empty := "2026-10-14", "bad", ""
	var missing *string
	for _, constraint := range []string{"re=^[0-9]{4}-[0-9]{2}-[0-9]{2}$", "in=2026-10-14:2026-10-15"} {
		for _, sample := range []struct {
			name  string
			value any
			ok    bool
		}{
			{"absent", nil, true}, {"typed-null", missing, true}, {"string", valid, true}, {"pointer", &valid, true}, {"bad-pointer", &invalid, false}, {"bad-string", invalid, false}, {"empty", &empty, false},
		} {
			t.Run(constraint+"/"+sample.name, func(t *testing.T) {
				for _, required := range []bool{false, true} {
					text := constraint
					if required {
						text = "req," + text
					}
					directives, err := ParseConstraintsChecked(text)
					if err != nil {
						t.Fatal(err)
					}
					want := sample.ok
					if required && (sample.value == nil || sample.name == "typed-null") {
						want = false
					}
					if err := NewValueValidator().ValidateValue(sample.value, directives.Directives()); (err == nil) != want {
						t.Fatal(text, sample.name, err)
					}
				}
			})
		}
	}
}
