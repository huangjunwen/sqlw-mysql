package datasrc

import (
	"context"
	"database/sql"
	"fmt"
	"time"
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

// NewLoader creates a loader.
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

// Conn returns the connection object. This connection object is also used by methods of Loader.
func (loader *Loader) Conn() *sql.Conn {
	return loader.conn
}

// LoadColumns loads result columns of a query.
func (loader *Loader) LoadColumns(query string, args ...interface{}) ([]*ExtColumnType, error) {

	rows, err := loader.conn.QueryContext(queryCtx(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// NOTE: Only return columns of the first result set.
	return ExtractExtColumnTypes(rows)

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

// LoadTableColumns returns all columns of the named table.
func (loader *Loader) LoadTableColumns(tableName string) ([]*ExtColumnType, []bool, error) {

	dbName, err := loader.LoadDBName()
	if err != nil {
		return nil, nil, err
	}

	ects, err := loader.LoadColumns("SELECT * FROM `" + tableName + "`")
	if err != nil {
		return nil, nil, err
	}

	hasDefault := []bool{}

	for _, ect := range ects {

		row := loader.conn.QueryRowContext(queryCtx(), `
		SELECT
			IF(EXTRA='auto_increment', 'auto_increment', COLUMN_DEFAULT)
		FROM
			INFORMATION_SCHEMA.COLUMNS
		WHERE
			TABLE_SCHEMA=? AND TABLE_NAME=? AND COLUMN_NAME=?
		`, dbName, tableName, ect.Name())

		defaultVal := sql.NullString{}
		if err := row.Scan(&defaultVal); err != nil {
			return nil, nil, err
		}

		hasDefault = append(hasDefault, defaultVal.Valid)
	}

	return ects, hasDefault, nil
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
