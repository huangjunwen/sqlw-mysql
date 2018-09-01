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
//   - https://github.com/go-sql-driver/mysql/blob/master/fields.go mysqlField
//   - https://github.com/go-sql-driver/mysql/blob/master/packets.go mysqlConn.readColumns
//   - https://dev.mysql.com/doc/internals/en/com-query-response.html#packet-Protocol::ColumnDefinition41
type ExtColumnType struct {
	sql.ColumnType
	tp     uint8
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
			tp:         uint8(mf.FieldByName("fieldType").Uint()),
			length:     uint32(mf.FieldByName("length").Uint()),
			flags:      uint16(mf.FieldByName("flags").Uint()),
		})
	}

	return ret, nil
}

// RawType returns the raw type code.
// See: https://dev.mysql.com/doc/internals/en/com-query-response.html#packet-Protocol::ColumnType
func (ect *ExtColumnType) RawType() uint8 {
	return ect.tp
}

// RawLength returns the raw length.
func (ect *ExtColumnType) RawLength() uint32 {
	return ect.length
}

// RawFlags returns the raw flags.
func (ect *ExtColumnType) RawFlags() uint16 {
	return ect.flags
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

	// See: https://github.com/go-sql-driver/mysql/blob/master/fields.go mysqlField.scanType
	switch ect.tp {
	// Int types.
	case tpTiny:
		if unsigned {
			return "uint8"
		}
		if length == 1 {
			return "bool"
		}
		return "int8"

	case tpShort, tpYear:
		if unsigned {
			return "uint16"
		}
		return "int16"

	case tpInt24, tpLong:
		if unsigned {
			return "uint32"
		}
		return "int32"

	case tpLongLong:
		if unsigned {
			return "uint64"
		}
		return "int64"

	// Float types.
	case tpFloat:
		return "float32"

	case tpDouble:
		return "float64"

	// Time types.
	case tpDate, tpNewDate, tpTimestamp, tpDateTime:
		return "time"

	// String types.
	case tpDecimal, tpNewDecimal:
		return "decimal"

	case tpBit:
		return "bit"

	case tpJSON:
		return "json"

	case tpVarChar, tpTinyBLOB, tpMediumBLOB, tpLongBLOB, tpBLOB, tpVarString,
		tpString, tpGeometry, tpTime, tpEnum, tpSet:
		return "string"

	default:
		return ""
	}
}

// Copy and modify from github.com/go-sql-driver/mysql/const.go
// https://dev.mysql.com/doc/internals/en/com-query-response.html#packet-Protocol::ColumnType
const (
	tpDecimal uint8 = iota
	tpTiny
	tpShort
	tpLong
	tpFloat
	tpDouble
	tpNULL
	tpTimestamp
	tpLongLong
	tpInt24
	tpDate
	tpTime
	tpDateTime
	tpYear
	tpNewDate
	tpVarChar
	tpBit
)
const (
	tpJSON uint8 = iota + 0xf5
	tpNewDecimal
	tpEnum
	tpSet
	tpTinyBLOB
	tpMediumBLOB
	tpLongBLOB
	tpBLOB
	tpVarString
	tpString
	tpGeometry
)
