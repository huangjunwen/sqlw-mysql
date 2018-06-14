package directives

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/infos"
)

// WildcardsInfo contains wildcard expansions information in a SELECT statement.
type WildcardsInfo struct {
	wildcards           []*WildcardInfo
	resultCols2Wildcard []int // Index of wildcards if the column is from a wildcard expansion or -1 otherwise.
}

// WildcardInfo contains a single wildcard expansion information in a SELECT statement.
type WildcardInfo struct {
	table  *infos.TableInfo
	alias  string
	offset int // Offset in result columns.
}

type wcDirective struct {
	// Context.
	loader *datasrc.Loader
	db     *infos.DBInfo
	stmt   *infos.StmtInfo

	// Table and optinal alias.
	table *infos.TableInfo
	alias string
}

var (
	_ infos.TerminalDirective = (*wcDirective)(nil)
)

type wcLocalsKeyType struct{}

var (
	wcLocalsKey = wcLocalsKeyType{}
)

// ExtractWildcardsInfo extracts wildcards information from a statement or nil if not exists.
func ExtractWildcardsInfo(stmt *infos.StmtInfo) *WildcardsInfo {
	locals := stmt.Locals(wcLocalsKey)
	if locals != nil {
		return locals.(*WildcardsInfo)
	}
	return nil
}

func newWildcardsInfo(loader *datasrc.Loader, db *infos.DBInfo, stmt *infos.StmtInfo) (*WildcardsInfo, error) {

	if stmt.StmtType() != "SELECT" {
		return nil, fmt.Errorf("<wc> is for SELECT only")
	}

	// Generate a random marker.
	var marker string
	{
		buf := make([]byte, 8)
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		marker = "wc" + hex.EncodeToString(buf)
	}

	fmtMarker := func(n int, start bool) string {
		if start {
			return fmt.Sprintf("%s_%d_s", marker, n)
		}
		return fmt.Sprintf("%s_%d_e", marker, n)
	}

	parseMarker := func(s string) (ok bool, n int, start bool) {
		parts := strings.Split(s, "_")
		if len(parts) != 3 || parts[0] != marker {
			return false, 0, false
		}

		i, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(fmt.Errorf("Invalid marker %+q", s))
		}

		switch parts[2] {
		case "s":
			return true, i, true
		case "e":
			return true, i, false
		default:
			panic(fmt.Errorf("Invalid marker %+q", s))
		}
	}

	// Construct a modified version query.
	directives := []*wcDirective{}
	query := ""
	{
		fragments := []string{}
		for _, directive := range stmt.Directives() {

			fragment, err := directive.QueryFragment()
			if err != nil {
				return nil, err
			}

			d, ok := directive.(*wcDirective)
			// For other Directive.
			if !ok {
				fragments = append(fragments, fragment)
				continue
			}

			// For wcDirective, add start/end marker
			directives = append(directives, d)
			n := len(directives) - 1
			fragments = append(fragments,
				fmt.Sprintf("1 AS %s, ", fmtMarker(n, true)),
				fragment,
				fmt.Sprintf(", 1 AS %s", fmtMarker(n, false)),
			)
		}
		query = strings.TrimSpace(strings.Join(fragments, ""))
	}

	// Query
	resultCols, err := loader.LoadCols(query)
	if err != nil {
		return nil, err
	}

	// Collect wildcards info.
	info := &WildcardsInfo{}
	{
		curN := -1
		curWildcardInfo := (*WildcardInfo)(nil)
		resultCols2 := []datasrc.Col{}

		for _, resultCol := range resultCols {

			ok, n, start := parseMarker(resultCol.Name)
			// Normal column.
			if !ok {
				resultCols2 = append(resultCols2, resultCol)
				continue
			}

			// Marker column.
			directive := directives[n]
			if start {
				// Start marker.
				if curN >= 0 || curWildcardInfo != nil {
					return nil, fmt.Errorf("Expect no wildcard start marker.")
				}

				curWildcardInfo = &WildcardInfo{
					table:  directive.table,
					alias:  directive.alias,
					offset: len(resultCols2),
				}
				curN = n
				info.wildcards = append(info.wildcards, curWildcardInfo)
			} else {
				// End marker.
				if curN < 0 || curWildcardInfo == nil {
					return nil, fmt.Errorf("Expect wildcard start marker.")
				}
				if n != curN {
					return nil, fmt.Errorf("Wildcard start/end marker mismatch: %d!=%d", curN, n)
				}
				// Column numbers between start/end markers should be the same as the number of table columns.
				if len(resultCols2) != curWildcardInfo.offset+curWildcardInfo.table.NumColumn() {
					return nil, fmt.Errorf("Wildcard column number mismatch.")
				}
				curN = -1
				curWildcardInfo = nil
			}

		}

		// Markers should be in pairs and properly closed.
		if curN >= 0 || curWildcardInfo != nil {
			return nil, fmt.Errorf("Expect not inside wildcard markers.")
		}

		// Check result columns
		resultCols = resultCols2
		expectResultCols := stmt.QueryResultCols()
		if len(expectResultCols) != len(resultCols) {
			return nil, fmt.Errorf("Query result column number mismatch: %d!=%d.", len(expectResultCols), len(resultCols))
		}
		for i, resultCol := range resultCols {
			if expectResultCols[i] != resultCol {
				return nil, fmt.Errorf("Query result column[%d] mismatch: %#v!=%#v.", i, expectResultCols[i], resultCols)
			}
		}

	}

	// Fill resultCols2Wildcard
	{
		info.resultCols2Wildcard = make([]int, len(resultCols))
		for i := 0; i < len(info.resultCols2Wildcard); i++ {
			info.resultCols2Wildcard[i] = -1
		}
		for i, wildcardInfo := range info.wildcards {
			for j := 0; j < wildcardInfo.table.NumColumn(); j++ {
				info.resultCols2Wildcard[wildcardInfo.offset+j] = i
			}
		}

	}

	return info, nil
}

