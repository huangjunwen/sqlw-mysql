package render

import (
	"fmt"
	"text/template"

	"github.com/huangjunwen/sqlw-mysql/infos/directives"
	"github.com/huangjunwen/sqlw-mysql/utils"
)

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

func ternary(b bool, t interface{}, f interface{}) interface{} {
	if b {
		return t
	}
	return f
}

func errorf(format string, a ...interface{}) (string, error) {
	return "", fmt.Errorf(format, a...)
}

func (r *Renderer) funcMap() template.FuncMap {

	return template.FuncMap{
		"Errorf": errorf,

		"Slice": slice,

		"Ternary": ternary,

		"UpperCamel": utils.UpperCamel,

		"LowerCamel": utils.LowerCamel,

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
