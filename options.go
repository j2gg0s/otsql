package otsql

// Option allows for managing otsql configuration using functional options.
type Option func(*Options)

// Options holds configuration of our otsql tracing middleware.
// By default all options are set to false intentionally when creating a wrapped
// driver and provide the most sensible default with both performance and
// security in mind.
type Options struct {
	InstanceName string

	// AllowRoot, if set to true, will allow otsql to create root spans in
	// absence of existing spans or even context.
	// Default is to not trace otsql calls if no existing parent span is found
	// in context or when using methods not taking context.
	AllowRoot bool

	// PingB, if set to true, will enable the creation of spans on PingB requests.
	PingB bool

	// RowsAffectedB, if set to true, will enable the creation of spans on
	// RowsAffectedB calls.
	RowsAffectedB bool

	// LastInsertIdB, if set to true, will enable the creation of spans on
	// LastInsertIdB calls.
	LastInsertIdB bool

	// RowsNextB, if set to true, will enable the creation of spans on RowsNextB
	// calls. This can result in many spans.
	RowsNextB bool

	// RowsCloseB, if set to true, will enable the creation of spans on RowsCloseB
	// calls.
	RowsCloseB bool

	// Hooks
	Hooks []Hook
}

func newOptions(opts []Option) *Options {
	o := &Options{}
	for _, option := range opts {
		option(o)
	}
	return o
}

// WithOptions sets our otsql tracing middleware options through a single
// TraceOptions object.
func WithOptions(options Options) Option {
	return func(o *Options) {
		*o = options
	}
}

// WithPing if set to true, will enable the creation of spans on Ping requests.
func WithPing(b bool) Option {
	return func(o *Options) {
		o.PingB = b
	}
}

// WithRowsNext if set to true, will enable the creation of spans on RowsNext
// calls. This can result in many spans.
func WithRowsNext(b bool) Option {
	return func(o *Options) {
		o.RowsNextB = b
	}
}

// WithRowsClose if set to true, will enable the creation of spans on RowsClose
// calls.
func WithRowsClose(b bool) Option {
	return func(o *Options) {
		o.RowsCloseB = b
	}
}

// WithRowsAffected if set to true, will enable the creation of spans on
// RowsAffected calls.
func WithRowsAffected(b bool) Option {
	return func(o *Options) {
		o.RowsAffectedB = b
	}
}

// WithLastInsertID if set to true, will enable the creation of spans on
// LastInsertId calls.
func WithLastInsertID(b bool) Option {
	return func(o *Options) {
		o.LastInsertIdB = b
	}
}

// WithHooks
func WithHooks(hooks ...Hook) Option {
	return func(o *Options) {
		o.Hooks = append(o.Hooks, hooks...)
	}
}
