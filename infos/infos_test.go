package infos

import (
	"context"
	"log"
	"testing"

	"github.com/huangjunwen/sqlw-mysql/datasrc"

	"github.com/huangjunwen/tstsvc/tstsvc/mysql"
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
	log.Printf("Starting testing MySQL server.\n")
	mysqlSvc, dsn, err := tstmysql.Run()
	if err != nil {
		log.Panic(err)
	}
	defer mysqlSvc.Close()
	log.Printf("Testing MySQL server up.\n")

	loader, err = datasrc.NewLoader(dsn)
	if err != nil {
		log.Panic(err)
	}
	defer loader.Close()

	m.Run()
}
