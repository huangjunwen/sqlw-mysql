// Auto generated by sqlw-mysql (https://github.com/huangjunwen/sqlw-mysql) default template.
// DON NOT EDIT.

package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/volatiletech/null.v6"
)

// Tst1 is test TableRow with two-column primary key.
type Tst1 struct {
	c1 string
	p2 int32
	p1 int32
	a1 int32
}

// Tst2 is test TableRow without primary key.
type Tst2 struct {
	c1 string
	c2 int32
}

type tst1PrimaryValue struct {
	p1 int32
	p2 int32
}

var (
	_ TableRow            = (*Tst1)(nil)
	_ TableRowWithPrimary = (*Tst1)(nil)
	_ TableRow            = (*Tst2)(nil)
)

func (tr *Tst1) nxPreScan(dest *[]interface{}) {
	*dest = append(*dest, &tr.c1, &tr.p2, &tr.p1, &tr.a1)
}

func (tr *Tst1) nxPostScan() error {
	return nil
}

var (
	tst1Meta = NewTableMeta(
		"tst1",
		[]string{"c1", "p2", "p1", "a1"},
		OptPrimaryColumns("p1", "p2"),
		OptColumnsWithDefault("p1", "a1"),
		OptAutoIncColumn("a1"),
	)
)

func (tr *Tst1) TableMeta() *TableMeta {
	return tst1Meta
}

func (tr *Tst1) Valid() bool {
	return tr != nil
}

func (tr *Tst1) ColumnValue(i int) interface{} {
	switch i {
	case 0:
		return tr.c1
	case 1:
		return tr.p2
	case 2:
		return tr.p1
	case 3:
		return tr.a1
	default:
		panic(fmt.Errorf("%d is out of range", i))
	}
}

func (tr *Tst1) ColumnPointer(i int) interface{} {
	switch i {
	case 0:
		return &tr.c1
	case 1:
		return &tr.p2
	case 2:
		return &tr.p1
	case 3:
		return &tr.a1
	default:
		panic(fmt.Errorf("%d is out of range", i))
	}
}

func (tr *Tst1) PrimaryValue() interface{} {
	if tr == nil {
		return nil
	}
	return tst1PrimaryValue{
		p1: tr.p1,
		p2: tr.p2,
	}
}

func (tr *Tst2) nxPreScan(dest *[]interface{}) {
	*dest = append(*dest, &tr.c1, &tr.c2)
}

func (tr *Tst2) nxPostScan() error {
	return nil
}

var (
	tst2Meta = NewTableMeta(
		"tst2",
		[]string{"c1", "c2"},
	)
)

func (tr *Tst2) TableMeta() *TableMeta {
	return tst2Meta
}

func (tr *Tst2) Valid() bool {
	return tr != nil
}

func (tr *Tst2) ColumnValue(i int) interface{} {
	switch i {
	case 0:
		return tr.c1
	case 1:
		return tr.c2
	default:
		panic(fmt.Errorf("%d is out of range", i))
	}
}

func (tr *Tst2) ColumnPointer(i int) interface{} {
	switch i {
	case 0:
		return &tr.c1
	case 1:
		return &tr.c2
	default:
		panic(fmt.Errorf("%d is out of range", i))
	}
}

func TestTableMeta(t *testing.T) {

	assert := assert.New(t)

	{
		meta := (*Tst1)(nil).TableMeta()
		assert.Equal("tst1", meta.tableName)
		assert.Equal([]string{"c1", "p2", "p1", "a1"}, meta.columnNames)
		assert.Equal([]int{2, 1}, meta.primaryColumnsPos)
		assert.Equal(3, meta.autoIncColumnPos)
		assert.Equal([]bool{false, false, true, true}, meta.hasDefault)
		assert.Equal(map[string]int{"c1": 0, "p2": 1, "p1": 2, "a1": 3}, meta.columnsPos)
		assert.Equal([]bool{false, true, true, false}, meta.isPrimary)
		assert.Equal("`p1`=? AND `p2`=?", meta.primaryCond)
		assert.Equal("SELECT `c1`, `p2`, `p1`, `a1` FROM `tst1`", meta.selectQuery)
		assert.Equal("DELETE FROM `tst1` WHERE `p1`=? AND `p2`=?", meta.deleteByPrimaryQuery)
	}

	{
		meta := (*Tst2)(nil).TableMeta()
		assert.Equal("tst2", meta.tableName)
		assert.Equal([]string{"c1", "c2"}, meta.columnNames)
		assert.Len(meta.primaryColumnsPos, 0)
		assert.Equal(-1, meta.autoIncColumnPos)
		assert.Equal([]bool{false, false}, meta.hasDefault)
		assert.Equal(map[string]int{"c1": 0, "c2": 1}, meta.columnsPos)
		assert.Equal([]bool{false, false}, meta.isPrimary)
		assert.Equal("", meta.primaryCond)
		assert.Equal("", meta.selectQuery)
		assert.Equal("", meta.deleteByPrimaryQuery)
	}
}

