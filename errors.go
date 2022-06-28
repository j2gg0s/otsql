package otsql

import (
	"github.com/go-sql-driver/mysql"
	"google.golang.org/grpc/codes"
)

func ErrToCode(err error) codes.Code {
	switch err {
	case nil:
		return codes.OK
	default:
		if code, ok := mysqlErr(err); ok {
			return code
		}
		return codes.Unknown
	}
}

func mysqlErr(err error) (codes.Code, bool) {
	me, ok := err.(*mysql.MySQLError)
	if !ok {
		return 0, false
	}
	return codes.Code(me.Number), true
}
