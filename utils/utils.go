package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// CamelName represents camel case of a name.
type CamelName struct {
	// Upper camel case name.
	UName string
	// Lower camel case name.
	LName string
}

// NewCamelName creates a CamelName.
func NewCamelName(s string) CamelName {
	uname := UpperCamel(s)
	if uname == "" {
		panic(fmt.Errorf("UpperCamel(%+q) returns empty string", s))
	}
	lname := LowerCamel(s)
	if lname == "" {
		panic(fmt.Errorf("LowerCamel(%+q) returns empty string", s))
	}
	return CamelName{UName: uname, LName: lname}
}

// UpperCamel convert `s` to upper camel case.
func UpperCamel(s string) string {
	return camel(s, true)
}

// LowerCamel convert `s` to lower camel case.
func LowerCamel(s string) string {
	return camel(s, false)
}

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
