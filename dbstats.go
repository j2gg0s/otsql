package otsql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/api/metric"
)

var dbs = map[*sql.DB]string{}
var dbsLock sync.Mutex

// RecordStats records database statistics for provided sql.DB.
// The interval is controlled by meter, default more than 10 seconds.
func RecordStats(db *sql.DB, instanceName string) (err error) {
	dbsLock.Lock()
	defer dbsLock.Unlock()
	defer func() {
		if err == nil {
			dbs[db] = instanceName
		}
	}()

	if len(dbs) > 0 {
		return nil
	}

	var stats sql.DBStats
	last := time.Time{}

	getDBStats := func() sql.DBStats {
		now := time.Now()
		if now.Sub(last) > time.Second {
			stats = db.Stats()
			last = now
		}
		return stats
	}

	formatter := func(s string) string { return s }
	if _, err = meter.NewInt64UpDownSumObserver(
		formatter("go.sql.conn.in_use"),
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(int64(getDBStats().InUse))
		},
		metric.WithDescription(fmt.Sprintf("The number of connections currently in use: %s", instanceName)),
	); err != nil {
		return err
	}

	if _, err = meter.NewInt64UpDownSumObserver(
		formatter("go.sql.conn.idle"),
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(int64(getDBStats().Idle))
		},
		metric.WithDescription(fmt.Sprintf("The number of idle connections: %s", instanceName)),
	); err != nil {
		return err
	}

	if _, err = meter.NewInt64SumObserver(
		formatter("go.sql.conn.wait"),
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(int64(getDBStats().WaitCount))
		},
		metric.WithDescription("The total number of connections wait for"),
	); err != nil {
		return err
	}

	if _, err = meter.NewInt64SumObserver(
		formatter("go.sql.conn.idle_closed"),
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(int64(getDBStats().MaxIdleClosed))
		},
		metric.WithDescription("The total number of connections closed because of SetMaxIdleConns"),
	); err != nil {
		return err
	}

	if _, err = meter.NewInt64SumObserver(
		formatter("go.sql.conn.lifetime_closed"),
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(int64(getDBStats().MaxLifetimeClosed))
		},
		metric.WithDescription("The total number of connections closed because of SetConnMaxLifetime"),
	); err != nil {
		return err
	}

	if _, err = meter.NewInt64SumObserver(
		formatter("go.sql.conn.wait_ns"),
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(int64(getDBStats().WaitDuration.Nanoseconds()))
		},
		metric.WithDescription("The total time blocked by waiting for a new connection, nanosecond."),
	); err != nil {
		return err
	}

	return nil
}
