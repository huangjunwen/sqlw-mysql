// Auto generated by sqlw-mysql (https://github.com/huangjunwen/sqlw-mysql) default template.
// DON NOT EDIT.

package models

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"text/template"
	"time"

	"github.com/jmoiron/sqlx"
	null "gopkg.in/volatiletech/null.v6"
)

// AllUserEmployeesResult is the result of `AllUserEmployees`.
type AllUserEmployeesResult struct {
	User       *User
	Age        null.Uint64
	Empl       *Employee
	nxNullUser nxNullUser
	nxNullEmpl nxNullEmployee
}

func (result *AllUserEmployeesResult) nxPreScan(dest *[]interface{}) {
	result.nxNullUser.nxPreScan(dest)
	*dest = append(*dest, &result.Age)
	result.nxNullEmpl.nxPreScan(dest)
}

func (result *AllUserEmployeesResult) nxPostScan() error {

	if err := result.nxNullUser.nxPostScan(); err != nil {
		return err
	}
	result.User = result.nxNullUser.Ordinary()

	if err := result.nxNullEmpl.nxPostScan(); err != nil {
		return err
	}
	result.Empl = result.nxNullEmpl.Ordinary()

	return nil
}

// AllUserEmployeesResultSlice is slice of AllUserEmployeesResult.
type AllUserEmployeesResultSlice []*AllUserEmployeesResult

// nxAllUserEmployeesResultSlices is slice of AllUserEmployeesResultSlice.
type nxAllUserEmployeesResultSlices []AllUserEmployeesResultSlice

func (slice *AllUserEmployeesResultSlice) nxLen() int {
	return len(*slice)
}

func (slice *AllUserEmployeesResultSlice) nxItem(i int) interface{} {
	return (*slice)[i]
}

func (slice *AllUserEmployeesResultSlice) nxAppend(item interface{}) {
	if item == nil {
		*slice = append(*slice, &AllUserEmployeesResult{})
	} else {
		*slice = append(*slice, item.(*AllUserEmployeesResult))
	}
}

// One returns a single AllUserEmployeesResult. It panics if the length of slice is not 1.
func (slice *AllUserEmployeesResultSlice) One() *AllUserEmployeesResult {
	if len(*slice) != 1 {
		panic(fmt.Errorf("AllUserEmployeesResultSlice.One is called but has %d rows", len(*slice)))
	}
	return (*slice)[0]
}

// DistinctUser returns distinct (by primary value) User in the slice.
func (slice *AllUserEmployeesResultSlice) DistinctUser() []*User {
	trs := UserSlice{}
	groupBy(slice, func(item interface{}) TableRowWithPrimary {
		return item.(*AllUserEmployeesResult).User
	}, &trs, nil)
	return trs
}

// GroupByUser groups by User and returns distinct (by primary value) User with
// their associated sub group of slices.
func (slice *AllUserEmployeesResultSlice) GroupByUser() ([]*User, []AllUserEmployeesResultSlice) {
	trs := UserSlice{}
	groups := nxAllUserEmployeesResultSlices{}
	groupBy(slice, func(item interface{}) TableRowWithPrimary {
		return item.(*AllUserEmployeesResult).User
	}, &trs, &groups)
	return trs, groups
}

// DistinctEmpl returns distinct (by primary value) Empl in the slice.
func (slice *AllUserEmployeesResultSlice) DistinctEmpl() []*Employee {
	trs := EmployeeSlice{}
	groupBy(slice, func(item interface{}) TableRowWithPrimary {
		return item.(*AllUserEmployeesResult).Empl
	}, &trs, nil)
	return trs
}

// GroupByEmpl groups by Empl and returns distinct (by primary value) Empl with
// their associated sub group of slices.
func (slice *AllUserEmployeesResultSlice) GroupByEmpl() ([]*Employee, []AllUserEmployeesResultSlice) {
	trs := EmployeeSlice{}
	groups := nxAllUserEmployeesResultSlices{}
	groupBy(slice, func(item interface{}) TableRowWithPrimary {
		return item.(*AllUserEmployeesResult).Empl
	}, &trs, &groups)
	return trs, groups
}

func (slices *nxAllUserEmployeesResultSlices) nxLen() int {
	return len(*slices)
}

