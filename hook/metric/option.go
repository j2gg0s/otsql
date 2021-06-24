package metric

import "github.com/prometheus/client_golang/prometheus"

type Option func(*Options)

type Options struct {
	InstanceName string

	// DBStats, if set to true, hook will monito dbstats
	DBStats bool

	Registerer prometheus.Registerer

	Latency *prometheus.HistogramVec

	ConnInUse          *prometheus.GaugeVec
	ConnIdle           *prometheus.GaugeVec
	ConnWait           *prometheus.GaugeVec
	ConnIdleClosed     *prometheus.GaugeVec
	ConnLifetimeClosed *prometheus.GaugeVec
	ConnWaitMS         *prometheus.GaugeVec
}

func newOptions(opts []Option) *Options {
	o := &Options{
		InstanceName: "default",

		Registerer: prometheus.DefaultRegisterer,

		Latency: DefaultLatency,

		ConnInUse:          DefaultConnInUse,
		ConnIdle:           DefaultConnIdle,
		ConnWait:           DefaultConnWait,
		ConnIdleClosed:     DefaultConnIdleClosed,
		ConnLifetimeClosed: DefaultConnLifetimeClosed,
		ConnWaitMS:         DefaultConnWaitMS,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

var (
	DefaultLatency *prometheus.HistogramVec

	DefaultConnInUse *prometheus.GaugeVec
	DefaultConnIdle  *prometheus.GaugeVec

	DefaultConnWait           *prometheus.GaugeVec
	DefaultConnIdleClosed     *prometheus.GaugeVec
	DefaultConnLifetimeClosed *prometheus.GaugeVec

	DefaultConnWaitMS *prometheus.GaugeVec
)

func init() {
	DefaultLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "go_sql_latency",
			Help: "The latency of sql calls in milliseconds.",
		},
		[]string{"sql_instance", "sql_method", "sql_status"},
	)

	DefaultConnInUse = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_in_use",
			Help: "The number of connections currently in use",
		},
		[]string{"sql_instance"},
	)
	DefaultConnIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_idle",
			Help: "The number of idle connections",
		},
		[]string{"sql_instance"},
	)
	DefaultConnWait = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_wait",
			Help: "The total number of connections wait for",
		},
		[]string{"sql_instance"},
	)
	DefaultConnIdleClosed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_idle_closed",
			Help: "The total number of connections closed because of SetMaxIdleConns",
		},
		[]string{"sql_instance"},
	)
	DefaultConnLifetimeClosed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_lifetime_closed",
			Help: "The total number of connections closed becase of SetConnMaxLifetime",
		},
		[]string{"sql_instance"},
	)
	DefaultConnWaitMS = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_wait_ms",
			Help: "The total time blocked by waiting for a new connection, millisecond.",
		},
		[]string{"sql_instance"},
	)
}
