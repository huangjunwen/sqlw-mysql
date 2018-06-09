package render

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/huangjunwen/sqlw-mysql/infos/directives"
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

func slice(s string, args ...int) (result string, err error) {
	defer func() {
		if e := recover(); e != nil {
			result = ""
			err = fmt.Errorf("Slice: %s", e)
		}
	}()

	switch len(args) {
	case 1:
		return s[args[0]:], nil
	case 2:
		return s[args[0]:args[1]], nil
	default:
		return "", fmt.Errorf("Slice: expect one or two integer but got %d", len(args))
	}
}

func (r *Renderer) funcMap() template.FuncMap {

	return template.FuncMap{

		"UpperCamel": func(s string) string {
			return camel(s, true)
		},

		"LowerCamel": func(s string) string {
			return camel(s, false)
		},

		"Slice": slice,

		"ScanType": func(s interface{}) (string, error) {
			return r.scanTypeMap.ScanType(s)
		},

		"NotNullScanType": func(s interface{}) (string, error) {
			return r.scanTypeMap.NotNullScanType(s)
		},

		"NullScanType": func(s interface{}) (string, error) {
			return r.scanTypeMap.NullScanType(s)
		},

		"ExtractVarsInfo": directives.ExtractVarsInfo,

		"ExtractArgsInfo": directives.ExtractArgsInfo,

		"ExtractWildcardsInfo": directives.ExtractWildcardsInfo,
	}

}