func TestInsertTR(t *testing.T) {

	assert := assert.New(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// All zero values.
	{
		mock.ExpectExec("INSERT INTO `tst1` \\(`c1`, `p2`\\) VALUES \\(\\?, \\?\\)").
			WithArgs("", 0).
			WillReturnResult(sqlmock.NewResult(13, 1))
		tr := Tst1{}
		assert.NoError(insertTR(context.Background(), db, &tr, ""))
		assert.NoError(mock.ExpectationsWereMet())
		assert.Equal(Tst1{c1: "", p2: 0, p1: 0, a1: 13}, tr)
	}

	// Explict auto increment value.
	{
		mock.ExpectExec("INSERT INTO `tst1` \\(`c1`, `p2`, `p1`, `a1`\\) VALUES \\(\\?, \\?, \\?, \\?\\)").
			WithArgs("", 2, 1, 3).
			WillReturnResult(sqlmock.NewResult(13, 1))
		tr := Tst1{
			p1: 1,
			p2: 2,
			a1: 3,
		}
		assert.NoError(insertTR(context.Background(), db, &tr, ""))
		assert.NoError(mock.ExpectationsWereMet())
		assert.Equal(Tst1{c1: "", p2: 2, p1: 1, a1: 3}, tr)
	}

	// Insert ignore.
	{
		mock.ExpectExec("INSERT IGNORE INTO `tst1` \\(`c1`, `p2`\\) VALUES \\(\\?, \\?\\)").
			WithArgs("", 0).
			WillReturnResult(sqlmock.NewResult(13, 1))
		tr := Tst1{}
		assert.NoError(insertTR(context.Background(), db, &tr, "ignore"))
		assert.NoError(mock.ExpectationsWereMet())
		assert.Equal(Tst1{c1: "", p2: 0, p1: 0, a1: 13}, tr)
	}

	// Replace.
	{
		mock.ExpectExec("REPLACE INTO `tst1` \\(`c1`, `p2`\\) VALUES \\(\\?, \\?\\)").
			WithArgs("", 0).
			WillReturnResult(sqlmock.NewResult(13, 1))
		tr := Tst1{}
		assert.NoError(insertTR(context.Background(), db, &tr, "replace"))
		assert.NoError(mock.ExpectationsWereMet())
		assert.Equal(Tst1{c1: "", p2: 0, p1: 0, a1: 13}, tr)
	}
}

func TestUpdateTR(t *testing.T) {

	assert := assert.New(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// No columns to update.
	{
		tr := Tst1{}
		newTr := tr
		assert.NoError(updateTR(context.Background(), db, &tr, &newTr))
	}

	// One column to update.
	{
		tr := Tst1{
			p2: 2,
			p1: 1,
		}
		newTr := tr
		newTr.c1 = "abc"
		mock.ExpectExec("UPDATE `tst1` SET `c1`=\\? WHERE `p1`=\\? AND `p2`=\\?").
			WithArgs("abc", 1, 2).
			WillReturnResult(sqlmock.NewResult(0, 1))
		assert.NoError(updateTR(context.Background(), db, &tr, &newTr))
		assert.NoError(mock.ExpectationsWereMet())

	}

}

func TestDeleteTR(t *testing.T) {

	assert := assert.New(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	{
		tr := Tst1{
			p2: 2,
			p1: 1,
		}
		mock.ExpectExec("DELETE FROM `tst1` WHERE `p1`=\\? AND `p2`=\\?").
			WithArgs(1, 2).
			WillReturnResult(sqlmock.NewResult(0, 1))
		assert.NoError(deleteTR(context.Background(), db, &tr))
		assert.NoError(mock.ExpectationsWereMet())

	}
}

func TestSeleteTR(t *testing.T) {

	assert := assert.New(t)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Query by primary.
	{
		tr := Tst1{
			p2: 2,
			p1: 1,
		}
		mock.ExpectQuery("SELECT `c1`, `p2`, `p1`, `a1` FROM `tst1` WHERE `p1`=\\? AND `p2`=\\?").
			WithArgs(1, 2).
			WillReturnRows(sqlmock.NewRows([]string{"c1", "p2", "p1", "a1"}).AddRow("c1", 2, 1, 3))
		assert.NoError(selectTR(context.Background(), db, &tr, false))
		assert.Equal(Tst1{c1: "c1", p2: 2, p1: 1, a1: 3}, tr)
	}

	// Query by primary with lock.
	{
		tr := Tst1{
			p2: 2,
			p1: 1,
		}
		mock.ExpectQuery("SELECT `c1`, `p2`, `p1`, `a1` FROM `tst1` WHERE `p1`=\\? AND `p2`=\\? FOR UPDATE").
			WithArgs(1, 2).
			WillReturnRows(sqlmock.NewRows([]string{"c1", "p2", "p1", "a1"}).AddRow("c1", 2, 1, 3))
		assert.NoError(selectTRCond(context.Background(), db, &tr, true, ""))
		assert.Equal(Tst1{c1: "c1", p2: 2, p1: 1, a1: 3}, tr)
	}

	// Query by custom condition.
	{
		tr := Tst1{}
		mock.ExpectQuery("SELECT `c1`, `p2`, `p1`, `a1` FROM `tst1` WHERE `c1`=\\? AND `a1`!=\\?").
			WithArgs("c1", 3).
			WillReturnRows(sqlmock.NewRows([]string{"c1", "p2", "p1", "a1"}).AddRow("c1", 2, 1, 4))
		assert.NoError(selectTRCond(context.Background(), db, &tr, false, "`c1`=? AND `a1`!=?", "c1", 3))
		assert.Equal(Tst1{c1: "c1", p2: 2, p1: 1, a1: 4}, tr)
	}

}

func TestIsZero(t *testing.T) {

	assert := assert.New(t)

	for _, testCase := range []struct {
		Val    interface{}
		IsZero bool
	}{
		{float32(0), true},
		{float32(1), false},
		{float64(0), true},
		{float64(1), false},
		{false, true},
		{true, false},
		{int8(0), true},
		{int8(1), false},
		{int16(0), true},
		{int16(1), false},
		{int32(0), true},
		{int32(1), false},
		{int64(0), true},
		{int64(1), false},
		{uint8(0), true},
		{uint8(1), false},
		{uint16(0), true},
		{uint16(1), false},
		{uint32(0), true},
		{uint32(1), false},
		{uint64(0), true},
		{uint64(1), false},
		{time.Time{}, true},
		{time.Now(), false},
		{"", true},
		{"1", false},
		{null.Float32{}, true},
		{null.Float32From(1), false},
		{null.Float64{}, true},
		{null.Float64From(1), false},
		{null.Bool{}, true},
		{null.BoolFrom(false), false},
		{null.Int8{}, true},
		{null.Int8From(1), false},
		{null.Int16{}, true},
		{null.Int16From(1), false},
		{null.Int32{}, true},
		{null.Int32From(1), false},
		{null.Int64{}, true},
		{null.Int64From(1), false},
		{null.Uint8{}, true},
		{null.Uint8From(1), false},
		{null.Uint16{}, true},
		{null.Uint16From(1), false},
		{null.Uint32{}, true},
		{null.Uint32From(1), false},
		{null.Uint64{}, true},
		{null.Uint64From(1), false},
		{null.Time{}, true},
		{null.TimeFrom(time.Now()), false},
		{null.String{}, true},
		{null.StringFrom(""), false},
	} {
		assert.Equal(testCase.IsZero, isZero(testCase.Val))
	}
}