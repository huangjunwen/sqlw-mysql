package infos

import (
	"strconv"
	"testing"

	"github.com/beevik/etree"
	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/stretchr/testify/assert"
)

// <nt>9</nt> -> <t>0</t><t>1</t>....<t>8</t>
type ntDirective int

// <t>10</t> -> "10"
type tDirective string

func init() {
	RegistDirectiveFactory(func() Directive {
		return new(ntDirective)
	}, "nt")
	RegistDirectiveFactory(func() Directive {
		return new(tDirective)
	}, "t")
}

func (d *ntDirective) Initialize(loader *datasrc.Loader, db *DBInfo, stmt *StmtInfo, tok etree.Token) error {
	i, err := strconv.Atoi(tok.(*etree.Element).Text())
	if err != nil {
		return err
	}
	*(*int)(d) = i
	return nil
}

func (d *ntDirective) Expand() ([]etree.Token, error) {
	ret := []etree.Token{}
	n := int(*d)
	for i := 0; i < n; i++ {
		elem := etree.NewElement("t")
		elem.SetText(strconv.Itoa(i))
		ret = append(ret, elem)
	}
	return ret, nil
}

func (d *tDirective) Initialize(loader *datasrc.Loader, db *DBInfo, stmt *StmtInfo, tok etree.Token) error {
	*(*string)(d) = tok.(*etree.Element).Text()
	return nil
}

func (d *tDirective) QueryFragment() (string, error) {
	return string(*d), nil
}

func (d *tDirective) TextFragment() (string, error) {
	return string(*d) + ".", nil
}

func (d *tDirective) ExtraProcess() error {
	return nil
}

func TestStmtInfo(t *testing.T) {

	assert := assert.New(t)

	exec("CREATE DATABASE stmtinfo")
	exec("USE stmtinfo")
	defer func() {
		exec("DROP DATABASE stmtinfo")
	}()

	db, err := NewDBInfo(loader)
	if err != nil {
		t.Fatal(err)
	}

	{
		// Not <stmt>
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`<notstmt></notstmt>`))
		_, err := NewStmtInfo(loader, db, doc.Root())
		assert.Error(err)
	}

	{
		// No `name` attribute.
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`<stmt></stmt>`))
		_, err := NewStmtInfo(loader, db, doc.Root())
		assert.Error(err)
	}

	{
		stmt := (*StmtInfo)(nil)
		assert.False(stmt.Valid())
		assert.Equal("", stmt.StmtName())
		assert.Equal("", stmt.UName())
		assert.Equal("", stmt.LName())
		assert.Equal("", stmt.Query())
		assert.Equal("", stmt.StmtType())
		assert.Equal("", stmt.Text())
		assert.Equal(0, stmt.NumQueryResultCol())
	}

	{
		// Simple select.
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`<stmt name="One">SELECT 1</stmt>`))
		stmt, err := NewStmtInfo(loader, db, doc.Root())
		assert.NoError(err)

		assert.True(stmt.Valid())
		assert.Equal("One", stmt.StmtName())
		assert.Equal("One", stmt.UName())
		assert.Equal("one", stmt.LName())
		assert.Equal("SELECT 1", stmt.Query())
		assert.Equal("SELECT", stmt.StmtType())
		assert.Equal("SELECT 1", stmt.Text())
		assert.Equal(1, stmt.NumQueryResultCol())
	}

	{
		// Directives.
		doc := etree.NewDocument()
		assert.NoError(doc.ReadFromString(`<stmt name="Directives">SELECT "<nt>9</nt>"</stmt>`))
		stmt, err := NewStmtInfo(loader, db, doc.Root())
		assert.NoError(err)

		assert.Equal("Directives", stmt.StmtName())
		assert.Equal("Directives", stmt.UName())
		assert.Equal("directives", stmt.LName())
		assert.Equal(`SELECT "012345678"`, stmt.Query())
		assert.Equal("SELECT", stmt.StmtType())
		assert.Equal(`SELECT "0.1.2.3.4.5.6.7.8."`, stmt.Text())
		assert.Equal(1, stmt.NumQueryResultCol())
	}
}
