package testutils

import (
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

// StartMySQL starts a MySQL server container. It panics if any error.
func StartMySQL() (dsn string, teardown func()) {

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
		log.Panic(err)
	}
	defer func() {
		if r := recover(); r != nil {
			// Purge resource if panic.
			pool.Purge(resource)
			// Re-panic
			panic(r)
		}
	}()

	// Wait server up.
	mysql.SetLogger(noopLogger{})
	log.Printf("Waiting for MySQL server...\n")
	if err := pool.Retry(func() error {

		dsn = fmt.Sprintf("root:123456@(localhost:%s)/mysql", resource.GetPort("3306/tcp"))
		connPool, err := sql.Open("mysql", dsn)
		if err != nil {
			return err
		}
		defer connPool.Close()
		return connPool.Ping()

	}); err != nil {
		log.Panic(err)
	}
	log.Printf("MySQL server up...")

	// Done
	return dsn, func() { pool.Purge(resource) }

}
