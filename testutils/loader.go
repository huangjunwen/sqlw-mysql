package testutils

import (
	"context"
	"fmt"

	"github.com/huangjunwen/sqlw-mysql/datasrc"
)

type loaderKeyType struct{}

var (
	loaderKey = loaderKeyType{}
)

// TstLoader returns the test loader from context.
func TstLoader(ctx context.Context) *datasrc.Loader {
	v := ctx.Value(loaderKey)
	if v != nil {
		return v.(*datasrc.Loader)
	}
	return nil
}

// WithTstLoader creates a test loader.
func WithTstLoader(do func(context.Context) error) func(context.Context) error {

	return func(ctx context.Context) error {

		dsn := TstMySQLServerDSN(ctx)
		if dsn == "" {
			return fmt.Errorf("No TstMySQLServerDSN found.")
		}

		loader, err := datasrc.NewLoader(dsn)
		if err != nil {
			return err
		}
		defer loader.Close()

		ctx2 := context.WithValue(ctx, loaderKey, loader)
		return do(ctx2)

	}

}
