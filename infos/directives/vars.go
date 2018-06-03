package directives

import (
	"github.com/beevik/etree"
	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/infos"
)

// VarsInfo contains custom variables in a statement.
type VarsInfo struct {
	values map[string]string
}

type varsDirective struct{}

var (
	_ infos.TerminalDirective = (*varsDirective)(nil)
)

type varsLocalsKeyType struct{}

var (
	varsLocalsKey = varsLocalsKeyType{}
)

// ExtractVarsInfo extracts custom var information from a statement or nil if not exists.
func ExtractVarsInfo(stmt *infos.StmtInfo) *VarsInfo {
	locals := stmt.Locals(varsLocalsKey)
	if locals != nil {
		return locals.(*VarsInfo)
	}
	return nil
}

// Valid returns true if info != nil
func (info *VarsInfo) Valid() bool {
	return info != nil
}

// Has returns true if the named var exists. It returns false if info is nil or not exists.
func (info *VarsInfo) Has(name string) bool {
	if info == nil {
		return false
	}
	_, ok := info.values[name]
	return ok
}

// Value returns the named var's value. It returns "" if info is nil or not exists or the value is just "".
func (info *VarsInfo) Value(name string) string {
	if info == nil {
		return ""
	}
	return info.values[name]
}

func (d *varsDirective) Initialize(loader *datasrc.Loader, db *infos.DBInfo, stmt *infos.StmtInfo, tok etree.Token) error {

	// Get/set VarsInfo
	locals := stmt.Locals(varsLocalsKey)
	if locals == nil {
		locals = &VarsInfo{
			values: make(map[string]string),
		}
		stmt.SetLocals(varsLocalsKey, locals)
	}
	info := locals.(*VarsInfo)

	// Add vars names and values.
	elem := tok.(*etree.Element)
	for _, attr := range elem.Attr {
		info.values[attr.Key] = attr.Value
	}

	return nil
}

func (d *varsDirective) QueryFragment() (string, error) {
	return "", nil
}

func (d *varsDirective) TextFragment() (string, error) {
	return "", nil
}

func (d *varsDirective) ExtraProcess() error {
	return nil
}

func init() {
	infos.RegistDirectiveFactory(func() infos.Directive {
		return &varsDirective{}
	}, "vars")
}