func (slices *nxAllUserEmployeesResultSlices) nxItem(i int) interface{} {
	return &((*slices)[i])
}

func (slices *nxAllUserEmployeesResultSlices) nxAppend(item interface{}) {
	if item == nil {
		*slices = append(*slices, AllUserEmployeesResultSlice{})
	} else {
		*slices = append(*slices, *(item.(*AllUserEmployeesResultSlice)))
	}
}

/*
AllUserEmployees is created from:

  <stmt name="AllUserEmployees">
    SELECT
      <wc table="user"/>,
      CAST(DATEDIFF(NOW(), birthday)/365 AS UNSIGNED) AS age,
      <wc table="employee" as="empl"/>
    FROM
      user LEFT JOIN employee AS empl ON user.id=empl.user_id
  </stmt>
*/
func AllUserEmployees(ctx context.Context, q Queryer) (AllUserEmployeesResultSlice, error) {
	// NOTE: Add a nested block to allow identifier shadowing.
	{
		// Build query.
		query, args, err := nxBuildAllUserEmployeesQuery(map[string]interface{}{})
		if err != nil {
			return nil, err
		}

		// Query.
		rows, err := q.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Scan.
		results := []*AllUserEmployeesResult{}
		pointers := make([]interface{}, 0, 9)
		for rows.Next() {
			// Fill scan pointers.
			pointers = pointers[0:0]
			result := &AllUserEmployeesResult{}
			result.nxPreScan(&pointers)

			// Scan.
			err = rows.Scan(pointers...)
			if err != nil {
				return nil, err
			}

			// Post scan process.
			if err := result.nxPostScan(); err != nil {
				return nil, err
			}
			results = append(results, result)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}
		return results, nil

	}

}

var (
	nxAllUserEmployeesQuery = "SELECT\n    `user`.`id`, `user`.`name`, `user`.`female`, `user`.`birthday`,\n    CAST(DATEDIFF(NOW(), birthday)/365 AS UNSIGNED) AS age,\n    `empl`.`id`, `empl`.`employee_sn`, `empl`.`user_id`, `empl`.`superior_id`\n  FROM\n    user LEFT JOIN employee AS empl ON user.id=empl.user_id"
)

func nxBuildAllUserEmployeesQuery(data map[string]interface{}) (string, []interface{}, error) {

	// Named query.
	namedQuery := nxAllUserEmployeesQuery

	// Named query -> query.
	query, args, err := sqlx.Named(namedQuery, data)
	if err != nil {
		return "", nil, err
	}

	return query, args, nil
}

// SubordinatesBySuperiorsResult is the result of `SubordinatesBySuperiors`.
type SubordinatesBySuperiorsResult struct {
	Superior          *Employee
	Subordinate       *Employee
	nxNullSuperior    nxNullEmployee
	nxNullSubordinate nxNullEmployee
}

func (result *SubordinatesBySuperiorsResult) nxPreScan(dest *[]interface{}) {
	result.nxNullSuperior.nxPreScan(dest)
	result.nxNullSubordinate.nxPreScan(dest)
}

func (result *SubordinatesBySuperiorsResult) nxPostScan() error {

	if err := result.nxNullSuperior.nxPostScan(); err != nil {
		return err
	}
	result.Superior = result.nxNullSuperior.Ordinary()

	if err := result.nxNullSubordinate.nxPostScan(); err != nil {
		return err
	}
	result.Subordinate = result.nxNullSubordinate.Ordinary()

	return nil
}

// SubordinatesBySuperiorsResultSlice is slice of SubordinatesBySuperiorsResult.
type SubordinatesBySuperiorsResultSlice []*SubordinatesBySuperiorsResult

// nxSubordinatesBySuperiorsResultSlices is slice of SubordinatesBySuperiorsResultSlice.
type nxSubordinatesBySuperiorsResultSlices []SubordinatesBySuperiorsResultSlice

func (slice *SubordinatesBySuperiorsResultSlice) nxLen() int {
	return len(*slice)
}

func (slice *SubordinatesBySuperiorsResultSlice) nxItem(i int) interface{} {
	return (*slice)[i]
}

func (slice *SubordinatesBySuperiorsResultSlice) nxAppend(item interface{}) {
	if item == nil {
		*slice = append(*slice, &SubordinatesBySuperiorsResult{})
	} else {
		*slice = append(*slice, item.(*SubordinatesBySuperiorsResult))
	}
}

