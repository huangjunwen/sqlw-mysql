package datasrc_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadCols(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE load_cols")
	exec("USE load_cols")
	defer func() {
		exec("DROP DATABASE load_cols")
	}()

	exec("" +
		"CREATE TABLE `types` (" +
		" `float_z_float32` FLOAT NOT NULL, " +
		" `float_n_float32` FLOAT, " +
		" `double_z_float64` DOUBLE NOT NULL, " +
		" `double_n_float64` DOUBLE, " +
		" `bool_z_bool` BOOL NOT NULL, " +
		" `bool_n_bool` BOOL, " +
		" `tiny_z_int8` TINYINT NOT NULL, " +
		" `tiny_n_int8` TINYINT, " +
		" `utiny_z_uint8` TINYINT UNSIGNED NOT NULL, " +
		" `utiny_n_uint8` TINYINT UNSIGNED, " +
		" `small_z_int16` SMALLINT NOT NULL, " +
		" `small_n_int16` SMALLINT, " +
		" `usmall_z_uint16` SMALLINT UNSIGNED NOT NULL, " +
		" `usmall_n_uint16` SMALLINT UNSIGNED, " +
		" `medium_z_int32` MEDIUMINT NOT NULL, " +
		" `medium_n_int32` MEDIUMINT, " +
		" `umedium_z_uint32` MEDIUMINT UNSIGNED NOT NULL, " +
		" `umedium_n_uint32` MEDIUMINT UNSIGNED, " +
		" `int_z_int32` INT NOT NULL, " +
		" `int_n_int32` INT, " +
		" `uint_z_uint32` INT UNSIGNED NOT NULL, " +
		" `uint_n_uint32` INT UNSIGNED, " +
		" `big_z_int64` BIGINT NOT NULL, " +
		" `big_n_int64` BIGINT, " +
		" `ubig_z_uint64` BIGINT UNSIGNED NOT NULL, " +
		" `ubig_n_uint64` BIGINT UNSIGNED, " +
		" `datetime_z_time` DATETIME NOT NULL, " +
		" `datetime_n_time` DATETIME, " +
		" `date_z_time` DATE NOT NULL, " +
		" `date_n_time` DATE, " +
		" `timestamp_z_time` TIMESTAMP NOT NULL DEFAULT NOW(), " +
		" `timestamp_n_time` TIMESTAMP NULL DEFAULT NOW(), " + // XXX
		" `bit_z_bit` BIT(10) NOT NULL, " +
		" `bit_n_bit` BIT(10), " +
		" `json_z_json` JSON NOT NULL, " +
		" `json_n_json` JSON, " +
		" `char_z_string` CHAR(32) NOT NULL, " +
		" `char_n_string` CHAR(32), " +
		" `vchar_z_string` VARCHAR(32) NOT NULL, " +
		" `vchar_n_string` VARCHAR(32), " +
		" `text_z_string` TEXT NOT NULL, " +
		" `text_n_string` TEXT, " +
		" `blob_z_string` BLOB NOT NULL, " +
		" `blob_n_string` BLOB " +
		")")

	cols, err := loader.LoadCols("SELECT * FROM `types`")
	assert.NoError(err)
	for _, col := range cols {
		parts := strings.Split(col.Name, "_")
		assert.Equal(parts[1] == "n", col.Nullable)
		assert.Equal(parts[2], col.DataType)
	}

}

func TestLoadDBName(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE load_dbname")
	exec("USE load_dbname")

	{
		dbName, err := loader.LoadDBName()
		assert.NoError(err)
		assert.Equal("load_dbname", dbName)
	}

	exec("DROP DATABASE load_dbname")

	{
		dbName, err := loader.LoadDBName()
		assert.Error(err)
		assert.Equal("", dbName)
	}

}

func TestLoadTableNames(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE load_tablenames")
	exec("USE load_tablenames")
	defer func() {
		exec("DROP DATABASE load_tablenames")
	}()

	exec("CREATE TABLE `item` (qty INT, price INT)")
	exec("CREATE VIEW `item2` AS SELECT qty, price, qty*price AS value FROM `item`")

	{
		tableNames, err := loader.LoadTableNames()
		assert.NoError(err)
		assert.Len(tableNames, 1)
		assert.Contains(tableNames, "item")
		assert.NotContains(tableNames, "item2")
	}

}

