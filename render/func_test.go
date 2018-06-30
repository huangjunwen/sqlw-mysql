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
		if testCase.ExpectError {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}
		assert.Equal(testCase.Expect, result)

	}
}

func TestEnum(t *testing.T) {

	assert := assert.New(t)

	for _, testCase := range []struct {
		Args        []int
		Expect      []int
		ExpectError bool
	}{
		{[]int{-1}, []int{}, true}, // end < start
		{[]int{0}, []int{}, false},
		{[]int{2}, []int{0, 1}, false},

		{[]int{1, 0}, []int{}, true}, // end < start
		{[]int{1, 1}, []int{}, false},
		{[]int{1, 2}, []int{1}, false},

		{[]int{1, 0, 2}, []int{}, true}, // end < start
		{[]int{0, 0, 2}, []int{}, false},
		{[]int{0, 4, 2}, []int{0, 2}, false},
		{[]int{0, 1, -2}, []int{}, true}, // end > start
		{[]int{0, 0, -2}, []int{}, false},
		{[]int{4, 0, -2}, []int{4, 2}, false},
	} {
		c, err := enum(testCase.Args...)
		if testCase.ExpectError {
			assert.Error(err)
		} else {
			assert.NoError(err)
			actual := []int{}
			if c != nil {
				for i := range c {
					actual = append(actual, i)
				}
			}
			assert.Equal(testCase.Expect, actual)
		}
	}

}
