package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/j2gg0s/otsql/example"
)

func main() {
	example.InitMeter()
	example.InitTracer()

	driverName, err := example.Register("mysql")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	db, err := sql.Open(driverName, example.MySQLDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("ping db: %w", err))
	}

	example.AsyncCronPrint(time.Second * 5)

	{
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

		fmt.Printf(
			"Current: %s, SELECT from db: %s\n",
			time.Now().String(),
			currentTime.String())
	}

	time.Sleep(600 * time.Second)
	// curl http://localhost:2222/metrics to get metrics
}
