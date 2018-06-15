package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlice(t *testing.T) {

	assert := assert.New(t)

	for _, testCase := range []struct {
		S           string
		Args        []int
		Expect      string
		ExpectError bool
	}{
		{"hello", []int{0}, "hello", false},
		{"hello", []int{1, 2}, "e", false},
		{"hello", []int{1, 2, 3}, "", true},
		{"hello", []int{2, 1}, "", true},
		{"hello", []int{-1}, "", true},
		{"hello", []int{5, 6}, "", true},
	} {

		result, err := slice(testCase.S, testCase.Args...)
		assert.Equal(testCase.Expect, result)
		if testCase.ExpectError {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}

	}
}
