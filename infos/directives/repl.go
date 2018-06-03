package directives

import (
	"fmt"

	"github.com/beevik/etree"
	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/infos"
)

type replDirective struct {
	origin string
	with   string
}

var (
	_ infos.TerminalDirective = (*replDirective)(nil)
)

func (d *replDirective) Initialize(loader *datasrc.Loader, db *infos.DBInfo, stmt *infos.StmtInfo, tok etree.Token) error {
	elem := tok.(*etree.Element)
	with := elem.SelectAttrValue("with", "")
	if with == "" {
		return fmt.Errorf("Missing 'with' attribute in <repl> directive")
	}
	d.origin = elem.Text()
	d.with = with
	return nil
}

func (d *replDirective) QueryFragment() (string, error) {
	return d.origin, nil
}

func (d *replDirective) ProcessQueryResultCols(resultCols *[]datasrc.Col) error {
	return nil
}

func (d *replDirective) Fragment() (string, error) {
	return d.with, nil
}

func init() {
	infos.RegistDirectiveFactory(func() infos.Directive {
		return &replDirective{}
	}, "repl")
}
