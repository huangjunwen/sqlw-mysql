package render

import (
	"encoding/json"
	"fmt"
	"io"
)

// ScanTypeMap maps data type to scan type.
// [0] is for not nullable types.
// [1] is for nullable types.
type ScanTypeMap map[string][2]string

// NewScanTypeMap loads scan type map from io.Reader.
func NewScanTypeMap(r io.Reader) (ScanTypeMap, error) {
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
