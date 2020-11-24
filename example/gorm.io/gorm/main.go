package main

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/j2gg0s/otsql"
	"github.com/j2gg0s/otsql/example"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/trace"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenMySQL() (db *gorm.DB, err error) {
	dbDSN := "otsql_user:otsql_password@tcp(localhost:3306)/otsql_db?charset=utf8&parseTime=True&loc=Local"

	driverName, err := otsql.Register(
		"mysql",
		otsql.WithAllowRoot(true),
		otsql.WithQuery(true),
		otsql.WithDefaultLabels(label.String("driver", "go-sql-driver/mysql")),
	)
	if err != nil {
		return nil, err
	}
	db, err = gorm.Open(mysql.New(mysql.Config{
		DriverName: driverName,
		DSN:        dbDSN,
	}), &gorm.Config{})
	return
}

func OpenPGWithPQ() (db *gorm.DB, err error) {
	dbDSN := "user=otsql_user password=otsql_password dbname=otsql_db host=localhost port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	driverName, err := otsql.Register(
		"postgres",
		otsql.WithAllowRoot(true),
		otsql.WithQuery(true),
		otsql.WithDefaultLabels(label.String("driver", "lib/pq")),
	)
	if err != nil {
		return nil, err
	}
	db, err = gorm.Open(postgres.New(postgres.Config{
		DriverName: driverName,
		DSN:        dbDSN,
	}), &gorm.Config{})
	return
}

func OpenPGWithPGX() (db *gorm.DB, err error) {
	dbDSN := "user=otsql_user password=otsql_password dbname=otsql_db host=localhost port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	connConfig, err := pgx.ParseConfig(dbDSN)
	if err != nil {
		return nil, err
	}
	connConfig.PreferSimpleProtocol = true
	pgxDSN := stdlib.RegisterConnConfig(connConfig)

	driverName, err := otsql.Register(
		"pgx",
		otsql.WithAllowRoot(true),
		otsql.WithQuery(true),
		otsql.WithDefaultLabels(label.String("driver", "jackc/pgx")),
	)
	if err != nil {
		return nil, err
	}
	db, err = gorm.Open(postgres.New(postgres.Config{
		DriverName: driverName,
		DSN:        pgxDSN,
	}), &gorm.Config{})
	return
}

func main() {
	example.InitMeter()
	flush := example.InitTracer()
	defer flush()

	mysqlDB, err := OpenMySQL()
	if err != nil {
		panic(err)
	}
	go func() {
		db, err := mysqlDB.DB()
		if err != nil {
			fmt.Println(err)
		}
		if err := otsql.RecordStats(db, "mysql"); err != nil {
			fmt.Println(err)
		}
	}()
	pgWithPGX, err := OpenPGWithPGX()
	if err != nil {
		panic(err)
	}
	go func() {
		db, err := pgWithPGX.DB()
		if err != nil {
			fmt.Println(err)
		}
		if err := otsql.RecordStats(db, "postgres-pgx"); err != nil {
			fmt.Println(err)
		}
	}()
	pgWithPQ, err := OpenPGWithPQ()
	if err != nil {
		panic(err)
	}
	go func() {
		db, err := pgWithPQ.DB()
		if err != nil {
			fmt.Println(err)
		}
		if err := otsql.RecordStats(db, "postgres-pq"); err != nil {
			fmt.Println(err)
		}
	}()

	ctx, span := otel.GetTracerProvider().Tracer("github.com/j2gg0s/otsql").Start(
		context.Background(),
		"demo",
		trace.WithNewRoot())
	defer span.End()

	for _, db := range []*gorm.DB{mysqlDB, pgWithPGX, pgWithPQ} {
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
