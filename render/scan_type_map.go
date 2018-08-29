package render

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/infos"
)

// ScanTypeMap maps data type to scan type.
// [0] is for not nullable types.
// [1] is for nullable types.
type ScanTypeMap map[string][2]string

// LoadScanTypeMap loads scan type map from io.Reader.
func LoadScanTypeMap(r io.Reader) (ScanTypeMap, error) {
	ret := ScanTypeMap{}
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&ret); err != nil {
		return nil, err
	}
	for dataType, scanTypes := range ret {
		if scanTypes[0] == "" {
			return nil, fmt.Errorf("Data type %+q has no not-nullable scan type", dataType)
		}
		if scanTypes[1] == "" {
			return nil, fmt.Errorf("Data type %+q has no nullable scan type", dataType)
		}
	}
	return ret, nil

}

func (m ScanTypeMap) scanType(val interface{}, i int) (string, error) {

	if m == nil {
		return "", fmt.Errorf("ScanTypeMap is empty")
	}

	dataType, nullable := "", true

	switch v := val.(type) {
	case datasrc.Col:
		dataType = v.DataType
		nullable = v.Nullable

	case *datasrc.Col:
		dataType = v.DataType
		nullable = v.Nullable

	case datasrc.Column:
		dataType = v.DataType
		nullable = v.Nullable

	case *datasrc.Column:
		dataType = v.DataType
		nullable = v.Nullable

	case *infos.ColumnInfo:
		dataType = v.DataType()
		nullable = v.Nullable()

	default:
		return "", fmt.Errorf("scanType: Expect table or query result column but got %T", val)
	}

	scanTypes, found := m[dataType]
	if !found {
		// Some default.
		scanTypes = [2]string{"[]byte", "[]byte"}
	}

	if i < 0 {
		if nullable {
			i = 1
		} else {
			i = 0
		}
	}

	return scanTypes[i], nil
}

// ScanType returns the scan type for the (table or query result) column.
func (m ScanTypeMap) ScanType(col interface{}) (string, error) {
	return m.scanType(col, -1)
}

// NotNullScanType returns the not nullable scan type for the (table or query result) column.
func (m ScanTypeMap) NotNullScanType(col interface{}) (string, error) {
	return m.scanType(col, 0)
}

// NullScanType returns the nullable scan type for the (table or query result) column.
func (m ScanTypeMap) NullScanType(col interface{}) (string, error) {
	return m.scanType(col, 1)
}
