package api

import (
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ivov/n8n-shortlink/internal/log"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

// SetupMiddleware sets up all middleware on the API.
func (api *API) SetupMiddleware(r *chi.Mux) {
	r.Use(api.recoverPanic)
	r.Use(middleware.RequestID)
	r.Use(middleware.CleanPath)
	r.Use(middleware.StripSlashes)
	r.Use(api.addRelaxedCorsHeaders)
	r.Use(api.addSecurityHeaders)
	r.Use(api.addCacheHeadersForStaticFiles)
	r.Use(api.rateLimit)
	r.Use(api.logRequest)
	r.Use(api.metrics)
}

func (api *API) addCacheHeadersForStaticFiles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			w.Header().Set("Cache-Control", "public, max-age=2592000") // 1 month
		}
		next.ServeHTTP(w, r)
	})
}

func (api *API) metrics(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMs := expvar.NewInt("total_processing_time_ms")
	totalResponsesSentbyStatus := expvar.NewMap("total_responses_sent_by_status")
	inFlightRequests := expvar.NewInt("in_flight_requests")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		totalRequestsReceived.Add(1)
		metrics := httpsnoop.CaptureMetrics(next, w, r)
		totalResponsesSent.Add(1)
		totalProcessingTimeMs.Add(metrics.Duration.Milliseconds())
		totalResponsesSentbyStatus.Add(strconv.Itoa(metrics.Code), 1)
		inFlightRequests.Set(totalRequestsReceived.Value() - totalResponsesSent.Value())
	})
}

func (api *API) addSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")

		next.ServeHTTP(w, r)
	})
}

func (api *API) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.Header().Set("Connection", "close")
					err := fmt.Errorf("%s", err)
					api.InternalServerError(err, w)
				}
			}()

			next.ServeHTTP(w, r)
		},
	)
}

func isLogIgnored(path string) bool {
	for _, denyExact := range []string{"/health", "/favicon.ico", "/metrics"} {
		if path == denyExact || strings.HasPrefix(path, "/static/") {
			return true
		}
	}
	return false
}

func (api *API) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if isLogIgnored(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			wrappedWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()

			defer func() {
				requestData := struct {
					Protocol    string `json:"protocol"`
					Method      string `json:"method"`
					Path        string `json:"path"`
					QueryString string `json:"query_string"`
					Latency     int    `json:"latency_ms"`
					Status      int    `json:"status"`
					SizeBytes   int    `json:"size_bytes"`
					RequestID   string `json:"request_id"`
				}{
					Protocol:    r.Proto,
					Method:      r.Method,
					Path:        r.URL.Path,
					QueryString: r.URL.Query().Encode(),
					Latency:     int(time.Duration(time.Since(start).Milliseconds())),
					Status:      wrappedWriter.Status(),
					SizeBytes:   wrappedWriter.BytesWritten(),
					RequestID:   middleware.GetReqID(r.Context()),
				}

				api.Logger.Info(
					"Served request",
					log.Str("protocol", requestData.Protocol),
					log.Str("method", requestData.Method),
					log.Str("path", requestData.Path),
					log.Str("query_string", requestData.QueryString),
					log.Int("latency_ms", requestData.Latency),
					log.Int("status", requestData.Status),
					log.Int("size_bytes", requestData.SizeBytes),
					log.Str("request_id", requestData.RequestID),
				)
			}()

			next.ServeHTTP(wrappedWriter, r)
		},
	)
}

func (api *API) addRelaxedCorsHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Referer, User-Agent")

			next.ServeHTTP(w, r)
		},
	)
}

var exemptedFromRateLimiting = []string{"/", "/docs"}

func isExemptedFromRateLimiting(path string) bool {
	if strings.HasPrefix(path, "/static/") {
		return true
	}

	for _, item := range exemptedFromRateLimiting {
		if path == item {
			return true
		}
	}

	return false
}

func (api *API) rateLimit(next http.Handler) http.Handler {
	type RateLimiterPerClient struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	type IPAddress = string

	var (
		mutex   sync.Mutex
		clients = make(map[IPAddress]*RateLimiterPerClient)
	)

	// every minute clear out inactive clients in the background
	go func() {
		for {
			time.Sleep(time.Minute)

			mutex.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > api.Config.RateLimiter.Inactivity {
					delete(clients, ip)
				}
			}

			mutex.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isExemptedFromRateLimiting(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		if api.Config.RateLimiter.Enabled {
			ip := realip.FromRequest(r)

			mutex.Lock()

			if _, ok := clients[ip]; !ok {
				clients[ip] = &RateLimiterPerClient{
					limiter: rate.NewLimiter(
						rate.Limit(api.Config.RateLimiter.RPS),
						api.Config.RateLimiter.Burst),
					lastSeen: time.Now(),
				}
			}

			if !clients[ip].limiter.Allow() {
				mutex.Unlock()
				api.RateLimitExceeded(w, ip)
				return
			}

			mutex.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}
