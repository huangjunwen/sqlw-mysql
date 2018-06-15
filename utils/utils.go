package utils

import (
	"regexp"
	"strings"
)

var (
	identifier = regexp.MustCompile(`([A-Za-z])([A-Za-z0-9]*)`)
)

func camel(s string, upper bool) string {
	parts := []string{}
	for _, id := range identifier.FindAllStringSubmatch(s, -1) {
		first, remain := id[1], id[2]
		if len(parts) == 0 && !upper {
			first = strings.ToLower(first)
		} else {
			first = strings.ToUpper(first)
		}
		parts = append(parts, first, remain)
	}
	return strings.Join(parts, "")
}

// UpperCamel convert `s` to upper camel case.
func UpperCamel(s string) string {
	return camel(s, true)
}

// LowerCamel convert `s` to lower camel case.
func LowerCamel(s string) string {
	return camel(s, false)
}
