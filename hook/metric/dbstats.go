package metric

import (
	"context"
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Stats, monitor db connection pool with interval every.
func Stats(ctx context.Context, db *sql.DB, name string, every time.Duration) {
	ticker := time.NewTicker(every)
	for {
		select {
		case <-ticker.C:
			stats := db.Stats()

			labels := prometheus.Labels{
				sqlInstance: name,
			}

			ConnInUse.With(labels).Set(float64(stats.InUse))
			ConnIdle.With(labels).Set(float64(stats.Idle))

			ConnWait.With(labels).Set(float64(stats.WaitCount))
			ConnIdleClosed.With(labels).Set(float64(stats.MaxIdleClosed))
			ConnIdleTimeClosed.With(labels).Set(float64(stats.MaxIdleTimeClosed))
			ConnLifetimeClosed.With(labels).Set(float64(stats.MaxLifetimeClosed))
			ConnWaitMS.With(labels).Set(float64(stats.WaitDuration.Milliseconds()))
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

var (
	ConnInUse = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_in_use",
			Help: "The number of connections currently in use",
		},
		[]string{sqlInstance},
	)
	ConnIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_idle",
			Help: "The number of idle connections",
		},
		[]string{sqlInstance},
	)
	ConnWait = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_wait",
			Help: "The total number of connections wait for",
		},
		[]string{sqlInstance},
	)
	ConnIdleClosed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_idle_closed",
			Help: "The total number of connections closed because of SetMaxIdleConns",
		},
		[]string{sqlInstance},
	)
	ConnIdleTimeClosed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_idle_time_closed",
			Help: "The total number of connections closed because of SetConnMaxIdleTime",
		},
		[]string{sqlInstance},
	)
	ConnLifetimeClosed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_lifetime_closed",
			Help: "The total number of connections closed becase of SetConnMaxLifetime",
		},
		[]string{sqlInstance},
	)
	ConnWaitMS = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_sql_conn_wait_ms",
			Help: "The total time blocked by waiting for a new connection, millisecond.",
		},
		[]string{sqlInstance},
	)
)

func init() {
	prometheus.MustRegister(ConnInUse)
	prometheus.MustRegister(ConnIdle)
	prometheus.MustRegister(ConnWait)
	prometheus.MustRegister(ConnIdleClosed)
	prometheus.MustRegister(ConnLifetimeClosed)
	prometheus.MustRegister(ConnWaitMS)
}
