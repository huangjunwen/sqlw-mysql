package render

import (
	"text/template"

	"github.com/Masterminds/sprig"

	"github.com/huangjunwen/sqlw-mysql/infos/directives"
)

func (r *Renderer) funcMap() template.FuncMap {

	// Add sprig's functions.
	fm := sprig.TxtFuncMap()

	// Add more.
	fm["scanType"] = func(col interface{}) (string, error) {
		return r.manifest.ScanTypeMap.ScanType(col)
	}
	fm["notNullScanType"] = func(col interface{}) (string, error) {
		return r.manifest.ScanTypeMap.NotNullScanType(col)
	}
	fm["nullScanType"] = func(col interface{}) (string, error) {
		return r.manifest.ScanTypeMap.NullScanType(col)
	}
	fm["extractVarsInfo"] = directives.ExtractVarsInfo
	fm["extractArgsInfo"] = directives.ExtractArgsInfo
	fm["extractWildcardsInfo"] = directives.ExtractWildcardsInfo

	return fm
}
