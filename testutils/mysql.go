package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

const (
	defaultMySQLVersion = "5.7.21"
)

type noopLogger struct{}

func (l noopLogger) Print(v ...interface{}) {}

type dsnKeyType struct{}

var (
	dsnKey = dsnKeyType{}
)

// TstMySQLServerDSN returns the DSN of the test MySQL server from context.
func TstMySQLServerDSN(ctx context.Context) string {
	v := ctx.Value(dsnKey)
	if v != nil {
		return v.(string)
	}
	return ""
}

// WithTstMySQLServer setup a test MySQL server.
func WithTstMySQLServer(do func(context.Context) error) func(context.Context) error {

	return func(ctx context.Context) error {

		// Get dockertest pool.
		pool := DockertestPool()

		// Get arguments from enviroment.
		ver := os.Getenv("MYSQL_VERSION")
		if ver == "" {
			ver = defaultMySQLVersion
		}
		log.Printf("Using MySQL version %q\n", ver)

		// Start MySQL server container.
		log.Printf("Starting MySQL server, may take a while to pull docker image if not exists\n")
		resource, err := pool.Run("mysql", ver, []string{"MYSQL_ROOT_PASSWORD=123456"})
		if err != nil {
			return err
		}
		defer func() {
			pool.Purge(resource)
			log.Printf("Purged MySQL server container.\n")
		}()

		// Get dsn.
		dsn := fmt.Sprintf("root:123456@(localhost:%s)/mysql", resource.GetPort("3306/tcp"))

		// Wait server up.
		mysql.SetLogger(noopLogger{})
		log.Printf("Waiting for MySQL server...\n")
		if err := pool.Retry(func() error {

			connPool, err := sql.Open("mysql", dsn)
			if err != nil {
				return err
			}
			defer connPool.Close()
			return connPool.Ping()

		}); err != nil {
			return err
		}
		log.Printf("MySQL server up...")

		// Do.
		ctx2 := context.WithValue(ctx, dsnKey, dsn)
		return do(ctx2)

	}

}
