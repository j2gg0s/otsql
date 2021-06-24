package metric

import (
	"database/sql"
	"time"
)

func Stats(hook *Hook, db *sql.DB, name string, every time.Duration) {
	// egg or chicken, which is first?
	ticker := time.NewTicker(every)
	for range ticker.C {
		stats := db.Stats()

		hook.ConnInUse.WithLabelValues(name).Set(float64(stats.InUse))
		hook.ConnIdle.WithLabelValues(name).Set(float64(stats.Idle))

		hook.ConnWait.WithLabelValues(name).Set(float64(stats.WaitCount))
		hook.ConnIdleClosed.WithLabelValues(name).Set(float64(stats.MaxIdleClosed))
		hook.ConnLifetimeClosed.WithLabelValues(name).Set(float64(stats.MaxLifetimeClosed))
		hook.ConnWaitMS.WithLabelValues(name).Set(float64(stats.WaitDuration.Milliseconds()))
	}
}
