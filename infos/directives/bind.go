package directives

import (
	"fmt"

	"github.com/beevik/etree"

	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/infos"
)

type bindDirective struct {
	bindName string
	v        string
}

var (
	_ infos.NonterminalDirective = (*bindDirective)(nil)
)

func (d *bindDirective) Initialize(loader *datasrc.Loader, db *infos.DBInfo, stmt *infos.StmtInfo, tok etree.Token) error {
	elem := tok.(*etree.Element)
	d.bindName = elem.SelectAttrValue("name", "")
	if d.bindName == "" {
		return fmt.Errorf("Missing 'name' attribute in <bind> directive")
	}
	d.v = elem.Text()
	return nil
}

func (d *bindDirective) Expand() ([]etree.Token, error) {
	// <bind name="xxx" /> -> <repl by=":xxx">NULL</repl>
	// <bind name="xxx">v</bind> -> <repl by=":xxx">v</repl>
	elem := etree.NewElement("repl")
	elem.CreateAttr("by", ":"+d.bindName)
	v := "NULL"
	if d.v != "" {
		v = d.v
	}
	elem.SetText(v)
	return []etree.Token{elem}, nil
}

func init() {
	infos.RegistDirectiveFactory(func() infos.Directive {
		return &bindDirective{}
	}, "bind", "b")
}
