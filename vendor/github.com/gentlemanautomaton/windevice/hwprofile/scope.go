package hwprofile

import "strings"

// Scope hold a set of hardware profile scope flags.
type Scope uint32

// Match returns true if v contains all of the flags specified by c.
func (v Scope) Match(c Scope) bool {
	return v&c == c
}

// String returns a string representation of the flags using a default
// separator and format.
func (v Scope) String() string {
	return v.Join("|", FormatGo)
}

// Join returns a string representation of the flags using the given
// separator and format.
func (v Scope) Join(sep string, format Format) string {
	if s, ok := format[v]; ok {
		return s
	}

	var matched []string
	for i := 0; i < 32; i++ {
		flag := Scope(1 << uint32(i))
		if v.Match(flag) {
			if s, ok := format[flag]; ok {
				matched = append(matched, s)
			}
		}
	}

	return strings.Join(matched, sep)
}
