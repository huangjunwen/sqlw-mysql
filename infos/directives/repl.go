package directives

import (
	"fmt"

	"github.com/beevik/etree"
	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/infos"
)

type replDirective struct {
	original string
	by       string
}

var (
	_ infos.TerminalDirective = (*replDirective)(nil)
)

func (d *replDirective) Initialize(loader *datasrc.Loader, db *infos.DBInfo, stmt *infos.StmtInfo, tok etree.Token) error {
	elem := tok.(*etree.Element)
	by := elem.SelectAttrValue("by", "")
	if by == "" {
		return fmt.Errorf("Missing 'by' attribute in <repl> directive")
	}
	d.original = elem.Text()
	d.by = by
	return nil
}

func (d *replDirective) QueryFragment() (string, error) {
	return d.original, nil
}

func (d *replDirective) TextFragment() (string, error) {
	return d.by, nil
}

func (d *replDirective) ExtraProcess() error {
	return nil
}

func init() {
	infos.RegistDirectiveFactory(func() infos.Directive {
		return &replDirective{}
	}, "repl", "r")
}
