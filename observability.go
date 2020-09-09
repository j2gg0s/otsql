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
	Meter                   = global.MeterProvider().Meter("github.com/j2gg0s/otsql")
	LatencyValueRecorder, _ = Meter.NewInt64ValueRecorder(
		"go.sql/latency",
		metric.WithDescription("The latency of calls in microsecond"),
	)
	ConnectionCounter, _ = Meter.NewInt64Counter(
		"go.sql/connections",
		metric.WithDescription("Count of connections in the pool"),
	)
)

const (
	SQLInstance = "sql.instance"
	SQLMethod   = "sql.method"
	SQLStatus   = "sql.status"
)

var (
	statusOK    = label.String(SQLStatus, "OK")
	statusError = label.String(SQLStatus, "Error")

	methodPing     = label.String(SQLMethod, "ping")
	methodExec     = label.String(SQLMethod, "exec")
	methodQuery    = label.String(SQLMethod, "query")
	methodPrepare  = label.String(SQLMethod, "preapre")
	methodBegin    = label.String(SQLMethod, "begin")
	methodCommit   = label.String(SQLMethod, "commit")
	methodRollback = label.String(SQLMethod, "rollback")

	methodLastInsertID = label.String(SQLMethod, "last_insert_id")
	methodRowsAffected = label.String(SQLMethod, "rows_affected")
	methodRowsClose    = label.String(SQLMethod, "rows_close")
	methodRowsNext     = label.String(SQLMethod, "rows_next")
)

func recordOp(ctx context.Context, method label.KeyValue, start time.Time, options TraceOptions) func(context.Context, error) {
	labels := []label.KeyValue{
		label.String(SQLInstance, options.InstanceName),
		method,
	}

	return func(ctx context.Context, err error) {
		if err != nil {
			labels = append(labels, statusOK)
		} else {
			labels = append(labels, statusError)
		}

		LatencyValueRecorder.Record(ctx, time.Since(start).Microseconds(), labels...)
	}
}

func startTrace(ctx context.Context, options TraceOptions, method label.KeyValue, query string, args interface{}) (context.Context, trace.Span, func(context.Context, error)) {
	start := time.Now()
	metricFunc := recordOp(ctx, method, start, options)
	if method == methodPing && !options.Ping {
		return ctx, nil, metricFunc
	}

	opts := []trace.StartOption{
		trace.WithSpanKind(trace.SpanKindClient),
	}
	if options.AllowRoot {
		opts = append(opts, trace.WithNewRoot())
	}
	attrs := []label.KeyValue{}
	if options.Query && len(query) > 0 {
		attrs = append(attrs, label.String("sql.query", query))
	}
	if len(options.DefaultAttributes) > 0 {
		attrs = append(attrs, options.DefaultAttributes...)
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
			// NOTE: 怎么处理
		}
	}

	if len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}
	ctx, span := tracer.Start(ctx, method.Value.AsString(), opts...)

	return ctx, span, func(ctx context.Context, err error) {
		metricFunc(ctx, err)
		if err != nil {
			span.RecordError(ctx, err)
		}
		code, msg := spanStatusFromSQLError(err)
		span.SetStatus(code, msg)
		span.End()
	}
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
