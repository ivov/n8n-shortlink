package api

import (
	"expvar"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HandleGetMetrics handles the /metrics endpoint exposing metrics for Prometheus.
func (api *API) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	updatePrometheusMetrics() // on every scrape

	promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	).ServeHTTP(w, r)
}

var (
	totalRequestsReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_requests_received",
		Help: "Total number of requests received",
	})
	totalResponsesSent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_responses_sent",
		Help: "Total number of responses sent",
	})
	totalProcessingTimeMs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_processing_time_ms",
		Help: "Total processing time in milliseconds",
	})
	inFlightRequests = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "Number of requests currently being processed",
	})
	responsesSentByStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "responses_sent_by_status",
			Help: "Total responses sent by HTTP status code",
		},
		[]string{"status"},
	)
)

var (
	lastTotalRequestsReceived float64
	lastTotalResponsesSent    float64
	lastTotalProcessingTimeMs float64
)

func init() {
	prometheus.MustRegister(totalRequestsReceived)
	prometheus.MustRegister(totalResponsesSent)
	prometheus.MustRegister(totalProcessingTimeMs)
	prometheus.MustRegister(inFlightRequests)
	prometheus.MustRegister(responsesSentByStatus)
}

func updatePrometheusMetrics() {
	expvar.Do(func(kv expvar.KeyValue) {
		switch kv.Key {
		case "total_requests_received":
			if v, err := strconv.ParseFloat(kv.Value.String(), 64); err == nil {
				totalRequestsReceived.Add(v)
			}
		case "total_responses_sent":
			if v, err := strconv.ParseFloat(kv.Value.String(), 64); err == nil {
				totalResponsesSent.Add(v)
			}
		case "total_processing_time_ms":
			if v, err := strconv.ParseFloat(kv.Value.String(), 64); err == nil {
				totalProcessingTimeMs.Add(v)
			}
		case "in_flight_requests":
			if v, err := strconv.ParseFloat(kv.Value.String(), 64); err == nil {
				inFlightRequests.Set(v)
			}
		case "total_responses_sent_by_status":
			m := kv.Value.(*expvar.Map)
			m.Do(func(kv expvar.KeyValue) {
				if v, err := strconv.ParseFloat(kv.Value.String(), 64); err == nil {
					responsesSentByStatus.WithLabelValues(kv.Key).Add(v)
				}
			})
		}
	})
}
