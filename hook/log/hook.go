package log

import (
	"context"
	"database/sql/driver"
	"errors"
	"time"

	"github.com/j2gg0s/otsql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Hook
type Hook struct {
	*Options
}

var _ otsql.Hook = (*Hook)(nil)

func (hook *Hook) Before(ctx context.Context, evt *otsql.Event) context.Context {
	return ctx
}

func (hook *Hook) After(ctx context.Context, evt *otsql.Event) {
	var e *zerolog.Event
	if evt.Err != nil && !errors.Is(evt.Err, driver.ErrSkip) {
		e = hook.Warn(ctx).Err(evt.Err)
	} else if time.Since(evt.BeginAt) > hook.Slow {
		e = hook.Warn(ctx).Bool("slow", true)
	} else {
		level, ok := hook.MethodLevels[evt.Method]
		if !ok {
			level = hook.DefaultLevel
		}
		if hook.GetLevel() <= level {
			e = hook.WithLevel(ctx, level)
		}
		if evt.Err != nil {
			e.Err(evt.Err)
		}
	}

	if e == nil {
		return
	}

	e = e.Str("kind", "sql")
	if evt.Instance != "" {
		e = e.Str("server", evt.Instance)
	}
	if evt.Conn != "" {
		e = e.Str("conn", evt.Conn)
	}
	if evt.Database != "" {
		e = e.Str("database", evt.Database)
	}
	if evt.Method != "" {
		e = e.Str("method", string(evt.Method))
	}
	e = e.Str("code", otsql.ErrToCode(evt.Err).String()).
		Dur("latency", time.Since(evt.BeginAt))

	if hook.Query && evt.Query != "" {
		e.Str("query", evt.Query)
		if hook.Args && evt.Args != nil {
			e.Interface("params", evt.Args)
		}
	}

	e.Msg("AccessLog")
}

func New(opts ...Option) *Hook {
	return &Hook{Options: newOptions(opts)}
}

// Option
type Option func(*Options)

type Options struct {
	Logger
	Slow time.Duration

	MethodLevels map[otsql.Method]zerolog.Level
	DefaultLevel zerolog.Level

	Query bool
	Args  bool
}

func newOptions(opts []Option) *Options {
	o := &Options{
		Logger: WrapZerolog(log.Logger),
		Slow:   time.Second * 3,

		MethodLevels: map[otsql.Method]zerolog.Level{
			otsql.MethodPing:     zerolog.DebugLevel,
			otsql.MethodQuery:    zerolog.DebugLevel,
			otsql.MethodPrepare:  zerolog.DebugLevel,
			otsql.MethodBegin:    zerolog.DebugLevel,
			otsql.MethodCommit:   zerolog.DebugLevel,
			otsql.MethodRollback: zerolog.DebugLevel,

			otsql.MethodLastInsertId: zerolog.DebugLevel,
			otsql.MethodRowsAffected: zerolog.DebugLevel,
			otsql.MethodRowsClose:    zerolog.DebugLevel,
			otsql.MethodRowsNext:     zerolog.DebugLevel,

			otsql.MethodExec:         zerolog.InfoLevel,
			otsql.MethodCreateConn:   zerolog.InfoLevel,
			otsql.MethodCloseConn:    zerolog.InfoLevel,
			otsql.MethodResetSession: zerolog.DebugLevel,
		},
		DefaultLevel: zerolog.InfoLevel,

		Query: true,
		Args:  false,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func WithLogger(logger Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

func WithSlow(d time.Duration) Option {
	return func(o *Options) {
		o.Slow = d
	}
}

func WithMethodLevel(method otsql.Method, level zerolog.Level) Option {
	return func(o *Options) {
		o.MethodLevels[method] = level
	}
}

func WithDefaultLevel(level zerolog.Level) Option {
	return func(o *Options) {
		o.DefaultLevel = level
	}
}

func WithQuery(b bool) Option {
	return func(o *Options) {
		o.Query = b
	}
}

func WithArgs(b bool) Option {
	return func(o *Options) {
		o.Args = b
	}
}

// Logger
type Logger interface {
	Debug(context.Context) *zerolog.Event
	Info(context.Context) *zerolog.Event
	Warn(context.Context) *zerolog.Event
	Error(context.Context) *zerolog.Event

	WithLevel(context.Context, zerolog.Level) *zerolog.Event
	GetLevel() zerolog.Level
}

type zerologger struct {
	zerolog.Logger
}

var _ Logger = (*zerologger)(nil)

func (logger *zerologger) Debug(ctx context.Context) *zerolog.Event {
	return logger.Logger.Debug()
}

func (logger *zerologger) Info(ctx context.Context) *zerolog.Event {
	return logger.Logger.Info()
}

func (logger *zerologger) Warn(ctx context.Context) *zerolog.Event {
	return logger.Logger.Warn()
}

func (logger *zerologger) Error(ctx context.Context) *zerolog.Event {
	return logger.Logger.Error()
}

func (logger *zerologger) WithLevel(ctx context.Context, level zerolog.Level) *zerolog.Event {
	return logger.Logger.WithLevel(level)
}

func (logger *zerologger) GetLevel() zerolog.Level {
	return logger.Logger.GetLevel()
}

func WrapZerolog(logger zerolog.Logger) Logger {
	return &zerologger{logger}
}
