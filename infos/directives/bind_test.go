package directives

import (
	"strings"
	"testing"

	"github.com/huangjunwen/sqlw-mysql/infos"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
)

func TestBindDirective(t *testing.T) {

	assert := assert.New(t)
	_ = assert

	exec("CREATE DATABASE test_binddirective")
	exec("USE test_binddirective")
	defer func() {
		exec("DROP DATABASE test_binddirective")
	}()

	exec(`
		CREATE TABLE user (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(32) NOT NULL,
			female BOOL,
			KEY ix_name (name)
		)
	`)

	db, err := infos.NewDBInfo(loader)
	assert.NoError(err)

	{
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`
			<stmt name="xxx">
				SELECT * FROM user WHERE id=<b name="notexists"/>
			</stmt>
		`))
		_, err := infos.NewStmtInfo(loader, db, doc.Root())
		assert.Error(err)

	}

	{
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`
			<stmt name="xxx">
				<a name="x" type="t" />
				SELECT * FROM user WHERE id=<b/>
			</stmt>
		`))
		_, err := infos.NewStmtInfo(loader, db, doc.Root())
		assert.Error(err)

	}

	{
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`
			<stmt name="xxx">
				<a name="x" type="t" />
				SELECT * FROM user WHERE id=<b name="y"/>
			</stmt>
		`))
		_, err := infos.NewStmtInfo(loader, db, doc.Root())
		assert.Error(err)

	}

	{
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`
			<stmt name="xxx">
				<a name="x" type="t" />
				SELECT * FROM user WHERE id=<b name="x"/>
			</stmt>
		`))
		stmt, err := infos.NewStmtInfo(loader, db, doc.Root())
		assert.NoError(err)
		assert.Equal("SELECT * FROM user WHERE id=NULL", strings.TrimSpace(stmt.Query()))
		assert.Equal("SELECT * FROM user WHERE id=:x", strings.TrimSpace(stmt.Text()))

	}

	{
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`
			<stmt name="xxx">
				<a name="limit" type="int" />
				SELECT * FROM user LIMIT <b name="limit">10</b>
			</stmt>
		`))
		stmt, err := infos.NewStmtInfo(loader, db, doc.Root())
		assert.NoError(err)
		assert.Equal("SELECT * FROM user LIMIT 10", strings.TrimSpace(stmt.Query()))
		assert.Equal("SELECT * FROM user LIMIT :limit", strings.TrimSpace(stmt.Text()))

	}
}
