package otsql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetInstanceFromDSN(t *testing.T) {
	fixtures := []struct {
		dsn      string
		instance string
		database string
	}{
		{
			// go-sql-driver/mysql
			"username:password@protocol(address)/dbname?param=value",
			"protocol(address)",
			"dbname",
		},
		{
			// jackc/pgx
			"postgres://username:password@localhost:5432/database_name",
			"localhost:5432",
			"database_name",
		},
		{
			// postgresql: https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
			"host=localhost port=5432 dbname=mydb connect_timeout=10",
			"localhost:5432",
			"mydb",
		},
	}

	for _, f := range fixtures {
		fixture := f
		t.Run(f.dsn, func(t *testing.T) {
			instance, database := parseDSN(fixture.dsn)
			require.Equal(t, fixture.instance, instance)
			require.Equal(t, fixture.database, database)
		})
	}
}
