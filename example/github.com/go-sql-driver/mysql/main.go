package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"

	_ "github.com/go-sql-driver/mysql"
	"github.com/j2gg0s/otsql"
	"github.com/j2gg0s/otsql/example"
)

var mysqlDSN = "otsql_user:otsql_password@/otsql_db?parseTime=true"

func main() {
	example.InitMeter()
	flush := example.InitTracer()
	defer flush()

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

	db, err := sql.Open(driverName, mysqlDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	go otsql.RecordStats(db, "mysql")

	{
		ctx, span := global.TracerProvider().Tracer("github.com/j2gg0s/otsql").Start(
			context.Background(),
			"demoTrace",
			trace.WithNewRoot())
		defer span.End()
		rows, err := db.QueryContext(ctx, `SELECT CURRENT_TIMESTAMP`)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		var currentTime time.Time
		for rows.Next() {
			err = rows.Scan(&currentTime)
			if err != nil {
				panic(err)
			}
		}

		fmt.Println(currentTime)
	}

	time.Sleep(10 * time.Second)
	// curl http://localhost:2222/metrics to get metrics
}
