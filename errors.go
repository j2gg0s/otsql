package otsql

import (
	"google.golang.org/grpc/codes"
)

func ErrToCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	return codes.Unknown
}