// One returns a single SubordinatesBySuperiorsResult. It panics if the length of slice is not 1.
func (slice *SubordinatesBySuperiorsResultSlice) One() *SubordinatesBySuperiorsResult {
	if len(*slice) != 1 {
		panic(fmt.Errorf("SubordinatesBySuperiorsResultSlice.One is called but has %d rows", len(*slice)))
	}
	return (*slice)[0]
}

// DistinctSuperior returns distinct (by primary value) Superior in the slice.
func (slice *SubordinatesBySuperiorsResultSlice) DistinctSuperior() []*Employee {
	trs := EmployeeSlice{}
	groupBy(slice, func(item interface{}) TableRowWithPrimary {
		return item.(*SubordinatesBySuperiorsResult).Superior
	}, &trs, nil)
	return trs
}

// GroupBySuperior groups by Superior and returns distinct (by primary value) Superior with
// their associated sub group of slices.
func (slice *SubordinatesBySuperiorsResultSlice) GroupBySuperior() ([]*Employee, []SubordinatesBySuperiorsResultSlice) {
	trs := EmployeeSlice{}
	groups := nxSubordinatesBySuperiorsResultSlices{}
	groupBy(slice, func(item interface{}) TableRowWithPrimary {
		return item.(*SubordinatesBySuperiorsResult).Superior
	}, &trs, &groups)
	return trs, groups
}

// DistinctSubordinate returns distinct (by primary value) Subordinate in the slice.
func (slice *SubordinatesBySuperiorsResultSlice) DistinctSubordinate() []*Employee {
	trs := EmployeeSlice{}
	groupBy(slice, func(item interface{}) TableRowWithPrimary {
		return item.(*SubordinatesBySuperiorsResult).Subordinate
	}, &trs, nil)
	return trs
}

// GroupBySubordinate groups by Subordinate and returns distinct (by primary value) Subordinate with
// their associated sub group of slices.
func (slice *SubordinatesBySuperiorsResultSlice) GroupBySubordinate() ([]*Employee, []SubordinatesBySuperiorsResultSlice) {
	trs := EmployeeSlice{}
	groups := nxSubordinatesBySuperiorsResultSlices{}
	groupBy(slice, func(item interface{}) TableRowWithPrimary {
		return item.(*SubordinatesBySuperiorsResult).Subordinate
	}, &trs, &groups)
	return trs, groups
}

func (slices *nxSubordinatesBySuperiorsResultSlices) nxLen() int {
	return len(*slices)
}

func (slices *nxSubordinatesBySuperiorsResultSlices) nxItem(i int) interface{} {
	return &((*slices)[i])
}

func (slices *nxSubordinatesBySuperiorsResultSlices) nxAppend(item interface{}) {
	if item == nil {
		*slices = append(*slices, SubordinatesBySuperiorsResultSlice{})
	} else {
		*slices = append(*slices, *(item.(*SubordinatesBySuperiorsResultSlice)))
	}
}

/*
SubordinatesBySuperiors is created from:

  <stmt name="SubordinatesBySuperiors">
    <a name="id" type="...int"/>
    <v in_query="1"/>
    SELECT
      <wc table="employee" as="superior"/>,
      <wc table="employee" as="subordinate"/>
    FROM
      employee AS superior LEFT JOIN employee AS subordinate ON subordinate.superior_id=superior.id
    WHERE
      superior.id IN (<b name="id"/>)
  </stmt>
*/
func SubordinatesBySuperiors(ctx context.Context, q Queryer, id ...int) (SubordinatesBySuperiorsResultSlice, error) {
	// NOTE: Add a nested block to allow identifier shadowing.
	{
		// Build query.
		query, args, err := nxBuildSubordinatesBySuperiorsQuery(map[string]interface{}{
			"id": id,
		})
		if err != nil {
			return nil, err
		}

		// Query.
		rows, err := q.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Scan.
		results := []*SubordinatesBySuperiorsResult{}
		pointers := make([]interface{}, 0, 8)
		for rows.Next() {
			// Fill scan pointers.
			pointers = pointers[0:0]
			result := &SubordinatesBySuperiorsResult{}
			result.nxPreScan(&pointers)

			// Scan.
			err = rows.Scan(pointers...)
			if err != nil {
				return nil, err
			}

			// Post scan process.
			if err := result.nxPostScan(); err != nil {
				return nil, err
			}
			results = append(results, result)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}
		return results, nil

	}

}

