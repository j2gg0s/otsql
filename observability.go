package otsql

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/trace"
)

var (
	instrumentationName = "github.com/j2gg0s/otsql"

	tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	meter  = global.GetMeterProvider().Meter(instrumentationName)

	latencyValueRecorder, _ = meter.NewInt64ValueRecorder(
		"go.sql.latency",
		metric.WithDescription("The latency of calls in microsecond"),
	)
)

const (
	sqlInstance = "sql.instance"
	sqlMethod   = "sql.method"
	sqlQuery    = "sql.query"
	sqlStatus   = "sql.status"
)

var (
	statusOK    = attribute.String(sqlStatus, "OK")
	statusError = attribute.String(sqlStatus, "Error")

	methodPing     = attribute.String(sqlMethod, "ping")
	methodExec     = attribute.String(sqlMethod, "exec")
	methodQuery    = attribute.String(sqlMethod, "query")
	methodPrepare  = attribute.String(sqlMethod, "preapre")
	methodBegin    = attribute.String(sqlMethod, "begin")
	methodCommit   = attribute.String(sqlMethod, "commit")
	methodRollback = attribute.String(sqlMethod, "rollback")

	methodLastInsertID = attribute.String(sqlMethod, "last_insert_id")
	methodRowsAffected = attribute.String(sqlMethod, "rows_affected")
	methodRowsClose    = attribute.String(sqlMethod, "rows_close")
	methodRowsNext     = attribute.String(sqlMethod, "rows_next")

	methodCreateConn = attribute.String(sqlMethod, "create_conn")
)

func startMetric(_ context.Context, method attribute.KeyValue, start time.Time, options TraceOptions) func(context.Context, error) {
	attributes := []attribute.KeyValue{
		attribute.String(sqlInstance, options.InstanceName),
		method,
	}

	return func(ctx context.Context, err error) {
		if err != nil {
			attributes = append(attributes, statusError)
		} else {
			attributes = append(attributes, statusOK)
		}

		latencyValueRecorder.Record(ctx, time.Since(start).Microseconds(), attributes...)
	}
}

func startTrace(ctx context.Context, options TraceOptions, method attribute.KeyValue, query string, args interface{}) (context.Context, trace.Span, func(context.Context, error)) {
	if !options.AllowRoot && !trace.SpanFromContext(ctx).IsRecording() {
		return ctx, nil, func(context.Context, error) {}
	}

	if method == methodPing && !options.Ping {
		return ctx, nil, func(context.Context, error) {}
	}

	start := time.Now()
	endMetric := startMetric(ctx, method, start, options)

	opts := []trace.SpanOption{
		trace.WithSpanKind(trace.SpanKindClient),
	}
	attrs := attrsFromSQL(ctx, options, method, query, args)
	if len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}
	spanName := options.SpanNameFormatter(ctx, method.Value.AsString(), query)
	ctx, span := tracer.Start(ctx, spanName, opts...)

	return ctx, span, func(ctx context.Context, err error) {
		endMetric(ctx, err)

		if err != nil {
			span.RecordError(err)
		}
		code, msg := spanStatusFromSQLError(err)
		span.SetStatus(code, msg)
		span.End()
	}
}

func attrsFromSQL(_ context.Context, options TraceOptions, _ attribute.KeyValue, query string, args interface{}) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0)
	if len(options.DefaultAttributes) > 0 {
		attrs = append(attrs, options.DefaultAttributes...)
	}

	if options.Query && len(query) > 0 {
		attrs = append(attrs, attribute.String(sqlQuery, query))
	}
	if options.QueryParams && args != nil {
		switch sqlArgs := args.(type) {
		case []driver.NamedValue:
			for _, arg := range sqlArgs {
				if len(arg.Name) > 0 {
					attrs = append(attrs, argToLabel(arg.Name, arg.Value))
				} else {
					attrs = append(attrs, argToLabel(strconv.Itoa(arg.Ordinal), arg.Value))
				}
			}
		case []driver.Value:
			for i, arg := range sqlArgs {
				attrs = append(attrs, argToLabel(strconv.Itoa(i), arg))
			}
		default:
			attrs = append(attrs, attributeUnknownArgs)
		}
	}
	return attrs
}

func spanStatusFromSQLError(err error) (code codes.Code, msg string) {
	switch err {
	case nil:
		code = codes.Ok
	default:
		code = codes.Error
	}
	return code, code.String()
}

func argToLabel(key string, value driver.Value) attribute.KeyValue {
	return attribute.Any(fmt.Sprintf("sql.arg.%s", key), value)
}
