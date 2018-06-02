package datasrc_test

import (
	"context"
	"log"
	"testing"

	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/testutils"
)

var (
	loader *datasrc.Loader
)

func exec(query string, args ...interface{}) {
	_, err := loader.Conn().ExecContext(context.Background(), query, args...)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {

	err := testutils.Chain{
		testutils.WithTstMySQLServer,
		testutils.WithTstLoader,
	}.Then(func(ctx context.Context) error {

		loader = testutils.TstLoader(ctx)
		m.Run()
		return nil

	})(context.Background())

	if err != nil {
		log.Fatal(err)
	}

}
