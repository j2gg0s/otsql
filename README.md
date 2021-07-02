# otsql

[![Docs](https://godoc.org/github.com/j2gg0s/otsql?status.svg)](https://pkg.go.dev/github.com/j2gg0s/otsql)
[![Go Report Card](https://goreportcard.com/badge/github.com/j2gg0s/otsql)](https://goreportcard.com/report/github.com/j2gg0s/otsql)

Add an otsql wrapper to your existing database code to hook any sql command.

-   Support tracing with [OpenTelemetry](https://opentelemetry.io/) by [otsql/hook/trace](https://github.com/j2gg0s/otsql/tree/bun/hook/metric).
-   Support monitor latency and connection pool stats with [Prometheus](https://github.com/prometheus/prometheus)
    by [otsql/hook/metric](https://github.com/j2gg0s/otsql/tree/bun/hook/metric).
-   Support acess log with [zerolog](https://github.com/rs/zerolog) by [otsql/hook/trace](https://github.com/j2gg0s/otsql/tree/bun/hook/trace).

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
    "github.com/j2gg0s/otsql/hook/trace"
    "github.com/j2gg0s/otsql/hook/metric"
    "github.com/j2gg0s/otsql/hook/log"
    _ "github.com/go-sql-driver/mysql"
)

var dsn = "postgres://otsql_user:otsql_password@localhost/otsql_db?sslmode=disable"

metricHook, err := metric.New()
if err != nil {
    return "", fmt.Errorf("new metric hook: %w", err)
}

driverName, err := otsql.Register(
    name,
    otsql.WithHooks(
        trace.New(
            trace.WithAllowRoot(true),
            trace.WithQuery(true),
            trace.WithQueryParams(true),
        ),
        metricHook,
        log.New(),
    ),
)
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
    "github.com/j2gg0s/otsql/hook/trace"
    _ "github.com/lib/pq"
)

var dsn = "postgres://otsql_user:otsql_password@localhost/otsql_db?sslmode=disable"

connector, err := pq.NewConnector(dsn)
if err != nil {
    panic(err)
}

db := sql.OpenDB(
    WrapConnector(connector, otsql.Hooks(trace.new())))
defer db.Close()
```

See more specific case in `example/`.

## Trace with opentelemetry

otsql support trace with opentelemetry by `hook/trace`.

## Metric with prometheus

otsql support metric with prometheus by `hook/metric`.

| Metric                 | Search suffix  | Tags                                 |
| ---------------------- | -------------- | ------------------------------------ |
| Latency in millisecond | go_sql_latency | sql_instance, sql_method, sql_status |

You can use `metric.Stats` to monitor connection pool, all metric supprt tag `sql_instance`.
| Metric | Search suffix |
|--------|---------------|
| The number of connections currently in use | go_sql_conn_in_use |
| The number of idle connections | go_sql_conn_idle |
| The total number of connections wait for | go_sql_conn_wait |
| The total number of connections closed because of SetMaxIdleConns | go_sql_conn_idle_closed |
| The total number of connections closed because of SetConnMaxLifetime | go_sql_conn_lifetime_closed |
| The total time blocked by waiting for a new connection, nanosecond | go_sql_conn_wait_ns |

## Test

Test by [bun](https://github.com/uptrace/bun)'s unittest with a special branch [otsql@bun](https://github.com/j2gg0s/bun/tree/otsql).

Test by [gorm](https://github.com/go-gorm/gorm)'s unittest with a special branch [otsql@gorm](https://github.com/j2gg0s/gorm/tree/otsql).
It also includes example for [pgx](https://github.com/jackc/pgx).

## Benchmark

Run benchmark with [bun](https://github.com/uptrace/bun):

```bash
goos: darwin
goarch: amd64
pkg: github.com/uptrace/bun/internal/dbtest
cpu: Intel(R) Core(TM) i7-8557U CPU @ 1.70GHz
BenchmarkSelectOne/pg-8             4579            281209 ns/op
BenchmarkSelectOne/mysql-8                  4146            292964 ns/op
BenchmarkSelectOne/sqlite-8                37898             28012 ns/op
BenchmarkSelectSlice/mysql-8                2874            397543 ns/op
BenchmarkSelectSlice/sqlite-8               6517            184153 ns/op
BenchmarkSelectSlice/pg-8                   2806            425935 ns/op
BenchmarkSelectError/pg-8                   4027            314399 ns/op
BenchmarkSelectError/mysql-8                2426            476420 ns/op
BenchmarkSelectError/sqlite-8             166348              7320 ns/op
PASS
ok      github.com/uptrace/bun/internal/dbtest  79.494s
```

Run benchmark with [otsql@bun](https://github.com/j2gg0s/bun/tree/otsql):

```bash
goos: darwin
goarch: amd64
pkg: github.com/uptrace/bun/internal/dbtest
cpu: Intel(R) Core(TM) i7-8557U CPU @ 1.70GHz
BenchmarkSelectOne/pg-8             3014            332012 ns/op
BenchmarkSelectOne/mysql-8                  3382            339056 ns/op
BenchmarkSelectOne/sqlite-8                25057             46897 ns/op
BenchmarkSelectSlice/pg-8                   2284            468973 ns/op
BenchmarkSelectSlice/mysql-8                3074            426501 ns/op
BenchmarkSelectSlice/sqlite-8               4725            234213 ns/op
BenchmarkSelectError/pg-8                   3760            337402 ns/op
BenchmarkSelectError/mysql-8                2647            485405 ns/op
BenchmarkSelectError/sqlite-8              73551             15706 ns/op
PASS
ok      github.com/uptrace/bun/internal/dbtest  67.224s
```
