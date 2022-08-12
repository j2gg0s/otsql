package metric

import (
	"github.com/go-sql-driver/mysql"
	"google.golang.org/grpc/codes"
)

func ErrToCodeMySQL(err error) string {
	if err == nil {
		return codes.OK.String()
	}

	me, ok := err.(*mysql.MySQLError)
	if !ok {
		return codes.Unknown.String()
	}

	switch me.Number() {
	case 1062:
		return codes.AlreadyExists.String()
	}

	return codes.Unknown.String()
}
