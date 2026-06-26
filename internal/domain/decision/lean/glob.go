package lean

import "regexp"

// globToRegexp converts a path glob into an anchored regular expression with
// doublestar (`**`) semantics, so an ADR's applies_to patterns can be matched
// against changed file paths. Patterns and paths use forward slashes.
//
//   - `/**/`        zero or more path segments  (`port/**/*.py` matches `port/x.py`)
//   - leading `**/` zero or more leading segments
//   - trailing `/**` everything below a directory
//   - bare `**`     any run of characters, including `/`
//   - `*`           any run within a single segment (does not cross `/`)
//   - `?`           a single non-`/` character
func globToRegexp(pattern string) string {
	var b []byte
	b = append(b, '^')
	for i := 0; i < len(pattern); {
		c := pattern[i]
		switch c {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				i += 2 // consume "**"
				if i < len(pattern) && pattern[i] == '/' {
					i++ // consume the trailing slash too
					b = append(b, "(?:.*/)?"...)
				} else {
					b = append(b, ".*"...)
				}
			} else {
				b = append(b, "[^/]*"...)
				i++
			}
		case '?':
			b = append(b, "[^/]"...)
			i++
		default:
			if isRegexMeta(c) {
				b = append(b, '\\')
			}
			b = append(b, c)
			i++
		}
	}
	b = append(b, '$')
	return string(b)
}

func isRegexMeta(c byte) bool {
	switch c {
	case '.', '+', '(', ')', '|', '[', ']', '{', '}', '^', '$', '\\':
		return true
	}
	return false
}

// MatchGlob reports whether path matches the doublestar glob pattern. An
// invalid pattern matches nothing.
func MatchGlob(pattern, path string) bool {
	re, err := regexp.Compile(globToRegexp(pattern))
	if err != nil {
		return false
	}
	return re.MatchString(path)
}
