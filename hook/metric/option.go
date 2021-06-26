package metric

import "github.com/prometheus/client_golang/prometheus"

type Option func(*Options)

// Options
type Options struct {
	// InstanceName
	InstanceName string

	// Regiterer is prometheus Registerer, default prometheus.DefaultRegisterer
	Registerer prometheus.Registerer

	// Latency histogram, default DefaultLatency
	Latency *prometheus.HistogramVec
}

func newOptions(opts []Option) *Options {
	o := &Options{
		InstanceName: "default",

		Registerer: prometheus.DefaultRegisterer,

		Latency: DefaultLatency,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func WithInstanceName(instanceName string) Option {
	return func(o *Options) {
		o.InstanceName = instanceName
	}
}

func WithRegisterer(registerer prometheus.Registerer) Option {
	return func(o *Options) {
		o.Registerer = registerer
	}
}

func WithLatency(latency *prometheus.HistogramVec) Option {
	return func(o *Options) {
		o.Latency = latency
	}
}

var (
	DefaultLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "go_sql_latency",
			Help: "The latency of sql calls in milliseconds.",
		},
		[]string{"sql_instance", "sql_method", "sql_status"},
	)
)
