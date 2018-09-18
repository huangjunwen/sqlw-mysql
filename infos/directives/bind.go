package directives

import (
	"fmt"

	"github.com/beevik/etree"

	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/infos"
)

type bindDirective struct {
	bindName string
	val      string
}

var (
	_ infos.NonterminalDirective = (*bindDirective)(nil)
)

func (d *bindDirective) Initialize(loader *datasrc.Loader, db *infos.DBInfo, stmt *infos.StmtInfo, tok etree.Token) error {
	argsInfo := ExtractArgsInfo(stmt)
	if argsInfo == nil {
		return fmt.Errorf("Can't use <bind> if you have not declare <arg>")
	}

	elem := tok.(*etree.Element)
	d.bindName = elem.SelectAttrValue("name", "")
	if d.bindName == "" {
		return fmt.Errorf("Missing 'name' attribute in <bind> directive")
	}

	found := false
	for _, argInfo := range argsInfo.Args() {
		if argInfo.ArgName() == d.bindName {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("No <arg> with name=%+q", d.bindName)
	}

	d.val = elem.Text()
	return nil
}

func (d *bindDirective) Expand() ([]etree.Token, error) {
	// <bind name="xxx" /> -> <repl by=":xxx">NULL</repl>
	// <bind name="xxx">val</bind> -> <repl by=":xxx">val</repl>
	elem := etree.NewElement("repl")
	elem.CreateAttr("by", ":"+d.bindName)
	val := "NULL"
	if d.val != "" {
		val = d.val
	}
	elem.SetText(val)
	return []etree.Token{elem}, nil
}

func init() {
	infos.RegistDirectiveFactory(func() infos.Directive {
		return &bindDirective{}
	}, "bind", "b")
}
