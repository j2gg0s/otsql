package main

import (
	"context"
	"database/sql"
	"fmt"

	"go.opentelemetry.io/otel/exporters/stdout"

	"github.com/j2gg0s/otsql"
	_ "github.com/lib/pq"
)

func initOtel() {
	_, err := stdout.InstallNewPipeline([]stdout.Option{
		stdout.WithPrettyPrint(),
	}, nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	initOtel()

	connURL := "postgres://postgres:rcrai@localhost/postgres?sslmode=disable"

	driverName, err := otsql.Register("postgres", otsql.TraceOptions{
		AllowRoot:   true,
		Query:       true,
		QueryParams: true,
	})
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(driverName, connURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.QueryContext(context.Background(), `SELECT id, name, status FROM "user" LIMIT 1`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var id, name string
	var status int
	for rows.Next() {
		err = rows.Scan(&id, &name, &status)
		if err != nil {
			panic(err)
		}

		fmt.Println(id, name, status)
	}
}
