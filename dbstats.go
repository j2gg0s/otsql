package otsql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/label"
)

var dbs = map[*sql.DB]string{}
var dbsLock sync.Mutex
var (
	batchObserver metric.BatchObserver

	connIdle  metric.Int64UpDownSumObserver
	connInUse metric.Int64UpDownSumObserver

	connWait           metric.Int64SumObserver
	connIdleClosed     metric.Int64SumObserver
	connLifetimeClosed metric.Int64SumObserver

	connWaitDurationNS metric.Int64SumObserver

	lastDBStats time.Time
	dbStats     sql.DBStats
)

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

	batchObserver = Meter.NewBatchObserver(func(ctx context.Context, result metric.BatchObserverResult) {
		now := time.Now()
		if now.Sub(lastDBStats) < time.Second {
			return
		}
		lastDBStats = now

		for db, instanceName := range dbs {
			dbStats = db.Stats()

			result.Observe(
				[]label.KeyValue{label.String(sqlInstance, instanceName)},

				connInUse.Observation(int64(dbStats.InUse)),
				connIdle.Observation(int64(dbStats.Idle)),

				connWait.Observation(dbStats.WaitCount),
				connIdleClosed.Observation(dbStats.MaxIdleClosed),
				connLifetimeClosed.Observation(dbStats.MaxLifetimeClosed),

				connWaitDurationNS.Observation(dbStats.WaitDuration.Nanoseconds()),
			)
		}
	})

	formatter := func(s string) string { return s }
	if connInUse, err = batchObserver.NewInt64UpDownSumObserver(
		formatter("go.sql.conn.in_use"),
		metric.WithDescription(fmt.Sprintf("The number of connections currently in use: %s", instanceName)),
	); err != nil {
		return err
	}

	if connIdle, err = batchObserver.NewInt64UpDownSumObserver(
		formatter("go.sql.conn.idle"),
		metric.WithDescription(fmt.Sprintf("The number of idle connections: %s", instanceName)),
	); err != nil {
		return err
	}

	if connWait, err = batchObserver.NewInt64SumObserver(
		formatter("go.sql.conn.wait"),
		metric.WithDescription("The total number of connections wait for"),
	); err != nil {
		return err
	}

	if connIdleClosed, err = batchObserver.NewInt64SumObserver(
		formatter("go.sql.conn.idle_closed"),
		metric.WithDescription("The total number of connections closed because of SetMaxIdleConns"),
	); err != nil {
		return err
	}

	if connLifetimeClosed, err = batchObserver.NewInt64SumObserver(
		formatter("go.sql.conn.lifetime_closed"),
		metric.WithDescription("The total number of connections closed because of SetConnMaxLifetime"),
	); err != nil {
		return err
	}

	if connWaitDurationNS, err = batchObserver.NewInt64SumObserver(
		formatter("go.sql.conn.wait_ns"),
		metric.WithDescription("The total time blocked by waiting for a new connection, nanosecond."),
	); err != nil {
		return err
	}

	return nil
}
