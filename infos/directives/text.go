package directives

import (
	"github.com/beevik/etree"
	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/infos"
)

type textDirective string

var (
	_ infos.NonterminalDirective = (*textDirective)(nil)
)

func (d *textDirective) Initialize(loader *datasrc.Loader, db *infos.DBInfo, stmt *infos.StmtInfo, tok etree.Token) error {
	elem := tok.(*etree.Element)
	*d = textDirective(elem.Text())
	return nil
}

func (d *textDirective) Expand() ([]etree.Token, error) {
	elem := etree.NewElement("repl")
	elem.CreateAttr("with", string(*d))
	return []etree.Token{elem}, nil
}

func init() {
	infos.RegistDirectiveFactory(func() infos.Directive {
		d := textDirective("")
		return &d
	}, "text")
}
