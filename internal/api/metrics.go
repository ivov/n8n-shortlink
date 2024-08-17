package api

import (
	"expvar"
	"runtime"
	"time"
)

var startTime = time.Now()

// InitMetrics initializes server-level expvar metrics, exposed via GET /debug/vars.
// Request-level expvar metrics are collected in middleware.
//
// Prometheus metrics are based on expvar metrics and exposed via GET /metrics.
// commit_sha, uptime_seconds, timestamp, goroutines are excluded from Prometheus metrics.
func (api *API) InitMetrics(commitSha string) {
	expvar.NewString("commit_sha").Set(commitSha)
	expvar.Publish("uptime_seconds", expvar.Func(func() interface{} {
		return int64(time.Since(startTime).Seconds())
	}))
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
}
