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
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var pgDSN = "postgres://otsql_user:otsql_password@localhost:5432/otsql_db?sslmode=disable"

func main() {
	example.InitMeter()
	flush := example.InitTracer()
	defer flush()

	connConfig, err := pgx.ParseConfig(pgDSN)
	if err != nil {
		panic(err)
	}
	// Change and register pgx config
	connConfig.PreferSimpleProtocol = true
	pgxDSN := stdlib.RegisterConnConfig(connConfig)

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

	db, err := sql.Open(driverName, pgxDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	go func() {
		if err := otsql.RecordStats(db, "postgres"); err != nil {
			fmt.Println(err)
		}
	}()

	{
		ctx, span := global.TracerProvider().Tracer("github.com/j2gg0s/otsql").Start(
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

	time.Sleep(10 * time.Second)
	// curl http://localhost:2222/metrics to get metrics
}
