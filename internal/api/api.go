package api

import (
	"expvar"
	"net/http"
	"strings"
	"sync"

	"github.com/ivov/n8n-shortlink/internal"
	"github.com/ivov/n8n-shortlink/internal/config"
	"github.com/ivov/n8n-shortlink/internal/log"
	"github.com/ivov/n8n-shortlink/internal/services"
)

// API encapsulates all routes and middleware.
type API struct {
	Config *config.Config
	Logger *log.Logger
	// WaitGroup tracks background tasks to wait for during server shutdown.
	WaitGroup        sync.WaitGroup
	ShortlinkService *services.ShortlinkService
	VisitService     *services.VisitService
}

// StaticFileHandler creates a handler serving static files with the correct MIME type
func StaticFileHandler(fileServer http.Handler, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		http.StripPrefix("/static/", fileServer).ServeHTTP(w, r)
	}
}

// Routes sets up middleware and routes on the API.
func (api *API) Routes() http.Handler {
	r := http.NewServeMux()
	static := internal.Static()
	fileServer := http.FileServer(http.FS(static))

	r.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, static, "index.html")
	})

	// /static/canvas.tmpl.html and /static/swagger.html blocked by reverse proxy

	r.HandleFunc("GET /health", api.HandleGetHealth)
	r.HandleFunc("GET /debug/vars", expvar.Handler().ServeHTTP) // blocked by reverse proxy
	r.HandleFunc("GET /metrics", api.HandleGetMetrics)          // blocked by reverse proxy
	r.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, static, "swagger.html")
	})
	r.HandleFunc("GET /spec", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yml")
	})

	r.HandleFunc("GET /static/js/{file}", StaticFileHandler(fileServer, "application/javascript; charset=utf-8"))
	r.HandleFunc("GET /static/styles/{file}", StaticFileHandler(fileServer, "text/css; charset=utf-8"))

	r.HandleFunc("GET /static/img/{file}", func(w http.ResponseWriter, r *http.Request) {
		contentType := "image/png" // default
		switch {
		case strings.HasSuffix(r.URL.Path, ".svg"):
			contentType = "image/svg+xml"
		case strings.HasSuffix(r.URL.Path, ".ico"):
			contentType = "image/x-icon"
		}
		StaticFileHandler(fileServer, contentType)(w, r)
	})

	r.HandleFunc("GET /static/fonts/{file}", func(w http.ResponseWriter, r *http.Request) {
		contentType := "font/woff2" // default
		if strings.HasSuffix(r.URL.Path, ".ttf") {
			contentType = "font/ttf"
		}
		StaticFileHandler(fileServer, contentType)(w, r)
	})

	r.HandleFunc("POST /shortlink", api.HandlePostShortlink)
	r.HandleFunc("GET /{slug}/view", api.HandleGetSlug)
	r.HandleFunc("GET /{slug}", api.HandleGetSlug)

	mw := api.SetupMiddleware()

	return mw(r)
}
