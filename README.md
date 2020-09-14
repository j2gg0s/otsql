# otsql

[OpenTelemetry](https://opentelemetry.io/) SQL database driver wrapper, not official.

Add an otsql wrapper to your existing database code to instrument the
interactions with the database.

First version transformed from [ocsql](https://github.com/opencensus-integrations/ocsql).

## Install
```
$ go get -u github.com/j2gg0s/otsql
```

## Usage
To use otsql with your application, register an otsql wrapper of a database driver
as shown below.

Example:
```go
import (
    "database/sql"

    "github.com/j2gg0s/otsql"
    _ "github.com/lib/pq"
)

var dsn = "postgres://otsql_user:otsql_password@localhost/otsql_db?sslmode=disable"

driverName, err := otsql.Register("postgres", otsql.WithQuery(true))
if err != nil {
    panic(err)
}

db, err := sql.Open(driverName, dsn)
if err != nil {
    panic(err)
}
defer db.Close()
```

Finally database drivers that support the driver.Connector
interface can be wrapped directly by otsql without the need for otsql to
register a driver.Driver.

Example:
```go
import (
    "database/sql"

    "github.com/j2gg0s/otsql"
    _ "github.com/lib/pq"
)

var dsn = "postgres://otsql_user:otsql_password@localhost/otsql_db?sslmode=disable"

connector, err := pq.NewConnector(dsn)
if err != nil {
    panic(err)
}

db := sql.OpenDB(
    WrapConnector(connector, otsql.WithQuery(true)))
defer db.Close()
```

See more specific case in ``example/``.

## Metric And Span
TODO

## Test
We add wrap to [gorm](https://github.com/go-gorm/gorm) and run its test with a forked repo [j2gg0s/gorm](https://github.com/j2gg0s/gorm) .
