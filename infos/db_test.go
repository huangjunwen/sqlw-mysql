package infos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBInfo(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE dbinfo")
	exec("USE dbinfo")
	defer func() {
		exec("DROP DATABASE dbinfo")
	}()

	exec(`
		CREATE TABLE user (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			created_at TIMESTAMP DEFAULT NOW(),
			surname VARCHAR(16) NOT NULL,
			name VARCHAR(16) NOT NULL,
			female BOOL,
			KEY idx_name (name, surname)
		)
	`)

	exec(`
		CREATE TABLE employee (
			id INT UNSIGNED,
			superior_id INT UNSIGNED,
			user_id INT UNSIGNED NOT NULL,
			title VARCHAR(32) NOT NULL,
			UNIQUE KEY idx_id (id),
			UNIQUE KEY idx_user_id (user_id),
			KEY idx_title (title),
			CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES user (id),
			CONSTRAINT fk_superior_id FOREIGN KEY (superior_id) REFERENCES employee (id)
		)
	`)

	nilDBInfo := (*DBInfo)(nil)
	nilTableInfo := (*TableInfo)(nil)
	nilColumnInfo := (*ColumnInfo)(nil)
	nilIndexInfo := (*IndexInfo)(nil)
	nilFKInfo := (*FKInfo)(nil)
	{
		assert.False(nilDBInfo.Valid())
		assert.False(nilTableInfo.Valid())
		assert.False(nilColumnInfo.Valid())
		assert.False(nilIndexInfo.Valid())
		assert.False(nilFKInfo.Valid())
	}

	var db *DBInfo
	// NewDBInfo
	{
		var err error
		db, err = NewDBInfo(loader)
		assert.NoError(err)
		assert.True(db.Valid())

	}

	// DBInfo
	{
		assert.Equal(2, db.NumTable())

		assert.Nil(db.Table(-1))
		assert.NotNil(db.Table(0))
		assert.NotNil(db.Table(1))
		assert.Nil(db.Table(2))

		assert.Len(db.Tables(), 2)

		assert.NotNil(db.TableByName("user"))
		assert.NotNil(db.TableByName("employee"))
		assert.Nil(db.TableByName("notexists"))
	}

	user := db.TableByName("user")
	employee := db.TableByName("employee")

	// TableInfo
	{

		assert.True(user.Valid())

		assert.Equal("User", user.UName())
		assert.Equal("", nilTableInfo.UName())
		assert.Equal("user", user.LName())
		assert.Equal("", nilTableInfo.LName())

		assert.Equal("user", user.TableName())
		assert.Equal("", nilTableInfo.TableName())

		assert.Equal(0, nilTableInfo.NumColumn())
		assert.Equal(5, user.NumColumn())

		assert.Nil(nilTableInfo.Column(0))
		assert.Nil(user.Column(-1))
		assert.NotNil(user.Column(0))
		assert.NotNil(user.Column(4))
		assert.Nil(user.Column(5))

		assert.Len(nilTableInfo.Columns(), 0)
		assert.Len(user.Columns(), 5)

		assert.Nil(nilTableInfo.ColumnByName("xxx"))
		assert.Nil(user.ColumnByName("notexists"))
		assert.NotNil(user.ColumnByName("female"))

		assert.Equal(0, nilTableInfo.NumIndex())
		assert.Equal(2, user.NumIndex())

		assert.Nil(nilTableInfo.Index(0))
		assert.Nil(user.Index(-1))
		assert.NotNil(user.Index(0))
		assert.NotNil(user.Index(1))
		assert.Nil(user.Index(2))

		assert.Len(nilTableInfo.Indices(), 0)
		assert.Len(user.Indices(), 2)
		assert.Len(employee.Indices(), 4)

		assert.Nil(nilTableInfo.IndexByName("xxx"))
		assert.Nil(user.IndexByName("notexists"))
		assert.NotNil(user.IndexByName("PRIMARY"))
		assert.NotNil(user.IndexByName("idx_name"))
		assert.NotNil(employee.IndexByName("idx_id"))
		assert.NotNil(employee.IndexByName("idx_user_id"))
		assert.NotNil(employee.IndexByName("idx_title"))
		// NOTE: fk_user_id does not create a new index since idx_user_id
		assert.NotNil(employee.IndexByName("fk_superior_id"))

		assert.Equal(nilTableInfo.NumFK(), 0)
		assert.Equal(user.NumFK(), 0)
		assert.Equal(employee.NumFK(), 2)

		assert.Nil(nilTableInfo.FK(0))
		assert.Nil(employee.FK(-1))
		assert.NotNil(employee.FK(0))
		assert.NotNil(employee.FK(1))
		assert.Nil(employee.FK(2))

		assert.Len(nilTableInfo.FKs(), 0)
		assert.Len(employee.FKs(), 2)

		assert.Nil(nilTableInfo.FKByName("xxx"))
		assert.Nil(employee.FKByName("notexists"))
		assert.NotNil(employee.FKByName("fk_user_id"))
		assert.NotNil(employee.FKByName("fk_superior_id"))

		assert.Nil(nilTableInfo.Primary())
		assert.NotNil(user.Primary())
		assert.Nil(employee.Primary())

		assert.Nil(nilTableInfo.AutoIncColumn())
		assert.NotNil(user.AutoIncColumn())
		assert.Nil(employee.AutoIncColumn())
	}

	// ColumnInfo
	{
		{
			col := nilColumnInfo
			assert.Equal("", col.UName())
			assert.Equal("", col.LName())
			assert.Nil(col.Table())
			assert.Equal("", col.ColumnName())
			assert.Equal("", col.DataType())
			assert.True(col.Nullable())
			assert.Equal(-1, col.Pos())
			assert.False(col.HasDefaultValue())
			assert.Nil(col.Col())
		}

		{
			col := user.ColumnByName("id")
			assert.Equal("Id", col.UName())
			assert.Equal("id", col.LName())
			assert.Equal(user, col.Table())
			assert.Equal("id", col.ColumnName())
			assert.Equal("uint32", col.DataType())
			assert.False(col.Nullable())
			assert.Equal(0, col.Pos())
			assert.True(col.HasDefaultValue())
			assert.NotNil(col.Col())
		}
		{
			col := user.ColumnByName("female")
			assert.Equal("Female", col.UName())
			assert.Equal("female", col.LName())
			assert.Equal(user, col.Table())
			assert.Equal("female", col.ColumnName())
			assert.Equal("bool", col.DataType())
			assert.True(col.Nullable())
			assert.Equal(4, col.Pos())
			assert.False(col.HasDefaultValue())
			assert.NotNil(col.Col())
		}

	}

	// IndexInfo
	{
		{
			idx := nilIndexInfo
			assert.Equal("", idx.UName())
			assert.Equal("", idx.LName())
			assert.Equal("", idx.IndexName())
			assert.Nil(idx.Table())
			assert.Len(idx.Columns(), 0)
			assert.False(idx.IsPrimary())
			assert.False(idx.IsUnique())
		}
		{
			idx := user.IndexByName("idx_name")
			assert.Equal("IdxName", idx.UName())
			assert.Equal("idxName", idx.LName())
			assert.Equal("idx_name", idx.IndexName())
			assert.NotNil(idx.Table())
			assert.Len(idx.Columns(), 2)
			assert.Equal([]*ColumnInfo{user.ColumnByName("name"), user.ColumnByName("surname")}, idx.Columns())
			assert.False(idx.IsPrimary())
			assert.False(idx.IsUnique())
		}

		{
			idx := employee.IndexByName("idx_user_id")
			assert.Equal("IdxUserId", idx.UName())
			assert.Equal("idxUserId", idx.LName())
			assert.Equal("idx_user_id", idx.IndexName())
			assert.NotNil(idx.Table())
			assert.Len(idx.Columns(), 1)
			assert.Equal([]*ColumnInfo{employee.ColumnByName("user_id")}, idx.Columns())
			assert.False(idx.IsPrimary())
			assert.True(idx.IsUnique())
		}

	}

	// FKInfo
	{
		{
			fk := nilFKInfo
			assert.Equal("", fk.UName())
			assert.Equal("", fk.LName())
			assert.Equal("", fk.FKName())
			assert.Nil(fk.Table())
			assert.Len(fk.Columns(), 0)
			assert.Nil(fk.RefTable())
			assert.Len(fk.RefColumns(), 0)
			assert.Nil(fk.RefUniqueIndex())
		}

		{
			fk := employee.FKByName("fk_superior_id")
			assert.Equal("FkSuperiorId", fk.UName())
			assert.Equal("fkSuperiorId", fk.LName())
			assert.Equal("fk_superior_id", fk.FKName())
			assert.NotNil(fk.Table())
			assert.Len(fk.Columns(), 1)
			assert.Equal([]*ColumnInfo{employee.ColumnByName("superior_id")}, fk.Columns())
			assert.Equal(employee, fk.RefTable())
			assert.Len(fk.RefColumns(), 1)
			assert.Equal([]*ColumnInfo{employee.ColumnByName("id")}, fk.RefColumns())
			assert.NotNil(fk.RefUniqueIndex())
			assert.Equal(employee.IndexByName("idx_id"), fk.RefUniqueIndex())
		}

	}

}
