package trace

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

// Option allows for managing trace configuration using functional options.
type Option func(*Options)

// Options holds configuration of our tracing hook.
// By default all options are set to false intentionally when creating a wrapped
// driver and provide the most sensible default with both performance and
// security in mind.
type Options struct {
	// AllowRoot, if set to true, will allow hook to create root spans in
	// absence of existing spans or even context.
	// Default is to not trace calls if no existing parent span is found
	// in context or when using methods not taking context.
	AllowRoot bool

	// Query, if set to true, will enable recording of sql queries in spans.
	// Only allow this if it is safe to have queries recorded with respect to
	// security.
	Query bool

	// QueryParams, if set to true, will enable recording of parameters used
	// with parametrized queries. Only allow this if it is safe to have
	// parameters recorded with respect to security.
	// This setting is a noop if the Query option is set to false.
	QueryParams bool

	// Ping, if set to true, will enable the creation of spans on Ping requests.
	Ping bool

	// RowsAffected, if set to true, will enable the creation of spans on
	// RowsAffected calls.
	RowsAffected bool

	// LastInsertId, if set to true, will enable the creation of spans on
	// LastInsertId calls.
	LastInsertId bool

	// RowsNext, if set to true, will enable the creation of spans on RowsNext
	// calls. This can result in many spans.
	RowsNext bool

	// RowsClose, if set to true, will enable the creation of spans on RowsClose
	// calls.
	RowsClose bool

	// ResetSession, if set to true, will enable the creation of spans on ResetSession
	// calls
	ResetSession bool

	// SpanNameFormatter will be called to produce span's name.
	// Default use method as span name
	SpanNameFormatter func(ctx context.Context, method string, query string) string

	// DefaultAttributes will be set to each span as default.
	DefaultAttributes []attribute.KeyValue

	// InstanceName identifies database.
	InstanceName string
}

func newOptions(opts []Option) *Options {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}

	if o.SpanNameFormatter == nil {
		o.SpanNameFormatter = func(ctx context.Context, method string, query string) string {
			return method
		}
	}

	if o.QueryParams && !o.Query {
		o.QueryParams = false
	}

	return o
}

// WithOptions sets our hook tracing middleware options through a single
// Options object.
func WithOptions(options Options) Option {
	return func(o *Options) {
		*o = options
		o.DefaultAttributes = append(
			[]attribute.KeyValue(nil), options.DefaultAttributes...,
		)
	}
}

// WithAllowRoot if set to true, will allow hook to create root spans in
// absence of exisiting spans or even context.
// Default is to not trace sql calls if no existing parent span is found
// in context or when using methods not taking context.
func WithAllowRoot(b bool) Option {
	return func(o *Options) {
		o.AllowRoot = b
	}
}

// WithPing if set to true, will enable the creation of spans on Ping requests.
func WithPing(b bool) Option {
	return func(o *Options) {
		o.Ping = b
	}
}

// WithRowsNext if set to true, will enable the creation of spans on RowsNext
// calls. This can result in many spans.
func WithRowsNext(b bool) Option {
	return func(o *Options) {
		o.RowsNext = b
	}
}

// WithRowsClose if set to true, will enable the creation of spans on RowsClose
// calls.
func WithRowsClose(b bool) Option {
	return func(o *Options) {
		o.RowsClose = b
	}
}

// WithRowsAffected if set to true, will enable the creation of spans on
// RowsAffected calls.
func WithRowsAffected(b bool) Option {
	return func(o *Options) {
		o.RowsAffected = b
	}
}

// WithLastInsertID if set to true, will enable the creation of spans on
// LastInsertId calls.
func WithLastInsertId(b bool) Option {
	return func(o *Options) {
		o.LastInsertId = b
	}
}

// WithResetSession if seto to true, will enable the creation of spans on
// ResetSession calls.
func WithResetRession(b bool) Option {
	return func(o *Options) {
		o.ResetSession = b
	}
}

// WithQuery if set to true, will enable recording of sql queries in spans.
// Only allow this if it is safe to have queries recorded with respect to
// security.
func WithQuery(b bool) Option {
	return func(o *Options) {
		o.Query = b
	}
}

// WithQueryParams if set to true, will enable recording of parameters used
// with parametrized queries. Only allow this if it is safe to have
// parameters recorded with respect to security.
// This setting is a noop if the Query option is set to false.
func WithQueryParams(b bool) Option {
	return func(o *Options) {
		o.QueryParams = b
	}
}

// WithDefaultAttributes will be set to each span as default.
func WithDefaultAttributes(attrs ...attribute.KeyValue) Option {
	return func(o *Options) {
		o.DefaultAttributes = attrs
	}
}

// WithInstanceName sets database instance name.
func WithInstanceName(instanceName string) Option {
	return func(o *Options) {
		o.InstanceName = instanceName
	}
}

// WithSpanNameFormatter sets name for each span.
func WithSpanNameFormatter(formatter func(context.Context, string, string) string) Option {
	return func(o *Options) {
		o.SpanNameFormatter = formatter
	}
}
