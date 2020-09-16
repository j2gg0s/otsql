package otsql

import (
	"context"

	"go.opentelemetry.io/otel/label"
)

// TraceOption allows for managing otsql configuration using functional options.
type TraceOption func(*TraceOptions)

// TraceOptions holds configuration of our otsql tracing middleware.
// By default all options are set to false intentionally when creating a wrapped
// driver and provide the most sensible default with both performance and
// security in mind.
type TraceOptions struct {
	// AllowRoot, if set to true, will allow otsql to create root spans in
	// absence of existing spans or even context.
	// Default is to not trace otsql calls if no existing parent span is found
	// in context or when using methods not taking context.
	AllowRoot bool

	// Ping, if set to true, will enable the creation of spans on Ping requests.
	Ping bool

	// Query, if set to true, will enable recording of sql queries in spans.
	// Only allow this if it is safe to have queries recorded with respect to
	// security.
	Query bool

	// QueryParams, if set to true, will enable recording of parameters used
	// with parametrized queries. Only allow this if it is safe to have
	// parameters recorded with respect to security.
	// This setting is a noop if the Query option is set to false.
	QueryParams bool

	// RowsAffected, if set to true, will enable the creation of spans on
	// RowsAffected calls.
	RowsAffected bool

	// LastInsertID, if set to true, will enable the creation of spans on
	// LastInsertId calls.
	LastInsertID bool

	// RowsNext, if set to true, will enable the creation of spans on RowsNext
	// calls. This can result in many spans.
	RowsNext bool

	// RowsClose, if set to true, will enable the creation of spans on RowsClose
	// calls.
	RowsClose bool

	// SpanNameFormatter will be called to produce span's name.
	// Default use method as span name
	SpanNameFormatter func(ctx context.Context, method string, query string) string

	// DefaultLabels will be set to each span as default.
	DefaultLabels []label.KeyValue

	// InstanceName identifies database.
	InstanceName string
}

func newTraceOptions(options ...TraceOption) TraceOptions {
	o := TraceOptions{}
	for _, option := range options {
		option(&o)
	}

	if o.InstanceName == "" {
		o.InstanceName = "default"
	} else {
		o.DefaultLabels = append(o.DefaultLabels, label.String(sqlInstance, o.InstanceName))
	}

	if o.SpanNameFormatter == nil {
		o.SpanNameFormatter = func(_ context.Context, method string, _ string) string { return method }
	}

	if o.QueryParams && !o.Query {
		o.QueryParams = false
	}
	return o
}

// WithOptions sets our otsql tracing middleware options through a single
// TraceOptions object.
func WithOptions(options TraceOptions) TraceOption {
	return func(o *TraceOptions) {
		*o = options
		o.DefaultLabels = append(
			[]label.KeyValue(nil), options.DefaultLabels...,
		)
	}
}

// WithAllowRoot if set to true, will allow otsql to create root spans in
// absence of exisiting spans or even context.
// Default is to not trace otsql calls if no existing parent span is found
// in context or when using methods not taking context.
func WithAllowRoot(b bool) TraceOption {
	return func(o *TraceOptions) {
		o.AllowRoot = b
	}
}

// WithPing if set to true, will enable the creation of spans on Ping requests.
func WithPing(b bool) TraceOption {
	return func(o *TraceOptions) {
		o.Ping = b
	}
}

// WithRowsNext if set to true, will enable the creation of spans on RowsNext
// calls. This can result in many spans.
func WithRowsNext(b bool) TraceOption {
	return func(o *TraceOptions) {
		o.RowsNext = b
	}
}

// WithRowsClose if set to true, will enable the creation of spans on RowsClose
// calls.
func WithRowsClose(b bool) TraceOption {
	return func(o *TraceOptions) {
		o.RowsClose = b
	}
}

// WithRowsAffected if set to true, will enable the creation of spans on
// RowsAffected calls.
func WithRowsAffected(b bool) TraceOption {
	return func(o *TraceOptions) {
		o.RowsAffected = b
	}
}

// WithLastInsertID if set to true, will enable the creation of spans on
// LastInsertId calls.
func WithLastInsertID(b bool) TraceOption {
	return func(o *TraceOptions) {
		o.LastInsertID = b
	}
}

// WithQuery if set to true, will enable recording of sql queries in spans.
// Only allow this if it is safe to have queries recorded with respect to
// security.
func WithQuery(b bool) TraceOption {
	return func(o *TraceOptions) {
		o.Query = b
	}
}

// WithQueryParams if set to true, will enable recording of parameters used
// with parametrized queries. Only allow this if it is safe to have
// parameters recorded with respect to security.
// This setting is a noop if the Query option is set to false.
func WithQueryParams(b bool) TraceOption {
	return func(o *TraceOptions) {
		o.QueryParams = b
	}
}

// WithDefaultLabels will be set to each span as default.
func WithDefaultLabels(attrs ...label.KeyValue) TraceOption {
	return func(o *TraceOptions) {
		o.DefaultLabels = attrs
	}
}

// WithInstanceName sets database instance name.
func WithInstanceName(instanceName string) TraceOption {
	return func(o *TraceOptions) {
		o.InstanceName = instanceName
	}
}

// WithSpanNameFormatter sets name for each span.
func WithSpanNameFormatter(formatter func(context.Context, string, string) string) TraceOption {
	return func(o *TraceOptions) {
		o.SpanNameFormatter = formatter
	}
}
