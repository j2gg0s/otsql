package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/j2gg0s/otsql/example"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

type User struct {
	bun.BaseModel `bun:"user"`

	ID   string `bun:",pk"`
	Name string `bun:"type:VARCHAR(255)"`

	Books []Book `bun:"type:JSON"`

	CreatedAt time.Time `bun:",nullzero,notnull,type:DATETIME,default:CURRENT_TIMESTAMP"`
}

type Book struct {
	ID   string
	Name string
}

func main() {
	example.InitMeter()
	example.InitTracer()

	driverName, err := example.Register("mysql")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	sqldb, err := sql.Open(driverName, example.MySQLDSN)
	if err != nil {
		panic(err)
	}
	defer sqldb.Close()

	db := bun.NewDB(sqldb, mysqldialect.New())
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("ping db: %w", err))
	}

	example.AsyncCronPrint(time.Second * 5)

	{
		_, err := db.NewCreateTable().Model((*User)(nil)).IfNotExists().Exec(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Println("crete table user")
	}

	id := fmt.Sprintf("j2gg0s-%d", time.Now().Unix())
	{
		user := User{
			ID:   id,
			Name: "j2gg0s",
			Books: []Book{
				Book{
					ID:   "mysql",
					Name: "mysql",
				},
				Book{
					ID:   "debezium",
					Name: "debezium",
				},
			},
		}
		_, err := db.NewInsert().Model(&user).Exec(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("insert user with id(%s), created_at: %s", id, user.CreatedAt.Format(time.RFC3339))
	}

	{
		user := User{}
		err := db.NewSelect().Model(&user).Where("id=?", id).Scan(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("select user with id(%s): name(%s), created_at: %s\n", id, user.Name, user.CreatedAt.Format(time.RFC3339))
	}

	{
		r, err := db.NewDelete().Model((*User)(nil)).Where("id=?", id).Exec(ctx)
		if err != nil {
			panic(err)
		}
		rowsAffected, err := r.RowsAffected()
		if err != nil {
			panic(err)
		}
		fmt.Printf("delete user with id(%s): %d\n", id, rowsAffected)
	}

	{
		_, err := db.ExecContext(ctx, `DROP TABLE user`)
		if err != nil {
			panic(err)
		}
		fmt.Printf("drop table user")
	}

	time.Sleep(1000 * time.Second)
	// curl http://localhost:2222/metrics to get metrics
}
