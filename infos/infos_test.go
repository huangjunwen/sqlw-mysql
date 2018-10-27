package infos

import (
	"context"
	"log"
	"testing"

	"github.com/huangjunwen/sqlw-mysql/datasrc"

	"github.com/huangjunwen/tstsvc/tstsvc"
	"github.com/huangjunwen/tstsvc/tstsvc/mysql"
	"github.com/ory/dockertest"
)

var (
	mysqlSvc *dockertest.Resource
	loader   *datasrc.Loader
)

func exec(query string, args ...interface{}) {
	_, err := loader.Conn().ExecContext(context.Background(), query, args...)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	log.Printf("Starting testing MySQL server.\n")
	var err error
	opts := &tstmysql.Options{
		HostPort: tstsvc.RandPort(),
	}
	mysqlSvc, err = opts.Run(nil)
	if err != nil {
		log.Panic(err)
	}
	defer mysqlSvc.Close()
	log.Printf("Testing MySQL server up.\n")

	loader, err = datasrc.NewLoader(opts.DSN())
	if err != nil {
		log.Panic(err)
	}
	defer loader.Close()
}
