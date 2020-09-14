package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"

	"github.com/j2gg0s/otsql"
	"github.com/j2gg0s/otsql/example"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var pgDSN = "postgres://otsql_user:otsql_password@localhost:5432/otsql_db?sslmode=disable"

func main() {
	example.InitMeter()
	flush := example.InitTracer()
	defer flush()

	driverName, err := otsql.Register(
		"pgx",
		otsql.WithAllowRoot(true),
		otsql.WithQuery(true),
		otsql.WithQueryParams(true),
		otsql.WithInstanceName("pgInDocker"),
	)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(driverName, pgDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	{
		ctx, span := global.TraceProvider().Tracer("github.com/j2gg0s/otsql").Start(
			context.Background(),
			"demoTrace",
			trace.WithNewRoot())
		defer span.End()
		rows, err := db.QueryContext(ctx, `SELECT NOW()`)
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
}
