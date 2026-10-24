package clirevision

import "testing"

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  Selector
	}{
		{"current", Selector{Current: true}},
		{"4", Selector{Ordinal: 4}},
		{"004", Selector{Ordinal: 4}},
		{"9223372036854775807", Selector{Ordinal: 9223372036854775807}},
		{"=123", Selector{Token: "123"}},
		{"=current", Selector{Token: "current"}},
		{"==opaque@value", Selector{Token: "=opaque@value"}},
		{"=値", Selector{Token: "値"}},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := Parse(tc.input, 100)
			if err != nil || got != tc.want {
				t.Fatalf("Parse = %#v, %v; want %#v", got, err, tc.want)
			}
		})
	}
}

func TestParseRefusesInvalid(t *testing.T) {
	for _, input := range []string{"", "=", "0", "000", "-1", "+1", "1.0", " 1", "١", "9223372036854775808", "abcdef", "=12345", string([]byte{0xff})} {
		if _, err := Parse(input, 4); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
}
