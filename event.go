package otsql

import (
	"context"
	"time"
)

type Hook interface {
	Before(context.Context, *Event) context.Context
	After(context.Context, *Event)
}

func before(hooks []Hook, ctx context.Context, evt *Event) context.Context {
	for _, hook := range hooks {
		ctx = hook.Before(ctx, evt)
	}
	return ctx
}

func after(hooks []Hook, ctx context.Context, evt *Event) {
	for _, hook := range hooks {
		hook.After(ctx, evt)
	}
}

type Method string

var (
	MethodPing Method = "ping"

	MethodExec     Method = "exec"
	MethodQuery    Method = "query"
	MethodPrepare  Method = "prepare"
	MethodBegin    Method = "begin"
	MethodCommit   Method = "commit"
	MethodRollback Method = "rollback"

	MethodLastInsertId Method = "last_insert_id"
	MethodRowsAffected Method = "rows_affected"
	MethodRowsClose    Method = "rows_close"
	MethodRowsNext     Method = "rows_next"

	MethodCreateConn   Method = "create_conn"
	MethodCloseConn    Method = "close_conn"
	MethodResetSession Method = "reset_session"
)

type Event struct {
	Instance string
	Database string

	Method  Method
	Query   string
	Args    interface{}
	BeginAt time.Time

	Err error

	CloseFuncs []func(context.Context, error)

	Conn string
}

func newEvent(o *Options, conn string, method Method, query string, args interface{}) *Event {
	return &Event{
		Instance: o.Instance,
		Database: o.Database,

		Conn:    conn,
		Method:  method,
		Query:   query,
		Args:    args,
		BeginAt: time.Now(),
	}
}