var (
	nxSubordinatesBySuperiorsQuery = "SELECT\n    `superior`.`id`, `superior`.`employee_sn`, `superior`.`user_id`, `superior`.`superior_id`,\n    `subordinate`.`id`, `subordinate`.`employee_sn`, `subordinate`.`user_id`, `subordinate`.`superior_id`\n  FROM\n    employee AS superior LEFT JOIN employee AS subordinate ON subordinate.superior_id=superior.id\n  WHERE\n    superior.id IN (:id)"
)

func nxBuildSubordinatesBySuperiorsQuery(data map[string]interface{}) (string, []interface{}, error) {

	// Named query.
	namedQuery := nxSubordinatesBySuperiorsQuery

	// Named query -> query.
	query, args, err := sqlx.Named(namedQuery, data)
	if err != nil {
		return "", nil, err
	}

	// Expand "in" args.
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return "", nil, err
	}

	return query, args, nil
}

/*
UsersByCond is created from:

  <stmt name="UsersByCond">
    <v use_template="1"/>
    <a name="id" type="int"/>
    <a name="name" type="string"/>
    <a name="birthday" type="time.Time"/>
    <a name="limit" type="int"/>
    SELECT
      <wc table="user"/>
    FROM
      user
    WHERE
      <t>{{ if ne .id 0 }}</t>
        id=<b name="id"/> AND
      <t>{{ end }}</t>

      <t>{{ if ne (len .name) 0 }}</t>
        name=<b name="name"/> AND
      <t>{{ end }}</t>

      <t>{{ if not .birthday.IsZero }}</t>
        birthday=<b name="birthday"/> AND
      <t>{{ end }}</t>
      1
    LIMIT <b name="limit">10</b>
  </stmt>
*/
func UsersByCond(ctx context.Context, q Queryer, id int, name string, birthday time.Time, limit int) (UserSlice, error) {
	// NOTE: Add a nested block to allow identifier shadowing.
	{
		// Build query.
		query, args, err := nxBuildUsersByCondQuery(map[string]interface{}{
			"id":       id,
			"name":     name,
			"birthday": birthday,
			"limit":    limit,
		})
		if err != nil {
			return nil, err
		}

		// Query.
		rows, err := q.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Scan.
		results := []*User{}
		pointers := make([]interface{}, 0, 4)
		for rows.Next() {
			// Fill scan pointers.
			pointers = pointers[0:0]
			result := &User{}
			result.nxPreScan(&pointers)

			// Scan.
			err = rows.Scan(pointers...)
			if err != nil {
				return nil, err
			}

			// Post scan process.
			if err := result.nxPostScan(); err != nil {
				return nil, err
			}
			results = append(results, result)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}
		return results, nil

	}

}

var (
	nxUsersByCondQueryTmpl = template.Must(template.New("UsersByCond").Parse("SELECT\n    `user`.`id`, `user`.`name`, `user`.`female`, `user`.`birthday`\n  FROM\n    user\n  WHERE\n    {{ if ne .id 0 }}\n      id=:id AND\n    {{ end }}\n\n    {{ if ne (len .name) 0 }}\n      name=:name AND\n    {{ end }}\n\n    {{ if not .birthday.IsZero }}\n      birthday=:birthday AND\n    {{ end }}\n    1\n  LIMIT :limit"))
)

func nxBuildUsersByCondQuery(data map[string]interface{}) (string, []interface{}, error) {

	// Template -> named query.
	buf := bytes.Buffer{}
	if err := nxUsersByCondQueryTmpl.Execute(&buf, data); err != nil {
		return "", nil, err
	}
	namedQuery := buf.String()

	// Named query -> query.
	query, args, err := sqlx.Named(namedQuery, data)
	if err != nil {
		return "", nil, err
	}

	return query, args, nil
}

var (
	// Suppress "imported and not used" errors.
	_ = fmt.Printf
	_ = context.Background
	_ = template.IsTrue
	_ = sql.Open
	_ = sqlx.Named
	_ = null.NewBool
	_ = bytes.NewBuffer
	_ = time.Now
)
