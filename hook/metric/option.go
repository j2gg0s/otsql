package metric

import "github.com/prometheus/client_golang/prometheus"

type Option func(*Options)

// Options
type Options struct {
	// Regiterer is prometheus Registerer, default prometheus.DefaultRegisterer
	Registerer prometheus.Registerer

	// Latency histogram, default DefaultLatency
	Latency *prometheus.HistogramVec
}

func newOptions(opts []Option) *Options {
	o := &Options{
		Registerer: prometheus.DefaultRegisterer,
		Latency:    DefaultLatency,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithRegisterer
func WithRegisterer(registerer prometheus.Registerer) Option {
	return func(o *Options) {
		o.Registerer = registerer
	}
}

// WithLatency
func WithLatency(latency *prometheus.HistogramVec) Option {
	return func(o *Options) {
		o.Latency = latency
	}
}

var (
	sqlInstance = "sql_instance"
	sqlDatabase = "sql_database"
	sqlMethod   = "sql_method"
	sqlStatus   = "sql_status"

	DefaultLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "go_sql_latency",
			Help: "The latency of sql calls in milliseconds.",
		},
		[]string{sqlInstance, sqlDatabase, sqlMethod, sqlStatus},
	)
)
