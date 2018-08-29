package datasrc

import (
	"database/sql"
)

// Col represents basic column information.
type Col struct {
	// Name is the column name.
	Name string

	// DataType is a 'translated' column type identifier (ignore nullable), such as:
	// - uint24
	// - json
	// - time
	// It is used in scan type mapping only, thus can be any valid identifier, no need to be a real type name.
	DataType string

	// Nullable of the column.
	Nullable bool

	// CT is the raw column type.
	CT *sql.ColumnType
}

// Column represents a table column.
type Column struct {
	Col

	// Pos is the position of the column in table.
	Pos int

	// HasDefaultValue is true if the column has default value.
	HasDefaultValue bool
}