func TestLoadColumns(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE load_columns")
	exec("USE load_columns")
	defer func() {
		exec("DROP DATABASE load_columns")
	}()

	exec("" +
		"CREATE TABLE `dft` (" +
		" `int_n_nodft` INT, " +
		" `int_z_nodft` INT NOT NULL, " +
		" `int_n_dft` INT AUTO_INCREMENT, " +
		" `int_z_dft` INT NOT NULL DEFAULT 101, " +
		" `ts_nimplicit_dft` TIMESTAMP, " +
		//" `ts_z` TIMESTAMP NOT NULL, " + // NOTE: Not allowed
		" `ts_n_dft` TIMESTAMP DEFAULT NOW(), " +
		" `ts_z_dft` TIMESTAMP NOT NULL DEFAULT '2018-01-01 00:00:00', " +
		" `dt_n_nodft` DATETIME, " +
		" `dt_z_nodft` DATETIME NOT NULL, " +
		" `dt_n_dft` DATETIME DEFAULT '2018-01-01 00:00:01', " +
		" `dt_z_dft` DATETIME NOT NULL DEFAULT NOW(), " +
		" `char_n_nodft` CHAR(32), " +
		" `char_z_nodft` CHAR(32) NOT NULL, " +
		" `char_n_dft` CHAR(32) DEFAULT '', " +
		" `char_z_dft` CHAR(32) NOT NULL DEFAULT 'NULL', " +
		" PRIMARY KEY (`int_n_dft`)" +
		")")

	columns, err := loader.LoadColumns("dft")
	assert.NoError(err)
	for _, column := range columns {
		parts := strings.Split(column.Name, "_")
		if parts[2] == "dft" {
			assert.True(column.HasDefaultValue)
		} else {
			assert.False(column.HasDefaultValue)
		}
	}

}

func TestLoadAutoIncColumn(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE load_autoinccolumn")
	exec("USE load_autoinccolumn")
	defer func() {
		exec("DROP DATABASE load_autoinccolumn")
	}()

	exec("CREATE TABLE `has_ai` (id INT AUTO_INCREMENT PRIMARY KEY)")
	exec("CREATE TABLE `hasnt_ai` (id INT)")

	{
		aiColumn, err := loader.LoadAutoIncColumn("has_ai")
		assert.NoError(err)
		assert.Equal("id", aiColumn)
	}

	{
		aiColumn, err := loader.LoadAutoIncColumn("hasnt_ai")
		assert.NoError(err)
		assert.Equal("", aiColumn)
	}

}

func TestLoadIndex(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE load_index")
	exec("USE load_index")
	defer func() {
		exec("DROP DATABASE load_index")
	}()

	exec("CREATE TABLE `no_indices` (id INT)")
	exec("CREATE TABLE `indices` (a INT PRIMARY KEY, b VARCHAR(32), c INT, d INT, UNIQUE KEY `idx_b` (b), KEY `idx_cd` (c, d))")

	{
		indexNames, err := loader.LoadIndexNames("no_indices")
		assert.NoError(err)
		assert.Len(indexNames, 0)
	}

	{
		indexNames, err := loader.LoadIndexNames("indices")
		assert.NoError(err)
		assert.Len(indexNames, 3)
		assert.Contains(indexNames, "PRIMARY")
		assert.Contains(indexNames, "idx_b")
		assert.Contains(indexNames, "idx_cd")
	}

	{
		columnNames, isPrimary, isUnique, err := loader.LoadIndex("indices", "PRIMARY")
		assert.NoError(err)
		assert.Len(columnNames, 1)
		assert.Equal("a", columnNames[0])
		assert.True(isPrimary)
		assert.True(isUnique)
	}

	{
		columnNames, isPrimary, isUnique, err := loader.LoadIndex("indices", "idx_b")
		assert.NoError(err)
		assert.Len(columnNames, 1)
		assert.Equal("b", columnNames[0])
		assert.False(isPrimary)
		assert.True(isUnique)
	}

	{
		columnNames, isPrimary, isUnique, err := loader.LoadIndex("indices", "idx_cd")
		assert.NoError(err)
		assert.Len(columnNames, 2)
		assert.Equal("c", columnNames[0])
		assert.Equal("d", columnNames[1])
		assert.False(isPrimary)
		assert.False(isUnique)
	}

	{
		columnNames, isPrimary, isUnique, err := loader.LoadIndex("indices", "idx_notexists")
		assert.Error(err)
		assert.Len(columnNames, 0)
		assert.False(isPrimary)
		assert.False(isUnique)
	}

}

func TestLoadFK(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE load_fk")
	exec("USE load_fk")
	defer func() {
		exec("DROP DATABASE load_fk")
	}()

	exec("CREATE TABLE `a` (u1 INT, u2 INT, UNIQUE KEY (u1, u2))")
	exec("CREATE TABLE `b` (id INT PRIMARY KEY, au1 INT, au2 INT, FOREIGN KEY (au1, au2) REFERENCES `a` (u1, u2))")

	{
		fkNames, err := loader.LoadFKNames("a")
		assert.NoError(err)
		assert.Len(fkNames, 0)
	}

	fkName := ""
	{
		fkNames, err := loader.LoadFKNames("b")
		assert.NoError(err)
		assert.Len(fkNames, 1)
		fkName = fkNames[0]
	}

	{
		columnNames, refTableName, refColumnNames, err := loader.LoadFK("b", fkName)
		assert.NoError(err)
		assert.Equal("a", refTableName)
		assert.Len(columnNames, 2)
		assert.Len(refColumnNames, 2)
		assert.Equal("au1", columnNames[0])
		assert.Equal("au2", columnNames[1])
		assert.Equal("u1", refColumnNames[0])
		assert.Equal("u2", refColumnNames[1])
	}

}
