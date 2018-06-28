# SQL wrapper code generator for MySQL [![Go Report Card](https://goreportcard.com/badge/github.com/huangjunwen/sqlw-mysql)](https://goreportcard.com/report/github.com/huangjunwen/sqlw-mysql)

`sqlw-mysql` is a CLI tool to generate go wrapper code for your MySQL database and queries.

## Table of Contents

- [Install](#install)
- [Design/Goals/Features](#designgoalsfeatures)
- [Quickstart](#quickstart)
- [Statement XML](#statement-xml)
  - [Directives](#directives)
    - [Arg directive](#arg-directive)
    - [Vars directive](#vars-directive)
    - [Replace directive](#replace-directive)
    - [Text directive](#text-directive)
    - [Wildcard directive](#wildcard-directive)
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
- Extensible DSL (through directives).

## Quickstart

Let's start with a small example. (See [here](https://github.com/huangjunwen/sqlw-mysql/tree/master/examples/quickstart) for complete source code)

Suppose you have a database with two tables: `user` and `employee`; An `employee` must be a `user`, but a `user` need not to be an `employee`; Each `employee` must have a superior except those top dogs.

``` sql
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

func (tr *User) Insert(ctx context.Context, e Execer) error {
  // ...
}

func (tr *User) Reload(ctx context.Context, q Queryer) error {
  // ...
}
```

But eventually, you will need more complex quries. For example if you want to query all `user` and its associated `employee` (e.g. `one2one` relationship), then you can write a statement XML like this:

``` xml
<!-- ./stmts/user.xml -->

<stmt name="AllUserEmployees">
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

// AllUserEmployeesResult is the result of AllUserEmployees.
type AllUserEmployeesResult struct {
  User       *User
  Age        null.Uint64
  Empl       *Employee
  nxNullUser nxNullUser
  nxNullEmpl nxNullEmployee
}

// AllUserEmployeesResultSlice is slice of AllUserEmployeesResult.
type AllUserEmployeesResultSlice []*AllUserEmployeesResult

// ...
func AllUserEmployees(ctx context.Context, q Queryer) (AllUserEmployeesResultSlice, error) {
  // ...
}

```

Notice that `User` and `Empl` fields in result struct are generated from those `<wc>` directives. `sqlw-mysql` is smart enough to figure out their correct positions. _See [here](#how-wildcard-directive-works) for detail._

Now you can use the newly created function to iterate through all `user` and `employee`:

``` go
slice, err := models.AllUserEmployees(ctx, tx)
if err != nil {
  log.Fatal(err)
}

for _, result := range slice {
  user := result.User
  empl := result.Empl

  if empl.Valid() {
    log.Printf("User %+q (age %d) is an employee, sn: %+q\n", user.Name, result.Age.Uint64, empl.EmployeeSn)
  } else {
    log.Printf("User %+q (age %d) is not an employee\n", user.Name, result.Age.Uint64)
  }
}
```

Another example, if you want to find subordinates of some employees (e.g. `one2many` relationship):

``` xml
<!-- ./stmts/user.xml -->

<stmt name="SubordinatesBySuperiors">
  <a name="id" type="...int" />
  <v in_query="1" />
  SELECT
    <wc table="employee" as="superior" />,
    <wc table="employee" as="subordinate" />
  FROM
    employee AS superior LEFT JOIN employee AS subordinate ON subordinate.superior_id=superior.id
  WHERE
    superior.id IN (<r by=":id">1</r>)
</stmt>
```

Brief explanation about new directives:
- `<a>` specifies an argument of the generated function.
- `<v>` specifies arbitary variables that the template can use. `in_query="1"` tells the template that the SQL use `IN` operator.
- `<r>` can replace arbitary statement text.

_See [Directives](#directives) for detail._

After re-running the command, the following code is generated:

``` go
// ./models/stmt_user.go

// SubordinatesBySuperiorsResult is the result of SubordinatesBySuperiors.
type SubordinatesBySuperiorsResult struct {
  Superior          *Employee
  Subordinate       *Employee
  nxNullSuperior    nxNullEmployee
  nxNullSubordinate nxNullEmployee
}

// SubordinatesBySuperiorsResultSlice is slice of SubordinatesBySuperiorsResult.
type SubordinatesBySuperiorsResultSlice []*SubordinatesBySuperiorsResult

// ...
func SubordinatesBySuperiors(ctx context.Context, q Queryer, id ...int) (SubordinatesBySuperiorsResultSlice, error) {
  // ...
}
```

Then, you can iterate the result like:

``` go
slice, err := models.SubordinatesBySuperiors(ctx, tx, 1, 2, 3, 4, 5, 6, 7)
if err != nil {
  log.Fatal(err)
}

superiors, groups := slice.GroupBySuperior()
for i, superior := range superiors {
  subordinates := groups[i].DistinctSubordinate()

  if len(subordinates) == 0 {
    log.Printf("Employee %+q has no subordinate.\n", superior.EmployeeSn)
  } else {
    log.Printf("Employee %+q has the following subordinates:\n", superior.EmployeeSn)
    for _, subordinate := range subordinates {
      log.Printf("\t%+q\n", subordinate.EmployeeSn)
    }
  }
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

- The first pass all directives should generate fragments that form a valid SQL statement (e.g. `SELECT * FROM user WHERE id=1`). This SQL statement is then used to determine statement type, to obtain result column information by querying against the database if it's a SELECT. 
- The second pass all directives should generate fragments that form a text statement for template renderring (e.g. `SELECT * FROM user WHERE id=:id`). It's no need to be a valid SQL statement, it's up to the template to decide how to use this text.
- Some directives may run extra pass.

The following are a list of current builtin directives. In future new directives may be added. And should be easy enough to implement one: impelemnts a go interface.

#### Arg directive

- Name: `<arg>`/`<a>`
- Example: `<a name="id" type="int" />`
- First pass result: `""`
- Second pass result: `""`

Declare a wrapper function argument's name and type. Always returns empty string.

#### Vars directive

- Name: `<vars>`/`<v>`
- Example: `<v flag1="true" flag2="true" />`
- First pass result: `""`
- Second pass result: `""`

Declare arbitary key/value pairs (XML attributes) for template to use. Always returns empty string.

#### Replace directive

- Name: `<repl>`/`<r>`
- Example: `<r by=":id">1</r>`
- First pass result: `"1"`
- Second pass result: `":id"`

Returns the inner text for the first pass and returns the value in `by` attribute for the second pass.

#### Text directive

- Name: `<text>`/`<t>`
- Example: `<t>{{ if ne .id 0 }}</t>`
- First pass result: `""`
- Second pass result: `"{{ if ne .id 0 }}"`

`<t>innerText</t>` is equivalent to `<r by="innerText"></r>`.

#### Wildcard directive

- Name: `<wc>`
- Example: `<wc table="employee" as="empl" />`
- First pass result: ```"`empl`.`id`, ..., `empl`.`superior_id`"```
- Second pass result: ```"`empl`.`id`, ..., `empl`.`superior_id`"```

Returns the expanded column list of the table. It runs an extra pass to determine fields positions, see [here](#how-wildcard-directive-works) for detail.

##### How wildcard directive works

`<wc>` (wildcard) directive serves several purposes:

- Reduce verbosity, also it's a 'safer' version of `table.*` (It expands all columns of the table).
- Figure out the positions of these expanded columns so that template can make use of.

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

`sqlw-mysql` itself only provides information extracted from database/DSL. Most features are in fact implemented in code template. A code template is a directory looks like:

``` bash
$ tree default
default
├── group.tmpl
├── headnote
├── intf.tmpl
├── manifest.json
├── meta.tmpl
├── meta_test.tmpl
├── scan_type_map.json
├── stmt.tmpl
├── table.tmpl
└── test.tmpl
```

A `manifest.json` list all files that the code template consist of:

``` json
{
  "scan_type_map": "scan_type_map.json",
  "headnote": "headnote",
  "tmpl": {
    "table": "table.tmpl",
    "stmt": "stmt.tmpl",
    "etc": [
      "intf.tmpl",
      "meta.tmpl",
      "group.tmpl",
      "test.tmpl",
      "meta_test.tmpl"
    ]
  }
}
```

`manifest["scan_type_map"]` is used to map database type (key) to go scan type (value, `value[0]` is for *NOT* nullable type and `value[1]` is for nullable type):

``` json
{
  "float32":   ["float32", "null.Float32"],
  "float64":   ["float64", "null.Float64"],
  "bool":      ["bool", "null.Bool"],
  "int8":      ["int8", "null.Int8"],
  "uint8":     ["uint8", "null.Uint8"],
  "int16":     ["int16", "null.Int16"],
  "uint16":    ["uint16", "null.Uint16"],
  "int32":     ["int32", "null.Int32"],
  "uint32":    ["uint32", "null.Uint32"],
  "int64":     ["int64", "null.Int64"],
  "uint64":    ["uint64", "null.Uint64"],
  "time":      ["time.Time", "null.Time"],
  "bit":       ["string", "null.String"],
  "json":      ["string", "null.String"],
  "string":    ["string", "null.String"]
}
```

`manifest["tmpl"]` list file templates: `manifest["tmpl"]["table"]` is used to render each table found in database; `manifest["tmpl"]["stmt"]` is used to render each statement XML found; `manifest["tmpl"]["etc"]` contains extra file templates.



### Default template

If no custom code template specified, or `-stmt @default` is given, then the default template is used.

Genreated code depends on these external libraries:
- [sqlx](https://github.com/jmoiron/sqlx).
- `sqlboiler`'s [null-extended](https://github.com/volatiletech/null) package.

For statement XML, the default template accept these `<vars>`:

| Name | Example | Note |
|------|---------|------|
| `use_template` | `use_template="1"` | If presented, then the statement text is treated as a go [template](https://godoc.org/text/template) |
| `in_query` | `in_query="1"` | If presented, then statement will do an "IN" expansion, see http://jmoiron.github.io/sqlx/#inQueries |
| `return` | `return="one"` | For SELECT statement only, by default the generated function is returns a slice, if `return="one"`, then returns a single item instead |

An example of `use_template`:

``` xml
<stmt name="UsersByCond">
  <v use_template="1" />
  <a name="id" type="int" />
  <a name="name" type="string" />
  <a name="birthday" type="time.Time" />
  <a name="limit" type="int" />
  SELECT
    <wc table="user" />
  FROM
    user
  WHERE
    <t>{{ if ne .id 0 }}</t>
      id=<r by=":id">1</r> AND
    <t>{{ end }}</t>

    <t>{{ if ne (len .name) 0 }}</t>
      name=<r by=":name">"hjw"</r> AND
    <t>{{ end }}</t>

    <t>{{ if not .birthday.IsZero }}</t>
      birthday=<r by=":birthday">NOW()</r> AND
    <t>{{ end }}</t>
    1
  LIMIT <r by=":limit">10</r>
</stmt>
```

Then the generated statement will be treated as a go template and will be renderred before normal execution. This is useful when you have many `WHERE` condtions combination.


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
- The genreated code is quite large.

## Licence

MIT

Author: huangjunwen (kassarar@gmail.com)