// Valid returns true if info != nil.
func (info *WildcardsInfo) Valid() bool {
	return info != nil
}

// Wildcards returns all WildcardInfo in a statement.
func (info *WildcardsInfo) Wildcards() []*WildcardInfo {
	if info == nil {
		return nil
	}
	return info.wildcards
}

// WildcardColumn returns the i-th query result column as a wildcard expansion table column.
// It returns nil if info is nil or i is out of range, or the i-th query result column is not from a wildcard expansion.
func (info *WildcardsInfo) WildcardColumn(i int) *infos.ColumnInfo {
	if info == nil {
		return nil
	}
	if i < 0 || i >= len(info.resultCols2Wildcard) {
		return nil
	}
	idx := info.resultCols2Wildcard[i]
	if idx < 0 {
		return nil
	}
	wildcard := info.wildcards[idx]
	return wildcard.table.Column(i - wildcard.offset)
}

// WildcardName returns the wildcard name of the i-th query result column.
// It returns "" if info is nil or i is out of range, or the i-th query result column is not from a wildcard expansion.
func (info *WildcardsInfo) WildcardName(i int) string {
	if info == nil {
		return ""
	}
	if i < 0 || i >= len(info.resultCols2Wildcard) {
		return ""
	}
	idx := info.resultCols2Wildcard[i]
	if idx < 0 {
		return ""
	}
	return info.wildcards[idx].WildcardName()

}

// SingleWildcard returns true if result columns of the statement are all from a single wildcard expansion.
func (info *WildcardsInfo) SingleWildcard() bool {
	if info == nil {
		return false
	}
	return len(info.wildcards) == 1 && info.wildcards[0].table.NumColumn() == len(info.resultCols2Wildcard)
}

// Valid returns true if info is not nil.
func (info *WildcardInfo) Valid() bool {
	return info != nil
}

// Table returns the wildcard table.
func (info *WildcardInfo) Table() *infos.TableInfo {
	if info == nil {
		return nil
	}
	return info.table
}

// Alias returns the optinal wildcard alias.
func (info *WildcardInfo) Alias() string {
	if info == nil {
		return ""
	}
	return info.alias
}

// WildcardName returns alias/table name.
func (info *WildcardInfo) WildcardName() string {
	if info == nil {
		return ""
	}
	if info.alias != "" {
		return info.alias
	}
	return info.table.TableName()
}

// Offset returns the offset of this wildcard expansion in query result columns. It returns -1 if info is nil.
func (info *WildcardInfo) Offset() int {
	if info == nil {
		return -1
	}
	return info.offset
}

func (d *wcDirective) Initialize(loader *datasrc.Loader, db *infos.DBInfo, stmt *infos.StmtInfo, tok etree.Token) error {

	// Extract Table info.
	elem := tok.(*etree.Element)
	tableName := elem.SelectAttrValue("table", "")
	if tableName == "" {
		return fmt.Errorf("Missing 'table' attribute in <wc> directive")
	}

	table := db.TableByName(tableName)
	if table == nil {
		return fmt.Errorf("Table %+q not found", tableName)
	}

	// Optinally alias
	alias := elem.SelectAttrValue("as", "")

	// Set fields
	d.loader = loader
	d.db = db
	d.stmt = stmt
	d.table = table
	d.alias = alias
	return nil

}

func (d *wcDirective) QueryFragment() (string, error) {

	// Expands to fields list.
	prefix := d.name()
	fragments := []string{}
	for i := 0; i < d.table.NumColumn(); i++ {
		if i != 0 {
			fragments = append(fragments, ", ")
		}
		fragments = append(fragments, fmt.Sprintf("`%s`.`%s`", prefix, d.table.Column(i).ColumnName()))
	}
	return strings.Join(fragments, ""), nil

}

func (d *wcDirective) TextFragment() (string, error) {
	// The same as QueryFragment.
	return d.QueryFragment()
}

func (d *wcDirective) ExtraProcess() error {
	locals := d.stmt.Locals(wcLocalsKey)
	if locals != nil {
		// Already has WildcardsInfo created, do nothing.
		return nil
	}

	// Creates a WildcardsInfo for this statement.
	info, err := newWildcardsInfo(d.loader, d.db, d.stmt)
	if err != nil {
		return err
	}
	d.stmt.SetLocals(wcLocalsKey, info)
	return nil
}

func (d *wcDirective) name() string {
	if d.alias != "" {
		return d.alias
	}
	return d.table.TableName()
}

func init() {
	infos.RegistDirectiveFactory(func() infos.Directive {
		return &wcDirective{}
	}, "wc")
}
