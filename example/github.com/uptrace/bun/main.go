package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/j2gg0s/otsql"
	"github.com/j2gg0s/otsql/example"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

var mysqlDSN = "otsql_user:otsql_password@/otsql_db?parseTime=true"

func main() {
	example.InitMeter()
	example.InitTracer()

	driverName, err := otsql.Register(
		"mysql",
		otsql.WithAllowRoot(true),
		otsql.WithQuery(true),
		otsql.WithQueryParams(true),
		otsql.WithInstanceName("mysqlInDocker"),
	)
	if err != nil {
		panic(err)
	}

	sqldb, err := sql.Open(driverName, mysqlDSN)
	if err != nil {
		panic(err)
	}
	defer sqldb.Close()
	go otsql.RecordStats(sqldb, "mysql")

	db := bun.NewDB(sqldb, mysqldialect.New())
	defer db.Close()

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT NOW()`)
	if err != nil {
		panic(err)
	}
	var t time.Time
	for rows.Next() {
		err := rows.Scan(&t)
		if err != nil {
			panic(err)
		}
		fmt.Println(t)
	}

	time.Sleep(1000 * time.Second)
	// curl http://localhost:2222/metrics to get metrics
}
