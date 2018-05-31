package datasrc

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	// ConnectTimeout is the timeout for connecting database.
	ConnectTimeout = time.Second * 5

	// QueryTimeout is the timeout for querying database.
	QueryTimeout = time.Second * 5
)

// Loader is used to load information from a MySQL database.
type Loader struct {
	dsn      string
	connPool *sql.DB
	// NOTE: Use this single conn instead of connPool to ensure querys are executing in the same session.
	conn *sql.Conn
}

func NewLoader(dsn string) (*Loader, error) {

	connPool, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	conn, err := connPool.Conn(connectCtx())
	if err != nil {
		return nil, err
	}

	return &Loader{
		dsn:      dsn,
		connPool: connPool,
		conn:     conn,
	}, nil

}

// Close release resources.
func (loader *Loader) Close() {
	loader.conn.Close()
	loader.connPool.Close()
}

// DSN returns the data source name.
func (loader *Loader) DSN() string {
	return loader.dsn
}

var (
	// Copy from github.com/go-sql-driver/mysql/fields.go
	scanTypeFloat32   = reflect.TypeOf(float32(0))
	scanTypeFloat64   = reflect.TypeOf(float64(0))
	scanTypeInt8      = reflect.TypeOf(int8(0))
	scanTypeInt16     = reflect.TypeOf(int16(0))
	scanTypeInt32     = reflect.TypeOf(int32(0))
	scanTypeInt64     = reflect.TypeOf(int64(0))
	scanTypeNullFloat = reflect.TypeOf(sql.NullFloat64{})
	scanTypeNullInt   = reflect.TypeOf(sql.NullInt64{})
	scanTypeNullTime  = reflect.TypeOf(mysql.NullTime{})
	scanTypeUint8     = reflect.TypeOf(uint8(0))
	scanTypeUint16    = reflect.TypeOf(uint16(0))
	scanTypeUint32    = reflect.TypeOf(uint32(0))
	scanTypeUint64    = reflect.TypeOf(uint64(0))
	scanTypeRawBytes  = reflect.TypeOf(sql.RawBytes{})
	scanTypeUnknown   = reflect.TypeOf(new(interface{}))
)

const (
	// Copy from github.com/go-sql-driver/mysql/const.go
	flagUnsigned = 1 << 5
)

