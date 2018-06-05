package render

import (
	"testing"

	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/stretchr/testify/assert"
)

func TestScanTypeMap(t *testing.T) {

	assert := assert.New(t)

	m := ScanTypeMap{
		"bool": [2]string{"bool", "null.Bool"},
	}

	for _, testCase := range []struct {
		Col                   interface{}
		ExpectScanType        string
		ExpectNotNullScanType string
		ExpectNullScanType    string
	}{
		{datasrc.Col{DataType: "bool", Nullable: true}, "null.Bool", "bool", "null.Bool"},
		{datasrc.Col{DataType: "bool", Nullable: false}, "bool", "bool", "null.Bool"},
		{datasrc.Col{DataType: "notexists", Nullable: false}, "[]byte", "[]byte", "[]byte"},
	} {

		{
			st, err := m.ScanType(testCase.Col)
			assert.NoError(err)
			assert.Equal(testCase.ExpectScanType, st)
		}
		{
			st, err := m.NotNullScanType(testCase.Col)
			assert.NoError(err)
			assert.Equal(testCase.ExpectNotNullScanType, st)
		}
		{
			st, err := m.NullScanType(testCase.Col)
			assert.NoError(err)
			assert.Equal(testCase.ExpectNullScanType, st)
		}

	}

}
