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

var (
	connInUse metric.Int64ValueObserver
	connIdle  metric.Int64ValueObserver

	connWait           metric.Int64SumObserver
	connIdleClosed     metric.Int64SumObserver
	connLifetimeClosed metric.Int64SumObserver

	connWaitNS metric.Int64SumObserver
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

	batchObserver := meter.NewBatchObserver(func(ctx context.Context, result metric.BatchObserverResult) {
		dbstats := getDBStats()

		result.Observe(
			nil,
			connInUse.Observation(int64(dbstats.InUse)),
			connIdle.Observation(int64(dbstats.Idle)),
			connWait.Observation(int64(dbstats.WaitCount)),
			connIdleClosed.Observation(int64(dbstats.MaxIdleClosed)),
			connLifetimeClosed.Observation(int64(dbstats.MaxLifetimeClosed)),
			connWaitNS.Observation(int64(dbstats.WaitDuration)),
		)
	})

	formatter := func(s string) string { return s }
	if connInUse, err = batchObserver.NewInt64ValueObserver(
		formatter("go.sql.conn.in_use"),
		metric.WithDescription(fmt.Sprintf("The number of connections currently in use: %s", instanceName)),
	); err != nil {
		return err
	}

	if connIdle, err = batchObserver.NewInt64ValueObserver(
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

	if connWaitNS, err = batchObserver.NewInt64SumObserver(
		formatter("go.sql.conn.wait_ns"),
		metric.WithDescription("The total time blocked by waiting for a new connection, nanosecond."),
	); err != nil {
		return err
	}

	return nil
}