// LoadCols loads result columns of a query.
func (loader *Loader) LoadCols(query string, args ...interface{}) ([]Col, error) {

	cols := []Col{}

	rows, err := loader.conn.QueryContext(queryCtx(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cts, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	for i, ct := range cts {

		scanType := ct.ScanType()
		databaseTypeName := ct.DatabaseTypeName()
		nullable, ok := ct.Nullable()
		if !ok {
			panic(fmt.Errorf("ColumnType.Nullable() not supported in MySQL driver"))
		}

		// XXX: From current driver's public API some information is lost:
		// - Column type's length is not support yet (see https://github.com/go-sql-driver/mysql/pull/667)
		// - Unsigned or not can't be determine when if type is NullInt64
		// Do some tricks to read them from private fields.
		//
		// XXX: In general reading data from private field is not a good idea, but i think here
		// is ok since we're only using them to generate code
		field := reflect.
			ValueOf(rows).          // *sql.Rows
			Elem().                 // sql.Rows
			FieldByName("rowsi").   // driver.Rows
			Elem().                 // *mysql.mysqlRows
			Elem().                 // mysql.mysqlRows
			FieldByName("rs").      // mysql.resultSet
			FieldByName("columns"). // []mysql.mysqlField
			Index(i)                // mysql.mysqlField

		length := field.FieldByName("length").Uint()
		flags := field.FieldByName("flags").Uint()
		unsigned := (flags & flagUnsigned) != 0

		bad := func() {
			panic(fmt.Errorf("Unsupported column type: ScanType=%#v, DatabaseTypeName=%+q", scanType, databaseTypeName))
		}

		// Translate to DataType.
		dataType := ""
		switch scanType {
		// Float types
		case scanTypeFloat32:
			dataType = "float32"
		case scanTypeFloat64:
			dataType = "float64"
		case scanTypeNullFloat:
			switch databaseTypeName {
			case "FLOAT":
				dataType = "float32"
			case "DOUBLE":
				dataType = "float64"
			default:
				bad()
			}

		// Int types, includeing bool type
		case scanTypeInt8:
			if length == 1 {
				// Special case for bool
				dataType = "bool"
			} else {
				dataType = "int8"
			}
		case scanTypeInt16:
			dataType = "int16"
		case scanTypeInt32:
			dataType = "int32"
		case scanTypeInt64:
			dataType = "int64"
		case scanTypeUint8:
			dataType = "uint8"
		case scanTypeUint16:
			dataType = "uint16"
		case scanTypeUint32:
			dataType = "uint32"
		case scanTypeUint64:
			dataType = "uint64"
		case scanTypeNullInt:
			switch databaseTypeName {
			case "TINYINT":
				if unsigned {
					dataType = "uint8"
				} else {
					if length == 1 {
						dataType = "bool"
					} else {
						dataType = "int8"
					}
				}
			case "SMALLINT", "YEAR":
				if unsigned {
					dataType = "uint16"
				} else {
					dataType = "int16"
				}
			case "MEDIUMINT", "INT":
				if unsigned {
					dataType = "uint32"
				} else {
					dataType = "int32"
				}
			case "BIGINT":
				if unsigned {
					dataType = "uint64"
				} else {
					dataType = "int64"
				}
			default:
				bad()
			}

		// Time types
		case scanTypeNullTime:
			dataType = "time"

			// String types
		case scanTypeRawBytes:
			switch databaseTypeName {
			case "BIT":
				dataType = "bit"
			case "JSON":
				dataType = "json"
			default:
				dataType = "string"
			}

		default:
			bad()
		}

		cols = append(cols, Col{
			Name:     ct.Name(),
			DataType: dataType,
			Nullable: nullable,
		})

	}

	return cols, nil

}

// LoadDBName returns current database name.
func (loader *Loader) LoadDBName() (string, error) {

	var dbName sql.NullString
	// NOTE: https://dev.mysql.com/doc/refman/5.7/en/information-functions.html#function_database
	// SELECT DATABASE() returns current database or NULL if there is no current default database.
	err := loader.conn.QueryRowContext(queryCtx(), "SELECT DATABASE()").Scan(&dbName)
	if err != nil {
		return "", err
	}
	if dbName.String == "" {
		return "", fmt.Errorf("No database selected")
	}
	return dbName.String, nil
}

// LoadTableNames returns all normal table names in current database.
func (loader *Loader) LoadTableNames() ([]string, error) {

	tableNames := []string{}

	dbName, err := loader.LoadDBName()
	if err != nil {
		return nil, err
	}

	rows, err := loader.conn.QueryContext(queryCtx(), `
	SELECT 
		TABLE_NAME
	FROM
		INFORMATION_SCHEMA.TABLES
	WHERE
		TABLE_SCHEMA=? AND TABLE_TYPE='BASE TABLE'
	`, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		tableName := ""
		if err = rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, tableName)
	}
	return tableNames, nil
}

// LoadColumns returns all columns of the named table.
func (loader *Loader) LoadColumns(tableName string) ([]Column, error) {

	columns := []Column{}

	dbName, err := loader.LoadDBName()
	if err != nil {
		return nil, err
	}

	cols, err := loader.LoadCols("SELECT * FROM `" + tableName + "`")
	if err != nil {
		return nil, err
	}

	for i, col := range cols {

		row := loader.conn.QueryRowContext(queryCtx(), `
		SELECT
			IF(EXTRA='auto_increment', 'auto_increment', COLUMN_DEFAULT)
		FROM
			INFORMATION_SCHEMA.COLUMNS
		WHERE
			TABLE_SCHEMA=? AND TABLE_NAME=? AND COLUMN_NAME=?
		`, dbName, tableName, col.Name)

		defaultVal := sql.NullString{}
		if err := row.Scan(&defaultVal); err != nil {
			return nil, err
		}

		columns = append(columns, Column{
			Col:             col,
			Pos:             i,
			HasDefaultValue: defaultVal.Valid,
		})

	}

	return columns, nil

}

// LoadAutoIncColumn returns the auto_increment column name of the named table or "" if not exists.
func (loader *Loader) LoadAutoIncColumn(tableName string) (string, error) {

	dbName, err := loader.LoadDBName()
	if err != nil {
		return "", err
	}

	row := loader.conn.QueryRowContext(queryCtx(), `
	SELECT
		COLUMN_NAME
	FROM
		INFORMATION_SCHEMA.COLUMNS
	WHERE
		TABLE_SCHEMA=? AND TABLE_NAME=? AND EXTRA LIKE ?
	`, dbName, tableName, "%auto_increment%")

	var columnName string
	if err := row.Scan(&columnName); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return columnName, nil

}

// LoadIndexNames returns all index names of the named table.
func (loader *Loader) LoadIndexNames(tableName string) ([]string, error) {

	indexNames := []string{}

	dbName, err := loader.LoadDBName()
	if err != nil {
		return nil, err
	}

	rows, err := loader.conn.QueryContext(queryCtx(), `
	SELECT 
		DISTINCT INDEX_NAME 
	FROM 
		INFORMATION_SCHEMA.STATISTICS 
	WHERE 
		TABLE_SCHEMA=? AND TABLE_NAME=?`, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		indexName := ""
		if err = rows.Scan(&indexName); err != nil {
			return nil, err
		}
		indexNames = append(indexNames, indexName)
	}
	return indexNames, nil
}

// LoadIndex loads the named index information.
func (loader *Loader) LoadIndex(tableName, indexName string) (columnNames []string, isPrimary bool, isUnique bool, err error) {

	dbName, err := loader.LoadDBName()
	if err != nil {
		return nil, false, false, err
	}

	rows, err := loader.conn.QueryContext(queryCtx(), `
	SELECT 
		NON_UNIQUE, COLUMN_NAME, SEQ_IN_INDEX 
	FROM
		INFORMATION_SCHEMA.STATISTICS
	WHERE
		TABLE_SCHEMA=? AND TABLE_NAME=? AND INDEX_NAME=?
	ORDER BY SEQ_IN_INDEX`, dbName, tableName, indexName)
	if err != nil {
		return nil, false, false, err
	}
	defer rows.Close()

	notUnique := true
	prevSeq := 0
	for rows.Next() {
		columnName := ""
		seq := 0
		if err := rows.Scan(&notUnique, &columnName, &seq); err != nil {
			return nil, false, false, err
		}

		// Check seq.
		if seq != prevSeq+1 {
			panic(fmt.Errorf("Bad SEQ_IN_INDEX, prev is %d, current is %d", prevSeq, seq))
		}
		prevSeq = seq

		columnNames = append(columnNames, columnName)
	}

	if len(columnNames) == 0 {
		return nil, false, false, fmt.Errorf("Index %+q in table %+q not found", indexName, tableName)
	}

	// https://dev.mysql.com/doc/refman/5.7/en/create-table.html
	// The name of a PRIMARY KEY is always PRIMARY, which thus cannot be used as the name for any other kind of index
	isPrimary = indexName == "PRIMARY"
	isUnique = !notUnique
	return
}

// LoadFKNames returns all foreign key names of the named table.
func (loader *Loader) LoadFKNames(tableName string) ([]string, error) {

	fkNames := []string{}

	dbName, err := loader.LoadDBName()
	if err != nil {
		return nil, err
	}

	rows, err := loader.conn.QueryContext(queryCtx(), `
	SELECT
		CONSTRAINT_NAME
	FROM
		INFORMATION_SCHEMA.TABLE_CONSTRAINTS
	WHERE
		TABLE_SCHEMA=? AND TABLE_NAME = ? AND CONSTRAINT_TYPE = 'FOREIGN KEY'`, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		fkName := ""
		if err := rows.Scan(&fkName); err != nil {
			return nil, err
		}
		fkNames = append(fkNames, fkName)
	}
	return fkNames, nil
}

// LoadFK returns the named foreign key information.
func (loader *Loader) LoadFK(tableName, fkName string) (columnNames []string, refTableName string, refColumnNames []string, err error) {

	dbName, err := loader.LoadDBName()
	if err != nil {
		return nil, "", nil, err
	}

	rows, err := loader.conn.QueryContext(queryCtx(), `
		SELECT
			COLUMN_NAME, ORDINAL_POSITION, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME
		FROM
			INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE
			TABLE_SCHEMA=? AND TABLE_NAME=? AND CONSTRAINT_NAME=? ORDER BY ORDINAL_POSITION`, dbName, tableName, fkName)
	if err != nil {
		return nil, "", nil, err
	}
	defer rows.Close()

	prevPos := 0
	for rows.Next() {
		columnName := ""
		refColumnName := ""
		pos := 0
		if err := rows.Scan(&columnName, &pos, &refTableName, &refColumnName); err != nil {
			return nil, "", nil, err
		}

		// Check pos.
		if pos != prevPos+1 {
			panic(fmt.Errorf("Bad ORDINAL_POSITION, prev is %d, current is %d", prevPos, pos))
		}
		prevPos = pos

		columnNames = append(columnNames, columnName)
		refColumnNames = append(refColumnNames, refColumnName)
	}

	if len(columnNames) == 0 {
		return nil, "", nil, fmt.Errorf("FK %+q in table %+q not found", fkName, tableName)
	}

	return columnNames, refTableName, refColumnNames, nil
}

func connectCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), ConnectTimeout)
	return ctx
}

func queryCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), QueryTimeout)
	return ctx
}
