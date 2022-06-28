package otsql

import (
	"database/sql/driver"
	"errors"

	"google.golang.org/grpc/codes"
)

func ErrToCode(err error) codes.Code {
	if err == nil || errors.Is(err, driver.ErrSkip) {
		return codes.OK
	}
	return codes.Unknown
}
