package otsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"
)

var instrumentationName = "github.com/j2gg0s/otsql"
var tracer = global.TraceProvider().Tracer(instrumentationName)

var (
	Meter = global.MeterProvider().Meter("github.com/j2gg0s/otsql")

	LatencyValueRecorder, _ = Meter.NewInt64ValueRecorder(
		"go.sql/latency",
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
	statusOK    = label.String(sqlStatus, "OK")
	statusError = label.String(sqlStatus, "Error")

	methodPing     = label.String(sqlMethod, "ping")
	methodExec     = label.String(sqlMethod, "exec")
	methodQuery    = label.String(sqlMethod, "query")
	methodPrepare  = label.String(sqlMethod, "preapre")
	methodBegin    = label.String(sqlMethod, "begin")
	methodCommit   = label.String(sqlMethod, "commit")
	methodRollback = label.String(sqlMethod, "rollback")

	methodLastInsertID = label.String(sqlMethod, "last_insert_id")
	methodRowsAffected = label.String(sqlMethod, "rows_affected")
	methodRowsClose    = label.String(sqlMethod, "rows_close")
	methodRowsNext     = label.String(sqlMethod, "rows_next")

	methodCreateConn = label.String(sqlMethod, "create_conn")
)

func startMetric(ctx context.Context, method label.KeyValue, start time.Time, options TraceOptions) func(context.Context, error) {
	labels := []label.KeyValue{
		label.String(sqlInstance, options.InstanceName),
		method,
	}

	return func(ctx context.Context, err error) {
		if err != nil {
			labels = append(labels, statusError)
		} else {
			labels = append(labels, statusOK)
		}

		LatencyValueRecorder.Record(ctx, time.Since(start).Microseconds(), labels...)
	}
}

func startTrace(ctx context.Context, options TraceOptions, method label.KeyValue, query string, args interface{}) (context.Context, trace.Span, func(context.Context, error)) {
	if !options.AllowRoot && !trace.SpanFromContext(ctx).SpanContext().IsValid() {
		return ctx, nil, func(context.Context, error) {}
	}
	if method == methodPing && !options.Ping {
		return ctx, nil, func(context.Context, error) {}
	}

	start := time.Now()
	endMetric := startMetric(ctx, method, start, options)

	opts := []trace.StartOption{
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
			span.RecordError(ctx, err)
		}
		code, msg := spanStatusFromSQLError(err)
		span.SetStatus(code, msg)
		span.End()
	}
}

func attrsFromSQL(ctx context.Context, options TraceOptions, method label.KeyValue, query string, args interface{}) []label.KeyValue {
	attrs := []label.KeyValue{}
	if len(options.DefaultLabels) > 0 {
		attrs = append(attrs, options.DefaultLabels...)
	}

	if options.Query && len(query) > 0 {
		attrs = append(attrs, label.String(sqlQuery, query))
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
			attrs = append(attrs, labelUnknownArgs)
		}
	}
	return attrs
}

func spanStatusFromSQLError(err error) (code codes.Code, msg string) {
	switch err {
	case nil:
		code = codes.OK
		return code, "Success"
	case driver.ErrSkip:
		code = codes.Unimplemented
	case context.Canceled:
		code = codes.Canceled
	case context.DeadlineExceeded:
		code = codes.DeadlineExceeded
	case sql.ErrNoRows:
		code = codes.NotFound
	case sql.ErrTxDone:
		code = codes.FailedPrecondition
	default:
		code = codes.Unknown
	}
	return code, fmt.Sprintf("Error: %v", err)
}

func argToLabel(key string, value driver.Value) label.KeyValue {
	return label.Any(fmt.Sprintf("sql.arg.%s", key), value)
}
