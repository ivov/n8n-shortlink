package api

import (
	"crypto/rand"
	"encoding/hex"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/ivov/n8n-shortlink/internal/log"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

// ------------
//    setup
// ------------

// MiddlewareFn is a function that wraps an http.Handler.
type MiddlewareFn func(http.Handler) http.Handler

func createStack(fns ...MiddlewareFn) MiddlewareFn {
	return func(next http.Handler) http.Handler {
		for _, fn := range fns {
			next = fn(next)
		}
		return next
	}
}

// SetupMiddleware sets up all middleware on the API.
func (api *API) SetupMiddleware() MiddlewareFn {
	return createStack(
		api.setRequestID,
		api.logRequest,
		api.recoverPanic,
		api.addRelaxedCorsHeaders,
		api.addSecurityHeaders,
		api.addCacheHeadersForStaticFiles,
		api.rateLimit,
		api.metrics,
	)
}

// ------------
//   headers
// ------------

func (api *API) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = generateRequestID()
			r.Header.Set("X-Request-ID", id)
		}
		next.ServeHTTP(w, r)
	})
}

func generateRequestID() string {
	bytes := make([]byte, 16) // 128 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "unknown"
	}
	return hex.EncodeToString(bytes)
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

func (api *API) addSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")

		next.ServeHTTP(w, r)
	})
}

func (api *API) addCacheHeadersForStaticFiles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			w.Header().Set("Cache-Control", "public, max-age=2592000") // 1 month
		}
		next.ServeHTTP(w, r)
	})
}

// ------------
//   metrics
// ------------

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

// ------------
//   recover
// ------------

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

// ------------
//   logging
// ------------

type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	wroteHeader  bool // prevent metrics middleware (using httpsnoop) from calling WriteHeader multiple times on the same response writer
}

func (rw *wrappedResponseWriter) WriteHeader(statusCode int) {
	if !rw.wroteHeader {
		rw.statusCode = statusCode
		rw.ResponseWriter.WriteHeader(statusCode)
		rw.wroteHeader = true
	}
}

func (rw *wrappedResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	bytes, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += bytes
	return bytes, err
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

			wrappedWriter := &wrappedResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

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
					Status:      wrappedWriter.statusCode,
					SizeBytes:   wrappedWriter.bytesWritten,
					RequestID:   r.Header.Get("X-Request-ID"),
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

// ------------
//   r-limit
// ------------

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
