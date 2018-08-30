package datasrc

import (
	"database/sql"
	"fmt"
	"reflect"
)

// ExtColumnType extends sql.ColumnType with some type information returned from wire but not exported.
//   - ColumnType.Length is not supported yet (https://github.com/go-sql-driver/mysql/pull/667).
//   - 'unsigned' or not can't be known if type ColumnType.ScanType returns NullInt64.
// See:
//   - github.com/go-sql-driver/mysql/fields.go:mysqlField
//   - github.com/go-sql-driver/mysql/packets.go:mysqlConn.readColumns
//   - https://dev.mysql.com/doc/internals/en/com-query-response.html#packet-Protocol::ColumnDefinition41
type ExtColumnType struct {
	sql.ColumnType
	typ    uint8
	length uint32
	flags  uint16
}

// ExtractExtColumnTypes extracts ExtColumnType(s) from sql.Rows.
func ExtractExtColumnTypes(rows *sql.Rows) ([]*ExtColumnType, error) {
	if rows == nil {
		panic(fmt.Errorf("ExtractColDef: Expect not nil rows"))
	}

	cts, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	// XXX: Read private field data.
	mfs := reflect.
		ValueOf(rows).         // *sql.Rows
		Elem().                // sql.Rows
		FieldByName("rowsi").  // driver.Rows
		Elem().                // *mysql.textRows
		Elem().                // mysql.textRows
		FieldByName("rs").     // mysql.resultSet
		FieldByName("columns") // []mysql.mysqlField

	if len(cts) != mfs.Len() {
		panic(fmt.Errorf("ExtractExtColumnTypes: len(ColumnTypes) != len(mysqlFields) %d != %d", len(cts), mfs.Len()))
	}

	ret := []*ExtColumnType{}
	for i, ct := range cts {
		mf := mfs.Index(i)
		ret = append(ret, &ExtColumnType{
			ColumnType: *ct,
			typ:        uint8(mf.FieldByName("fieldType").Uint()),
			length:     uint32(mf.FieldByName("length").Uint()),
			flags:      uint16(mf.FieldByName("flags").Uint()),
		})
	}

	return ret, nil
}

// Length returns the raw length.
func (ect *ExtColumnType) Length() uint32 {
	return ect.length
}

// Unsigned returns true if the column is unsigned.
func (ect *ExtColumnType) Unsigned() bool {
	// see: github.com/go-sql-driver/mysql/const.go:flagUnsigned
	return (ect.flags & (1 << 5)) != 0
}

// Nullable returns true if the column is nullable.
func (ect *ExtColumnType) Nullable() bool {
	nullable, ok := ect.ColumnType.Nullable()
	if !ok {
		panic(fmt.Errorf("ExtColumnType: ColumnType.Nullable() returns not ok"))
	}
	return nullable
}

// DataType is a 'translated' column type identifier (ignore nullability). Used in scan type mapping.
// Available data types:
//   - bool
//   - int8/uint8/int16/uint16/int32/uint32/int64/uint64
//   - float32/float64
//   - time
//   - decimal
//   - bit
//   - enum/set
//   - json
//   - string
// It returns "" for unknown type.
func (ect *ExtColumnType) DataType() string {
	unsigned := ect.Unsigned()
	length := ect.length

	// See: github.com/go-sql-driver/mysql/fields.go:mysqlField.scanType
	switch ect.typ {
	// Int types.
	case typTiny:
		if unsigned {
			return "uint8"
		}
		if length == 1 {
			return "bool"
		}
		return "int8"

	case typShort, typYear:
		if unsigned {
			return "uint16"
		}
		return "int16"

	case typInt24, typLong:
		if unsigned {
			return "uint32"
		}
		return "int32"

	case typLongLong:
		if unsigned {
			return "uint64"
		}
		return "int64"

	// Float types.
	case typFloat:
		return "float32"

	case typDouble:
		return "float64"

	// Time types.
	case typDate, typNewDate, typTimestamp, typDateTime:
		return "time"

	// String types.
	case typDecimal, typNewDecimal:
		return "decimal"

	case typBit:
		return "bit"

	case typEnum:
		return "enum"

	case typSet:
		return "set"

	case typJSON:
		return "json"

	case typVarChar, typTinyBLOB, typMediumBLOB, typLongBLOB, typBLOB, typVarString,
		typString, typGeometry, typTime:
		return "string"

	default:
		return ""
	}
}

// Copy and modify from github.com/go-sql-driver/mysql/const.go
// https://dev.mysql.com/doc/internals/en/com-query-response.html#packet-Protocol::ColumnType
const (
	typDecimal uint8 = iota
	typTiny
	typShort
	typLong
	typFloat
	typDouble
	typNULL
	typTimestamp
	typLongLong
	typInt24
	typDate
	typTime
	typDateTime
	typYear
	typNewDate
	typVarChar
	typBit
)
const (
	typJSON uint8 = iota + 0xf5
	typNewDecimal
	typEnum
	typSet
	typTinyBLOB
	typMediumBLOB
	typLongBLOB
	typBLOB
	typVarString
	typString
	typGeometry
)
