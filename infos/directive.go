package infos

import (
	"github.com/beevik/etree"
	"github.com/huangjunwen/sqlw-mysql/datasrc"
)

// Directive represents a fragment of a statement.
type Directive interface {
	// Initialize the directive.
	Initialize(loader *datasrc.Loader, db *DBInfo, stmt *StmtInfo, tok etree.Token) error
}

// NonterminalDirective can expand to other directives.
type NonterminalDirective interface {
	Directive

	// Expand to a list of xml tokens which will be converted to directives later.
	Expand() ([]etree.Token, error)
}

// TerminalDirective can not expand to other directives.
type TerminalDirective interface {
	Directive

	// QueryFragment returns the fragment of this directive to construct a valid SQL query.
	// The SQL query is used to determine statement type, to obtain result column information for SELECT query,
	// and optionally to check SQL correctness.
	QueryFragment() (string, error)

	// TextFragment returns the final fragment of this directive to construct a final statement text.
	// The statement text is no need to be a valid SQL query. It is up to the template to determine how to use it.
	TextFragment() (string, error)

	// ExtraProcess runs some extra process.
	ExtraProcess() error
}

// textDirective is a special directive.
type textDirective struct {
	data string
}

var (
	_ TerminalDirective = (*textDirective)(nil)
)

func (d *textDirective) Initialize(loader *datasrc.Loader, db *DBInfo, stmt *StmtInfo, tok etree.Token) error {
	d.data = tok.(*etree.CharData).Data
	return nil
}

func (d *textDirective) QueryFragment() (string, error) {
	return d.data, nil
}

func (d *textDirective) TextFragment() (string, error) {
	return d.data, nil
}

func (d *textDirective) ExtraProcess() error {
	return nil
}

var (
	// Map tag name -> factory
	directiveFactories = map[string]func() Directive{}
)

// RegistDirectiveFactory regist directive factories.
func RegistDirectiveFactory(factory func() Directive, tags ...string) {
	for _, tag := range tags {
		directiveFactories[tag] = factory
	}
}
