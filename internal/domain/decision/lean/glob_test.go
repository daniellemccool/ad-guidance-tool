package lean

import "testing"

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		// port/**/*.py — ** matches zero or more segments
		{"port/**/*.py", "port/script.py", true},
		{"port/**/*.py", "port/helpers/flow_builder.py", true},
		{"port/**/*.py", "lib/script.py", false},
		{"port/**/*.py", "port/data.json", false},
		// **/name.py — any depth
		{"**/flow_builder.py", "flow_builder.py", true},
		{"**/flow_builder.py", "port/helpers/flow_builder.py", true},
		{"**/flow_builder.py", "port/helpers/flow_builder.pyc", false},
		// single * does not cross a separator
		{"*.py", "script.py", true},
		{"*.py", "port/script.py", false},
		// nested ** on both sides
		{"lib/**/projections/**", "lib/a/projections/b.ex", true},
		{"lib/**/projections/**", "lib/projections/b.ex", true},
		{"lib/**/projections/**", "lib/a/read_models/b.ex", false},
		// bare ** matches anything
		{"**", "any/where/at/all.py", true},
		// ? matches exactly one non-separator char
		{"v?.py", "v1.py", true},
		{"v?.py", "v10.py", false},
	}
	for _, c := range cases {
		if got := MatchGlob(c.pattern, c.path); got != c.want {
			t.Errorf("MatchGlob(%q, %q) = %v, want %v (regex %q)", c.pattern, c.path, got, c.want, globToRegexp(c.pattern))
		}
	}
}
