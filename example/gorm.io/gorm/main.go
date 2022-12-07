package main

import (
	"context"
	"fmt"
	"time"

	"github.com/90poe/otsql/example"
	"github.com/90poe/otsql/hook/metric"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func WithMySQL() (db *gorm.DB, err error) {
	driverName, err := example.Register("mysql")
	if err != nil {
		panic(err)
	}
	return gorm.Open(mysql.New(mysql.Config{
		DriverName: driverName,
		DSN:        example.MySQLDSN,
	}), &gorm.Config{})
}

func WithPQ() (db *gorm.DB, err error) {
	driverName, err := example.Register("postgres")
	if err != nil {
		panic(err)
	}
	return gorm.Open(postgres.New(postgres.Config{
		DriverName: driverName,
		DSN:        example.PostgreSQLDSN,
	}), &gorm.Config{})
}

func main() {
	example.InitMeter()
	example.InitTracer()

	ctx := context.Background()
	mysqlDB, err := WithMySQL()
	if err != nil {
		panic(err)
	}
	go func() {
		db, _ := mysqlDB.DB()
		metric.Stats(ctx, db, "mysql@90POE", 5*time.Second)
	}()

	pgDB, err := WithPQ()
	if err != nil {
		panic(err)
	}
	go func() {
		db, _ := pgDB.DB()
		metric.Stats(ctx, db, "pg@90POE", 5*time.Second)
	}()

	example.AsyncCronPrint(time.Second * 5)

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
