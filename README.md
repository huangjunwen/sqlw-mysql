# SQL wrapper code generator for MySQL [![Go Report Card](https://goreportcard.com/badge/github.com/huangjunwen/sqlw-mysql)](https://goreportcard.com/report/github.com/huangjunwen/sqlw-mysql)

`sqlw-mysql` is a CLI tool to generate go wrapper code for your MySQL database and queries.

## Table of Contents

- [Install](#install)
- [Design/Goals/Features](#designgoalsfeatures)
- [Quickstart](#quickstart)
- [Statement XML](#statement-xml)
  - [Directives](#directives)
    - [How wildcard directive works](#how-wildcard-directive-works)
- [Code template](#code-template)
  - [Default template](#default-template)
- [Command line options](#command-line-options)
- [Motivation](#motivation)
- [Licence](#licence)

## Install

``` bash
$ go get -u github.com/huangjunwen/sqlw-mysql
```

## Design/Goals/Features

- Not an `ORM`, but provide similar features.
- Database first, `sqlw-mysql` generate wrapper code for your database tables.
- Use XML as DSL to describe query statements, `sqlw-mysql` generate wrapper code for them.
- Should be work for all kinds of queries, from simple ones to complex ones.
- Genreated code should be simple, easy to understand, and convenient enough to use.
- Highly customizable code template.
- Extensible DSL.

## Quickstart

Let's start with a small example.

Suppose you have a database with two tables: `user` and `employee`; An `employee` must be a `user`, but a `user` need not to be an `employee`; Each `employee` must have a superior except those top dogs.

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


Now run `sqlw-mysql` against the database. You will see a `models` directory is created with several source files generated:

``` bash
$ sqlw-mysql -dsn "user:passwd@tcp(host:port)/db?parseTime=true"
$ ls ./models
... table_user.go ... table_employee.go

```

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

But eventually, you will need more complex quries. For example if you want to query all `user` and its associated `employee` (e.g. `one2one` relationship), then you can write a statement XML like this:

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

A statement XML contains SQL statement with special directives embeded in. Here you can see two `<wc table="table_name">` directives, which are roughly equal to expanded `table_name.*`. 

_See [Statement XML](#statement_xml) for detail._

Now run `sqlw-mysql` again with the statement XML directory:

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
func AllUserEmployeeInfo(ctx context.Context, q Queryer) (AllUserEmployeeInfoResultSlice, error) {
  // ...
}

```

Notice that `User` and `Empl` fields in result struct are generated from those `<wc>` directives. `sqlw-mysql` is smart enough to figure out their correct positions. _See [here](#how-wildcard-directive-works) for detail._

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

_See [Directives](#directives) for detail._

After re-running the command, the following code is generated:

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
func EmployeeInfo(ctx context.Context, q Queryer, id ...int) (EmployeeInfoResultSlice, error) {
  // ...
}
```

Then, you can iterate the result like:

``` go
slice, err := EmployeeInfo(ctx, db, ids...)
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

In fact, `sqlw-mysql` doesn't care about what kind of relationships between result fields. It just generate helper methods such as `GroupByXXX`/`DistinctXXX` for these fields, thus it works for all kinds of relationships.


## Statement XML

`sqlw-mysql` use XML as DSL to describe quries, since XML is suitable for mixed content: raw SQL query and special directives. 

The simplest one is a `<stmt>` element with `name` attribute, without any directive, like this:

``` xml
<stmt name="One">
  SELECT 1
</stmt>
```

But this is not very useful, sometimes we want to add meta data to it, sometimes we want to reduce verbosity ...

That's why we need directives:

### Directives

Directive represents a fragment of SQL query, usually declared by an XML element. `sqlw-mysql` processes directives in several passes:

- The first pass all directives should generate fragments that form a valid SQL statement. This SQL statement is then used to determine statement type, to obtain result column information by querying against the database if it's a SELECT, e.g. `SELECT * FROM user WHERE id=1`
- The second pass all directives should generate fragments that form a text statement for template renderring. It's no need to be a valid SQL statement, it's up to the template to decide how to use this text, e.g. `SELECT * FROM user WHERE id=:id`
- Some directives may run extra pass.

Here is a list of current builtin directives:

| Directive | Example | First pass result | Second pass result | Extra pass | Note |
|-----------|---------|-------------------|--------------------|------------|------|
| `<arg>` | `<arg name="id" type="int" />` | `""` | `""` | | Declare a wrapper function argument. It always returns empty string |
| `<vars>` | `<vars flag1="true" flag2="false" />` | `""` | `""` | | Declare arbitary key/value pairs (XML attributes) for template to use. It always returns empty string |
| `<repl>` | `<repl with=":id">1</repl>` | `"1"` | `":id"` | | It returns the inner text for the first pass and returns the value in `with` attribute for the second pass |
| `<wc>` | `<wc table="employee" as="empl" />` | ```"`empl`.`id`, ..., `empl`.`superior_id`"``` | ```"`empl`.`id`, ..., `empl`.`superior_id`"``` | Run an extra pass to determine fields positions, see [here](#how-wildcard-directive-works) for detail | Always returns an expanded column list of the table |

#### How wildcard directive works

In the extra pass of `<wc>` directives, special marker columns are added before and after each `<wc>` directive, for example:

``` xml
  SELECT NOW(), <wc table="user" />, NOW() FROM user
```

will be expanded to something like:

```
  SELECT NOW(), 1 AS wc456958346a616564_0_s, `user`.`id`, ..., `user`.`birthday`, 1 AS wc456958346a616564_0_e, NOW() FROM user
```

By finding these marker column name in the result columns, `sqlw-mysql` can determine their positions.

This even works for subquery:

``` xml
  SELECT * FROM (SELECT <wc table="user" /> FROM user) AS u
```

And if you only selects a single column (or a few columns) like:

``` xml
  SELECT birthday FROM (SELECT <wc table="user" /> FROM user) AS u
```

Then the wildcard directive is ignored since you're not selecting all columns of the table.

## Code template

_TODO_

### Default template

_TODO_

## Command line options

``` bash
$ sqlw-mysql -h
Usage of sqlw-mysql:
  -blacklist value
    	(Optional) Comma separated table names not to render.
  -dsn string
    	(Required) Data source name. e.g. "user:passwd@tcp(host:port)/db?parseTime=true"
  -out string
    	(Optional) Output directory for generated code. (default "models")
  -pkg string
    	(Optional) Alternative package name of the generated code.
  -stmt string
    	(Optional) Statement xmls directory.
  -tmpl string
    	(Optional) Custom templates directory.
  -whitelist value
    	(Optional) Comma separated table names to render.

```

## Motivation

This tool is inspired by [xo](https://github.com/xo/xo) and influenced by tools like [sqlboiler](https://github.com/volatiletech/sqlboiler) since I'm a big fan of code generation and database first approach. However, some issues arised when using them.

For xo:
- The support for MySQL seems not very well
  - [Issue #123](https://github.com/xo/xo/issues/123)
  - Can't handle queries like: `SELECT user.*, employee.* FROM user LEFT JOIN employee ON employee.user_id=user.id;` since it use `CREATE VIEW ...` to obtain result column information but `CREATE VIEW` does not permit duplicate column names: both `user` and `employee` have `id`. As a result, you must hand write aliases for all the columns if you want to select them all.
- The DSL is quite limited, you really need to hand write every bit of the SQL.

For sqlboiler:
- It seems that outer join is not supported yet? [Issue #153](https://github.com/volatiletech/sqlboiler/issues/153)
- The genreated code quite large.

## Licence

MIT

Author: huangjunwen (kassarar@gmail.com)
