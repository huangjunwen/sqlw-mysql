package datasrc

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadColumns(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE load_columns")
	exec("USE load_columns")
	defer func() {
		exec("DROP DATABASE load_columns")
	}()

	// Field format "<name>_<data type>_<tp>_<nullable>_<unsigned>"
	exec("" +
		"CREATE TABLE `types` (" +
		" `bool_bool_1_n_n` BOOL NOT NULL, " +
		" `bool_bool_1_y_n` BOOL, " +
		" `tiny_int8_1_n_n` TINYINT NOT NULL, " +
		" `tiny_int8_1_y_n` TINYINT, " +
		" `tiny_uint8_1_n_y` TINYINT UNSIGNED NOT NULL, " +
		" `tiny_uint8_1_y_y` TINYINT UNSIGNED, " +
		" `small_int16_2_n_n` SMALLINT NOT NULL, " +
		" `small_int16_2_y_n` SMALLINT, " +
		" `small_uint16_2_n_y` SMALLINT UNSIGNED NOT NULL, " +
		" `small_uint16_2_y_y` SMALLINT UNSIGNED, " +
		" `year_uint16_13_n_y` YEAR NOT NULL, " + // XXX: YEAR is always unsigned
		" `year_uint16_13_y_y` YEAR, " +
		" `medium_int32_9_n_n` MEDIUMINT NOT NULL, " +
		" `medium_int32_9_y_n` MEDIUMINT, " +
		" `medium_uint32_9_n_y` MEDIUMINT UNSIGNED NOT NULL, " +
		" `medium_uint32_9_y_y` MEDIUMINT UNSIGNED, " +
		" `int_int32_3_n_n` INT NOT NULL, " +
		" `int_int32_3_y_n` INT, " +
		" `int_uint32_3_n_y` INT UNSIGNED NOT NULL, " +
		" `int_uint32_3_y_y` INT UNSIGNED , " +
		" `big_int64_8_n_n` BIGINT NOT NULL, " +
		" `big_int64_8_y_n` BIGINT, " +
		" `big_uint64_8_n_y` BIGINT UNSIGNED NOT NULL, " +
		" `big_uint64_8_y_y` BIGINT UNSIGNED, " +
		" `float_float32_4_n_n` FLOAT NOT NULL, " +
		" `float_float32_4_y_n` FLOAT, " +
		" `double_float64_5_n_n` DOUBLE NOT NULL, " +
		" `double_float64_5_y_n` DOUBLE, " +
		" `date_time_10_n_n` DATE NOT NULL, " +
		" `date_time_10_y_n` DATE, " +
		" `timestamp_time_7_n_n` TIMESTAMP NOT NULL DEFAULT NOW(), " +
		" `timestamp_time_7_y_n` TIMESTAMP NULL DEFAULT NOW(), " +
		" `datetime_time_12_n_n` DATETIME NOT NULL, " +
		" `datetime_time_12_y_n` DATETIME, " +
		" `decimal_decimal_246_n_n` DECIMAL(9, 2) NOT NULL," +
		" `decimal_decimal_246_y_n` DECIMAL(10, 3)," +
		" `bit_bit_16_n_y` BIT(10) NOT NULL, " + // XXX: bit is always unsigned
		" `bit_bit_16_y_y` BIT(10), " +
		" `json_json_245_n_n` JSON NOT NULL, " +
		" `json_json_245_y_n` JSON, " +
		" `char_string_254_n_n` CHAR(100) NOT NULL, " +
		" `char_string_254_y_n` CHAR(1), " +
		" `varchar_string_253_n_n` VARCHAR(128) NOT NULL," +
		" `varchar_string_253_y_n` VARCHAR(64)" +
		")")

	cols, err := loader.LoadColumns("SELECT * FROM types")
	assert.NoError(err)

	for _, col := range cols {
		name := col.Name()
		parts := strings.Split(name, "_")

		assert.Equal(parts[1], col.DataType(), "%+q: data type mismatch", name)

		tp, _ := strconv.Atoi(parts[2])
		assert.Equal(tp, int(col.RawType()), "%+q: raw type mismatch", name)

		assert.Equal(parts[3] == "y", col.Nullable(), "%+q: nullable mismatch", name)

		assert.Equal(parts[4] == "y", col.Unsigned(), "%+q: unsigned mismatch", name)
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

func TestLoadTableColumns(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE load_table_columns")
	exec("USE load_table_columns")
	defer func() {
		exec("DROP DATABASE load_table_columns")
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

	cols, hasDefault, err := loader.LoadTableColumns("dft")
	assert.NoError(err)

	for i, col := range cols {
		parts := strings.Split(col.Name(), "_")
		if parts[2] == "dft" {
			assert.True(hasDefault[i])
		} else {
			assert.False(hasDefault[i])
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
		assert.Equal([]string{"a"}, columnNames)
		assert.True(isPrimary)
		assert.True(isUnique)
	}

	{
		columnNames, isPrimary, isUnique, err := loader.LoadIndex("indices", "idx_b")
		assert.NoError(err)
		assert.Equal([]string{"b"}, columnNames)
		assert.False(isPrimary)
		assert.True(isUnique)
	}

	{
		columnNames, isPrimary, isUnique, err := loader.LoadIndex("indices", "idx_cd")
		assert.NoError(err)
		assert.Equal([]string{"c", "d"}, columnNames)
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
		assert.Equal([]string{"au1", "au2"}, columnNames)
		assert.Equal([]string{"u1", "u2"}, refColumnNames)
	}

}
