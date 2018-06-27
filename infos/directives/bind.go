package directives

import (
	"fmt"

	"github.com/beevik/etree"
	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/infos"
)

type bindDirective struct {
	name string
	val  string
}

var (
	_ infos.NonterminalDirective = (*bindDirective)(nil)
)

func (d *bindDirective) Initialize(loader *datasrc.Loader, db *infos.DBInfo, stmt *infos.StmtInfo, tok etree.Token) error {
	elem := tok.(*etree.Element)

	// Get bind name.
	name := elem.SelectAttrValue("name", "")
	if name == "" {
		return fmt.Errorf("Missing 'name' attribute in <bind> directive")
	}

	// Find the named arg.
	arg := (*ArgInfo)(nil)
	for _, a := range ExtractArgsInfo(stmt).Args() {
		if a.ArgName() == name {
			arg = a
			break
		}
	}
	if arg == nil {
		return fmt.Errorf("Can't find arg named %+q", name)
	}
	d.name = name

	// Default to NULL
	d.val = elem.Text()
	if d.val == "" {
		d.val = "NULL"
	}
	return nil
}

func (d *bindDirective) Expand() ([]etree.Token, error) {
	elem := etree.NewElement("repl")
	elem.CreateAttr("with", ":"+d.name)
	elem.SetText(d.val)
	return []etree.Token{elem}, nil
}

func init() {
	infos.RegistDirectiveFactory(func() infos.Directive {
		return &bindDirective{}
	}, "bind")
}
