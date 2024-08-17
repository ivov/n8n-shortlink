package api

import (
	"expvar"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
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
	r := chi.NewRouter()

	api.SetupMiddleware(r)

	static := internal.Static()

	fileServer := http.FileServer(http.FS(static))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// /static/canvas.tmpl.html and /static/swagger.html blocked by reverse proxy

	r.Get("/health", api.HandleGetHealth)
	r.Get("/debug/vars", expvar.Handler().ServeHTTP) // blocked by reverse proxy
	r.Get("/metrics", api.HandleGetMetrics)          // blocked by reverse proxy
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, static, "swagger.html")
	})
	r.Get("/spec", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yml")
	})

	r.Post("/shortlink", api.HandlePostShortlink)
	r.Get("/{slug}/view", api.HandleGetSlug)
	r.Get("/{slug}", api.HandleGetSlug)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, static, "index.html")
	})

	return r
}
