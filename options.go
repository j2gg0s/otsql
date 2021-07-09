package otsql

// Option allows for managing otsql configuration using functional options.
type Option func(*Options)

// Options holds configuration of our otsql hook.
// By default all options are set to false intentionally when creating a wrapped
// driver and provide the most sensible default with both performance and
// security in mind.
type Options struct {
	// Instance, default parse from dsn.
	Instance string

	// Database, default parse from dsn.
	Database string

	// PingB, if set to true, will enable the hook of Ping requests.
	PingB bool

	// RowsAffectedB, if set to true, will enable the hook of RowsAffected calls.
	RowsAffectedB bool

	// LastInsertIdB, if set to true, will enable the hook LastInsertId calls.
	LastInsertIdB bool

	// RowsNextB, if set to true, will enable the hook of calls.
	// This can result in many calls.
	RowsNextB bool

	// RowsCloseB, if set to true, will enable the hook of RowsClose calls.
	RowsCloseB bool

	// ResetSessionB, if set to true, will enable the hook of ResetSession calls.
	ResetSessionB bool

	// Hooks, enabled hooks.
	Hooks []Hook
}

func newOptions(opts []Option) *Options {
	o := &Options{}
	for _, option := range opts {
		option(o)
	}
	return o
}

// WithOptions sets our otsql options through a single
// Options object.
func WithOptions(options Options) Option {
	return func(o *Options) {
		*o = options
	}
}

// WithInstance sets instance name, default parse from dsn when create conn.
func WithInstance(name string) Option {
	return func(o *Options) {
		o.Instance = name
	}
}

// WithDatabase sets instance name, default parse from dsn when create conn.
func WithDatabaase(name string) Option {
	return func(o *Options) {
		o.Database = name
	}
}

// WithPing if set to true, will enable the hook of Ping requests.
func WithPing(b bool) Option {
	return func(o *Options) {
		o.PingB = b
	}
}

// WithRowsNext if set to true, will enable of RowsNext calls.
// This can result in many calls.
func WithRowsNext(b bool) Option {
	return func(o *Options) {
		o.RowsNextB = b
	}
}

// WithRowsClose if set to true, will enable the of RowsClose calls.
func WithRowsClose(b bool) Option {
	return func(o *Options) {
		o.RowsCloseB = b
	}
}

// WithRowsAffected if set to true, will enable the of RowsAffected calls.
func WithRowsAffected(b bool) Option {
	return func(o *Options) {
		o.RowsAffectedB = b
	}
}

// WithLastInsertID if set to true, will enable the hook of LastInsertId calls.
func WithLastInsertID(b bool) Option {
	return func(o *Options) {
		o.LastInsertIdB = b
	}
}

// WithResetSession if set to true, will enable the hook of ResetSession calls.
func WithResetSession(b bool) Option {
	return func(o *Options) {
		o.ResetSessionB = b
	}
}

// WithHooks set hook.
func WithHooks(hooks ...Hook) Option {
	return func(o *Options) {
		o.Hooks = append(o.Hooks, hooks...)
	}
}
