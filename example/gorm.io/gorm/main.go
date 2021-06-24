package main

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/j2gg0s/otsql/example"
	_ "github.com/lib/pq"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func WithMySQL() (db *gorm.DB, err error) {
	driverName, err := example.Register("mysql")
	return gorm.Open(mysql.New(mysql.Config{
		DriverName: driverName,
		DSN:        example.MySQLDSN,
	}), &gorm.Config{})
}

func WithPQ() (db *gorm.DB, err error) {
	driverName, err := example.Register("postgres")
	return gorm.Open(postgres.New(postgres.Config{
		DriverName: driverName,
		DSN:        example.PostgreSQLDSN,
	}), &gorm.Config{})
}

func main() {
	example.InitMeter()
	example.InitTracer()

	mysqlDB, err := WithMySQL()
	if err != nil {
		panic(err)
	}
	pgDB, err := WithPQ()
	if err != nil {
		panic(err)
	}

	example.AsyncCronPrint(time.Second * 5)

	ctx := context.Background()
	for _, db := range []*gorm.DB{mysqlDB, pgDB} {
		rows, err := db.WithContext(ctx).Raw(`SELECT NOW()`).Rows()
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

	time.Sleep(1000 * time.Second)
	// curl http://localhost:2222/metrics to get metrics
}
