package api

import (
	"expvar"
	"net/http"
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

// Routes sets up middleware and routes on the API.
func (api *API) Routes() http.Handler {
	r := http.NewServeMux()

	static := internal.Static()

	fileServer := http.FileServer(http.FS(static))
	r.Handle("GET /static/*", http.StripPrefix("/static", fileServer))

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

	r.HandleFunc("POST /shortlink", api.HandlePostShortlink)
	r.HandleFunc("GET /{slug}/view", api.HandleGetSlug)
	r.HandleFunc("GET /{slug}", api.HandleGetSlug)

	r.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, static, "index.html")
	})

	mw := api.SetupMiddleware()

	return mw(r)
}
