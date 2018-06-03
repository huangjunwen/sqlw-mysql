package infos

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/beevik/etree"
	"github.com/huangjunwen/sqlw-mysql/datasrc"
)

// StmtInfo contains information of a statement.
type StmtInfo struct {
	stmtName        string
	directives      []TerminalDirective
	query           string        // Construct by QueryFragment
	stmtType        string        // 'SELECT'/'UPDATE'/'INSERT'/'DELETE'
	queryResultCols []datasrc.Col // For SELECT stmt only
	text            string        // Construct by TextFragment

	locals map[interface{}]interface{} // directive locals
}

// NewStmtInfo creates a new StmtInfo from an xml element, example statement xml element:
//
//   <stmt name="BlogByUser">
//     <arg name="userId" type="int" />
//     SELECT <wc table="blog" /> FROM blog WHERE user_id=<repl with=":userId">1</repl>
//   </stmt>
//
// which contains SQL statement fragments and special directives.
func NewStmtInfo(loader *datasrc.Loader, db *DBInfo, stmtElem *etree.Element) (*StmtInfo, error) {

	info := &StmtInfo{
		locals: map[interface{}]interface{}{},
	}

	if stmtElem.Tag != "stmt" {
		return nil, fmt.Errorf("Expect <stmt> but got <%s>", stmtElem.Tag)
	}

	// Name attribute.
	{
		info.stmtName = stmtElem.SelectAttrValue("name", "")
		if info.stmtName == "" {
			return nil, fmt.Errorf("Missing 'name' attribute of <%s>", info.stmtType)
		}
	}

	// Convert to directives and Initialize.
	{
		info.directives = []TerminalDirective{}
		for _, token := range stmtElem.Child {
			directives, err := info.token2TerminalDirectives(loader, db, token)
			if err != nil {
				return nil, err
			}
			info.directives = append(info.directives, directives...)
		}
	}

	// Construct query.
	{
		fragments := []string{}
		for _, directive := range info.directives {
			fragment, err := directive.QueryFragment()
			if err != nil {
				return nil, err
			}
			fragments = append(fragments, fragment)
		}
		info.query = strings.TrimSpace(strings.Join(fragments, ""))
	}

	// Determine statement type.
	// TODO: union ?
	{
		sp := strings.IndexFunc(info.query, unicode.IsSpace)
		if sp < 0 {
			return nil, fmt.Errorf("Can't determine statement type for %+q", info.query)
		}
		verb := strings.ToUpper(info.query[:sp])
		switch verb {
		case "SELECT", "INSERT", "UPDATE", "DELETE", "REPLACE":
		default:
			return nil, fmt.Errorf("Not supported statement type %+q", verb)
		}

		info.stmtType = verb
	}

	// Get query result columns if it is a SELECT.
	if info.StmtType() == "SELECT" {
		cols, err := loader.LoadCols(info.query)
		if err != nil {
			return nil, err
		}
		info.queryResultCols = cols
	}

	// Construct text.
	{
		fragments := []string{}
		for _, directive := range info.directives {
			fragment, err := directive.TextFragment()
			if err != nil {
				return nil, err
			}
			fragments = append(fragments, fragment)
		}
		info.text = strings.TrimSpace(strings.Join(fragments, ""))
	}

	// Extra process
	{
		for _, directive := range info.directives {
			if err := directive.ExtraProcess(); err != nil {
				return nil, err
			}
		}
	}

	return info, nil

}

func (info *StmtInfo) token2TerminalDirectives(loader *datasrc.Loader, db *DBInfo, token etree.Token) ([]TerminalDirective, error) {

	directive := Directive(nil)

	// Token -> Directive.
	switch tok := token.(type) {
	case *etree.CharData:
		directive = &textDirective{}

	case *etree.Element:
		factory := directiveFactories[tok.Tag]
		if factory == nil {
			return nil, fmt.Errorf("Unknown directive <%s>", tok.Tag)
		}
		directive = factory()

	default:
		return nil, nil
	}

	// Initialize
	if err := directive.Initialize(loader, db, info, token); err != nil {
		return nil, err
	}

	// Expand directive recursively if it is NonterminalDirective.
	switch d := directive.(type) {

	case TerminalDirective:
		return []TerminalDirective{d}, nil

	case NonterminalDirective:
		ds, err := d.Expand()
		if err != nil {
			return nil, err
		}

		ret := []TerminalDirective{}
		for _, d := range ds {
			terminalDirectives, err := info.token2TerminalDirectives(loader, db, d)
			if err != nil {
				return nil, err
			}
			ret = append(ret, terminalDirectives...)
		}

		return ret, nil

	default:
		panic(fmt.Errorf("Directive must be either TerminalDirective or NonterminalDirective"))
	}

}

// Valid returns true if info != nil.
func (info *StmtInfo) Valid() bool {
	return info != nil
}

// String is same as StmtName.
func (info *StmtInfo) String() string {
	return info.StmtName()
}

// StmtName returns the name of the StmtInfo. It returns "" if info is nil.
func (info *StmtInfo) StmtName() string {
	if info == nil {
		return ""
	}
	return info.stmtName
}

// StmtType returns the statement type, one of "SELECT"/"UPDATE"/"INSERT"/"UPDATE". It returns "" if info is nil.
func (info *StmtInfo) StmtType() string {
	if info == nil {
		return ""
	}
	return info.stmtType
}

// Directives returns the list of terminal directives the statement composed by.
func (info *StmtInfo) Directives() []TerminalDirective {
	if info == nil {
		return nil
	}
	return info.directives
}

// Text returns the statment text. It returns "" if info is nil.
func (info *StmtInfo) Text() string {
	if info == nil {
		return ""
	}
	return info.text
}

// Query returns the statement query. This is a valid SQL. It returns "" if info is nil.
func (info *StmtInfo) Query() string {
	if info == nil {
		return ""
	}
	return info.query
}

// QueryResultCols returns the result columns of the query if the statement is a SELECT.
func (info *StmtInfo) QueryResultCols() []datasrc.Col {
	if info == nil {
		return nil
	}
	return info.queryResultCols
}

// Locals returns the associated value for the given key in StmtInfo's locals map.
// This map is used by directives to store directive specific variables.
func (info *StmtInfo) Locals(key interface{}) interface{} {
	return info.locals[key]
}

// SetLocals set key/value into StmtInfo's locals map. See document in Locals.
func (info *StmtInfo) SetLocals(key, val interface{}) {
	info.locals[key] = val
}
