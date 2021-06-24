module github.com/j2gg0s/otsql/example/gorm.io/gorm

go 1.14

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/j2gg0s/otsql v0.0.0-20200914101133-297167817f42
	github.com/lib/pq v1.3.0
	gorm.io/driver/mysql v1.0.1
	gorm.io/driver/postgres v1.0.0
	gorm.io/gorm v1.20.1
)

replace github.com/j2gg0s/otsql => ../../../
