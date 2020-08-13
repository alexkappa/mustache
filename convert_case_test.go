package mustache

import "testing"

func Test_toSnakeCase(t *testing.T) {
	tests := []struct {
		want  string
		input string
	}{
		{"", ""},
		{"A", "a"},
		{"AaAa", "aa_aa"},
		{"BatteryLifeValue", "battery_life_value"},
		{"Id0Value", "id0_value"},
	}
	for _, test := range tests {
		have := toCamelCase(test.input)
		if have != test.want {
			t.Errorf("input=%q:\nhave: %q\nwant: %q", test.input, have, test.want)
		}
	}
}
