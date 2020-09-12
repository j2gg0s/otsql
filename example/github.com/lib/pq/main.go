package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/j2gg0s/otsql"
	_ "github.com/lib/pq"
)

func initTracer() func() {
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "otsql",
		}),
		jaeger.WithSDK(&trace.Config{
			DefaultSampler: trace.AlwaysSample(),
		}),
	)
	if err != nil {
		panic(err)
	}
	return flush
}

func initMeter() {
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{})
	if err != nil {
		panic(err)
	}

	go func() {
		http.ListenAndServe(":2222", exporter)
	}()
}

func main() {
	initMeter()
	flush := initTracer()
	defer flush()

	connURL := "postgres://otsql_user:otsql_password@localhost:5432/otsql_db?sslmode=disable"

	driverName, err := otsql.Register(
		// github.com/lib/pq register with this driver name
		"postgres",
		otsql.WithAllowRoot(true),
		otsql.WithQuery(true),
		otsql.WithQueryParams(true),
	)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(driverName, connURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT 1+1`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var c int
	for rows.Next() {
		err = rows.Scan(&c)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println(c)
	// output: 2
}
