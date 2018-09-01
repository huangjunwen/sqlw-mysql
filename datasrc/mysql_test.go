package datasrc_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/huangjunwen/sqlw-mysql/datasrc"

	"github.com/stretchr/testify/assert"
)

func TestExtColumnType(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE ext_column_type")
	exec("USE ext_column_type")
	defer func() {
		exec("DROP DATABASE ext_column_type")
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
		" `varchar_string_253_n_n` VARCHAR(128) NOT NULL," +
		" `varchar_string_253_y_n` VARCHAR(64)" +
		")")

	rows, err := loader.Conn().QueryContext(context.Background(), "SELECT * FROM types")
	assert.NoError(err)
	defer rows.Close()

	ects, err := datasrc.ExtractExtColumnTypes(rows)
	assert.NoError(err)

	for _, ect := range ects {
		name := ect.Name()
		parts := strings.Split(name, "_")

		assert.Equal(parts[1], ect.DataType(), "%+q: data type mismatch", name)

		tp, _ := strconv.Atoi(parts[2])
		assert.Equal(tp, int(ect.RawType()), "%+q: raw type mismatch", name)

		assert.Equal(parts[3] == "y", ect.Nullable(), "%+q: nullable mismatch", name)

		assert.Equal(parts[4] == "y", ect.Unsigned(), "%+q: unsigned mismatch", name)
	}
}
