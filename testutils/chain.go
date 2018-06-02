package testutils

import (
	"context"
)

// Chain is a chain of testing middlewares.
type Chain []func(func(context.Context) error) func(context.Context) error

// Then assembly all the testing middlewares in chain with do.
func (c Chain) Then(do func(context.Context) error) func(context.Context) error {
	for i := len(c) - 1; i >= 0; i-- {
		do = c[i](do)
	}
	return do
}
