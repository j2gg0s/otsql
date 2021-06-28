package otsql

import "google.golang.org/grpc/codes"

func ErrToCode(err error) codes.Code {
	switch err {
	case nil:
		return codes.OK
	default:
		return codes.Unknown
	}
}
