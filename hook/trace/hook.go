package trace

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strconv"

	"github.com/j2gg0s/otsql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	sqlQuery = "sql.query"
)

type Hook struct {
	*Options

	Tracer trace.Tracer
}

func New(options ...Option) *Hook {
	return &Hook{
		Options: newOptions(options),
		Tracer:  otel.GetTracerProvider().Tracer("github.com/j2gg0s/otsql"),
	}
}

var _ otsql.Hook = (*Hook)(nil)

func (hook *Hook) Before(ctx context.Context, evt *otsql.Event) context.Context {
	if !hook.AllowRoot && !trace.SpanFromContext(ctx).IsRecording() {
		return ctx
	}

	switch evt.Method {
	case otsql.MethodPing:
		if !hook.Ping {
			return ctx
		}
	case otsql.MethodRowsAffected:
		if !hook.RowsAffected {
			return ctx
		}
	case otsql.MethodLastInsertId:
		if !hook.LastInsertId {
			return ctx
		}
	case otsql.MethodRowsNext:
		if !hook.RowsNext {
			return ctx
		}
	case otsql.MethodRowsClose:
		if !hook.RowsClose {
			return ctx
		}
	case otsql.MethodResetSession:
		if !hook.ResetSession {
			return ctx
		}
	}

	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
	}

	attrs := hook.attrsFromSQL(evt.Query, evt.Args)
	attrs = append(
		attrs,
		sqlInstance.String(evt.Instance),
		sqlDatabase.String(evt.Database),
	)
	opts = append(opts, trace.WithAttributes(attrs...))

	spanName := hook.SpanNameFormatter(ctx, string(evt.Method), evt.Query)
	ctx, _ = hook.Tracer.Start(ctx, spanName, opts...)

	return ctx
}

func (h *Hook) After(ctx context.Context, evt *otsql.Event) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	code, err := codes.Ok, evt.Err
	if err != nil {
		span.RecordError(err)
		code = codes.Error
	}
	span.SetStatus(code, otsql.ErrToCode(err).String())
	span.End()
}

var (
	attributeUnknownArgs = attribute.String("otsql.warning", "unknown args type")
)

func (hook *Hook) attrsFromSQL(query string, args interface{}) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	if len(hook.DefaultAttributes) > 0 {
		attrs = append(attrs, hook.DefaultAttributes...)
	}

	if hook.Query && len(query) > 0 {
		attrs = append(attrs, attribute.String(sqlQuery, query))
	}

	if hook.QueryParams && args != nil {
		switch sqlArgs := args.(type) {
		case []driver.NamedValue:
			for _, arg := range sqlArgs {
				if len(arg.Name) > 0 {
					attrs = append(attrs, argToAttr(arg.Name, arg.Value))
				} else {
					attrs = append(attrs, argToAttr(strconv.Itoa(arg.Ordinal), arg.Value))
				}
			}
		case []driver.Value:
			for i, arg := range sqlArgs {
				attrs = append(attrs, argToAttr(strconv.Itoa(i), arg))
			}
		default:
			attrs = append(attrs, attributeUnknownArgs)
		}
	}

	return attrs
}

func argToAttr(k string, v driver.Value) attribute.KeyValue {
	return attribute.String(fmt.Sprintf("sql.arg.%s", k), fmt.Sprintf("%v", v))
}

var (
	sqlInstance = attribute.Key("sql.instance")
	sqlDatabase = attribute.Key("sql.database")
)
