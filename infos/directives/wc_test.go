package directives

import (
	"testing"

	"github.com/beevik/etree"
	"github.com/huangjunwen/sqlw-mysql/infos"
	"github.com/stretchr/testify/assert"
)

func TestWildcardsInfo(t *testing.T) {

	assert := assert.New(t)
	_ = assert

	exec("CREATE DATABASE test_wildcardsinfo")
	exec("USE test_wildcardsinfo")
	defer func() {
		exec("DROP DATABASE test_wildcardsinfo")
	}()

	exec(`
		CREATE TABLE user (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(32) NOT NULL,
			female BOOL,
			KEY ix_name (name)
		)
	`)

	exec(`
		CREATE TABLE employee (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id INT UNSIGNED NOT NULL,
			title VARCHAR(32) NOT NULL,
			superior_id INT UNSIGNED,
			KEY ix_title (title),
			FOREIGN KEY fk_user_id (user_id) REFERENCES user (id),
			FOREIGN KEY fk_superior_id (superior_id) REFERENCES employee (id)
		)
	`)

	db, err := infos.NewDBInfo(loader)
	assert.NoError(err)

	{
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`
			<stmt name="xxx">
				SELECT * FROM user
			</stmt>
		`))
		stmt, err := infos.NewStmtInfo(loader, db, doc.Root())
		assert.NoError(err)

		wildcards := ExtractWildcardsInfo(stmt)
		assert.False(wildcards.Valid())
	}

	{
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`
			<stmt name="xxx">
				SELECT <wc table="user" /> FROM user
			</stmt>
		`))
		stmt, err := infos.NewStmtInfo(loader, db, doc.Root())
		assert.NoError(err)
		wildcards := ExtractWildcardsInfo(stmt)

		assert.True(wildcards.Valid())
		assert.Len(wildcards.Wildcards(), 1)
		assert.Equal(0, wildcards.Wildcards()[0].Offset())
		assert.Equal([]int{0, 0, 0}, wildcards.resultCols2Wildcard)

	}
	{
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`
			<stmt name="xxx">
				SELECT NOW(), u.*, NOW(), u.*, NOW() FROM (SELECT <wc table="user" /> FROM user) u
			</stmt>
		`))
		stmt, err := infos.NewStmtInfo(loader, db, doc.Root())
		assert.NoError(err)
		wildcards := ExtractWildcardsInfo(stmt)

		assert.True(wildcards.Valid())
		assert.Len(wildcards.Wildcards(), 2)

		assert.Equal(1, wildcards.Wildcards()[0].Offset())
		assert.Equal("user", wildcards.Wildcards()[0].WildcardName())
		assert.Equal(5, wildcards.Wildcards()[1].Offset())
		assert.Equal("user", wildcards.Wildcards()[1].WildcardName())

		assert.Len(stmt.QueryResultCols(), 9)
		assert.False(wildcards.WildcardColumn(0).Valid())
		assert.Equal("id", wildcards.WildcardColumn(1).ColumnName())
		assert.Equal("name", wildcards.WildcardColumn(2).ColumnName())
		assert.Equal("female", wildcards.WildcardColumn(3).ColumnName())
		assert.False(wildcards.WildcardColumn(4).Valid())
		assert.Equal("id", wildcards.WildcardColumn(5).ColumnName())
		assert.Equal("name", wildcards.WildcardColumn(6).ColumnName())
		assert.Equal("female", wildcards.WildcardColumn(7).ColumnName())
		assert.False(wildcards.WildcardColumn(8).Valid())
		assert.Equal([]int{-1, 0, 0, 0, -1, 1, 1, 1, -1}, wildcards.resultCols2Wildcard)
	}

	{
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`
			<stmt name="xxx">
			SELECT
				<wc table="employee" as="e1" />, <wc table="user" as="u" />
			FROM
				employee e1
					LEFT JOIN employee e2 ON e1.superior_id = e2.id
					JOIN user u ON e1.user_id = u.id
			WHERE
				e2.id IS NULL
			</stmt>
		`))
		stmt, err := infos.NewStmtInfo(loader, db, doc.Root())
		assert.NoError(err)
		wildcards := ExtractWildcardsInfo(stmt)

		assert.True(wildcards.Valid())
		assert.Len(wildcards.Wildcards(), 2)

		assert.Equal(0, wildcards.Wildcards()[0].Offset())
		assert.Equal("e1", wildcards.Wildcards()[0].WildcardName())
		assert.Equal(4, wildcards.Wildcards()[1].Offset())
		assert.Equal("u", wildcards.Wildcards()[1].WildcardName())
		assert.Equal([]int{0, 0, 0, 0, 1, 1, 1}, wildcards.resultCols2Wildcard)
	}
	{
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`
			<stmt name="xxx">
			SELECT female FROM (SELECT <wc table="user" /> FROM user) u
			</stmt>
		`))
		stmt, err := infos.NewStmtInfo(loader, db, doc.Root())
		assert.NoError(err)
		wildcards := ExtractWildcardsInfo(stmt)

		assert.True(wildcards.Valid())
		assert.Len(wildcards.Wildcards(), 0)
		assert.Equal([]int{-1}, wildcards.resultCols2Wildcard)
	}
}
