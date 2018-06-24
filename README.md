# SQL wrapper code generator for MySQL [![Go Report Card](https://goreportcard.com/badge/github.com/huangjunwen/sqlw-mysql)](https://goreportcard.com/report/github.com/huangjunwen/sqlw-mysql)

`sqlw-mysql` is a CLI tool to generate go wrapper code for your MySQL table schemas and queries.

## Table of Contents

- [Install](#install)
- [Quickstart](#quickstart)

<a name="install" />

## Install

``` bash
$ go get -u github.com/huangjunwen/sqlw-mysql
```

<a name="quickstart" />

## Quickstart

Let's start with a small example. Suppose you have a database with these tables:

``` sql
-- Database: `db`

CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `female` tinyint(1) DEFAULT NULL,
  `birthday` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE `employee` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `employee_sn` char(32) NOT NULL,
  `user_id` int(11) NOT NULL,
  `superior_id` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `employee_sn` (`employee_sn`),
  UNIQUE KEY `user_id` (`user_id`),
  KEY `fk_superior` (`superior_id`),
  CONSTRAINT `fk_superior` FOREIGN KEY (`superior_id`) REFERENCES `employee` (`id`),
  CONSTRAINT `fk_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
);

```

_Some explanation: An `employee` must be a `user`, but a `user` need not to be an `employee`; Each `employee` must have a superior except those top dogs._

Now run `sqlw-mysql` against the database:

``` bash
$ sqlw-mysql -dsn "user:passwd@tcp(host:port)/db?parseTime=true"
$ ls ./models
... table_user.go ... table_employee.go

```

You will see a `models` directory is created with several source files generated:

Especially, a `table_<table name>.go` is generated for each table containing structs/methods for some common single table operations:

``` go
// ./models/table_user.go

// User represents a row of table `user`.
type User struct {
  Id       int32     `json:"id" db:"id"`
  Name     string    `json:"name" db:"name"`
  Female   null.Bool `json:"female" db:"female"`
  Birthday null.Time `json:"birthday" db:"birthday"`
}

// ...
```

But eventually, you will need more complex quries. For example you want to query all `user` and its associated `employee` (e.g. `one2one` relationship). Then you can write a statement XML like this:

``` xml
<!-- ./stmts/user.xml -->

<stmt name="AllUserEmployeeInfo">
  SELECT 
    <wc table="user" />,
    CAST(DATEDIFF(NOW(), birthday)/365 AS UNSIGNED) AS age,
    <wc table="employee" as="empl" />     
  FROM
    user LEFT JOIN employee AS empl ON user.id=empl.user_id
</stmt>

```

A statement XML contains SQL statements with special directives embeded in. Here you can see two `<wc table="table_name">` directives, which are roughly equal to expanded `table_name.*`.

Now run `sqlw-mysql` again specifying statement XML directory:

``` bash
$ sqlw-mysql -dsn "user:passwd@tcp(host:port)/db?parseTime=true" -stmt ./stmts
$ ls ./models
... table_user.go ... table_employee.go ... stmt_user.go

```

A new file `stmt_user.go` is generated from `user.xml`:

``` go
// ./models/stmt_user.go

// AllUserEmployeeInfoResult is the result of AllUserEmployeeInfo.
type AllUserEmployeeInfoResult struct {
  User       *User
  Age        null.Uint64
  Empl       *Employee
  nxNullUser nxNullUser
  nxNullEmpl nxNullEmployee
}

// AllUserEmployeeInfoResultSlice is slice of AllUserEmployeeInfoResult.
type AllUserEmployeeInfoResultSlice []*AllUserEmployeeInfoResult

// ...

/*
AllUserEmployeeInfo is created from:

  <stmt name="AllUserEmployeeInfo">
    SELECT
      <wc table="user"/>,
      CAST(DATEDIFF(NOW(), birthday)/365 AS UNSIGNED) AS age,
      <wc table="employee" as="empl"/>
    FROM
      user LEFT JOIN employee AS empl ON user.id=empl.user_id
  </stmt>
*/
func AllUserEmployeeInfo(ctx context.Context, q Queryer) (AllUserEmployeeInfoResultSlice, error) {
  // ...
}

```

Notice that `User` and `Empl` fields in result struct are generated from those `<wc>` directives. `sqlw-mysql` is smart enough to figure out their correct positions.

Now you can use the newly created function to iterate through all `user` and `employee`:

``` go
slice, err := AllUserEmployeeInfo(ctx, db)
if err != nil {
  panic(err)
}

for _, result := range slice {
  user := result.User
  empl := result.Empl

  if !empl.Valid() {
    // The user is not an employee.
    // ...
  } else {
    // The user is an employee.
    // ...
  }

  // ...
}

```

Another example, if you want to find subordinates of some employees (e.g. `one2many` relationship):

``` xml
<!-- ./stmts/user.xml -->

<stmt name="EmployeeInfo">
  <arg name="id" type="...int" />
  <vars in_query="1" />
  SELECT 
    <wc table="employee" as="superior" />,
    <wc table="employee" as="subordinate" />
  FROM
    employee AS superior LEFT JOIN employee AS subordinate ON subordinate.superior_id=superior.id
  WHERE
    superior.id IN (<repl with=":id">1</repl>)
</stmt>

```

Brief explanation about new directives:
- `<arg>` specifies an argument of the generated function.
- `<vars>` specifies arbitary variables that the template can use. `in_query="1"` tells the template that the function use `IN` operator.
- `<repl>` can replace arbitary statement text.

After re-run the command, the following code is generated:

``` go
// ./models/stmt_user.go

// EmployeeInfoResult is the result of EmployeeInfo.
type EmployeeInfoResult struct {
  Superior          *Employee
  Subordinate       *Employee
  nxNullSuperior    nxNullEmployee
  nxNullSubordinate nxNullEmployee
}

// EmployeeInfoResultSlice is slice of EmployeeInfoResult.
type EmployeeInfoResultSlice []*EmployeeInfoResult

// ...

/*
EmployeeInfo is created from:

  <stmt name="EmployeeInfo">
    <arg name="id" type="...int"/>
    <vars in_query="1"/>
    SELECT
      <wc table="employee" as="superior"/>,
      <wc table="employee" as="subordinate"/>
    FROM
      employee AS superior LEFT JOIN employee AS subordinate ON subordinate.superior_id=superior.id
    WHERE
      superior.id IN (<repl with=":id">1</repl>)
    </stmt>
*/
func EmployeeInfo(ctx context.Context, q Queryer, id ...int) (EmployeeInfoResultSlice, error) {
  // ...
}
```

Then, you can iterate the result like:

``` go
slice, err := EmployeeInfo(ctx, db)
if err != nil {
  panic(err)
}

// Group result slice by distinct superior.
superiors, groups := slice.GroupBySuperior()
for i, superior := range superiors {
  // All rows in groups[i] have the same superior.
  subordinates := groups[i].DistinctSubordinate()

  // Process with superior/subordinates.
  // ...
}
```

TODO ...
