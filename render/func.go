package render

import (
	"fmt"
	"strings"
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

// Three forms to use enum:
// Enum N: iterate [0, N) in step 1
// Enum N, M: iterate [N, M) in step 1
// Enum N, M, S: iterate [N, M) in step S where S can be postive/negative but can't be 0
// https://stackoverflow.com/questions/22713500/iterating-a-range-of-integers-in-go-templates
func enum(args ...int) (chan int, error) {
	var start, end, step int
	switch len(args) {
	case 1:
		start = 0
		end = args[0]
		step = 1
	case 2:
		start = args[0]
		end = args[1]
		step = 1
	case 3:
		start = args[0]
		end = args[1]
		step = args[2]
	default:
		return nil, fmt.Errorf("Enum: expect 1 to 3 args but got %d args", len(args))
	}

	if step > 0 {
		if end < start {
			return nil, fmt.Errorf("Enum: step(%d) > 0 but end(%d) < start(%d)", step, end, start)
		}
	} else if step < 0 {
		if end > start {
			return nil, fmt.Errorf("Enum: step(%d) < 0 but end(%d) > start(%d)", step, end, start)
		}
	} else {
		return nil, fmt.Errorf("Enum: step can't be 0")
	}

	ret := make(chan int)
	go func() {
		if step > 0 {
			i := start
			for i < end {
				ret <- i
				i += step
			}
		} else {
			i := start
			for i > end {
				ret <- i
				i += step
			}
		}
		close(ret)
	}()
	return ret, nil
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

func replace(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

func (r *Renderer) funcMap() template.FuncMap {

	return template.FuncMap{
		"Errorf": errorf,

		"Replace": replace,

		"Slice": slice,

		"Enum": enum,

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
