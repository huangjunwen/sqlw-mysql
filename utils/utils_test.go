package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamel(t *testing.T) {

	assert := assert.New(t)

	for _, testCase := range []struct {
		Origin      string
		ExpectUpper string
		ExpectLower string
	}{
		{"    ", "", ""},
		{"OneTwo", "OneTwo", "oneTwo"},
		{"_hello_world__", "HelloWorld", "helloWorld"},
		{"count(id)", "CountId", "countId"},
	} {

		assert.Equal(testCase.ExpectUpper, camel(testCase.Origin, true))
		assert.Equal(testCase.ExpectLower, camel(testCase.Origin, false))

	}

}
