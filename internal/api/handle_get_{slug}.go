package api

import (
	stdErrors "errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/ivov/n8n-shortlink/internal"
	"github.com/ivov/n8n-shortlink/internal/errors"
)

// HandleGetSlug handles a GET /{slug} request by resolving a regular shortlink.
func (api *API) HandleGetSlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	shortlink, err := api.ShortlinkService.GetBySlug(slug)
	if err != nil {
		if stdErrors.Is(err, errors.ErrShortlinkNotFound) {
			w.WriteHeader(http.StatusNotFound)
			http.ServeFileFS(w, r, internal.Static(), "404.html")
		} else {
			api.InternalServerError(err, w)
		}
		return
	}

	if shortlink.Password != "" {
		api.HandleGetProtectedSlug(w, r, slug, shortlink)
		return
	}

	err = api.VisitService.SaveVisit(slug, shortlink.Kind, r.Referer(), r.UserAgent())
	if err != nil {
		api.Logger.Error(err) // log and move on
	}

	switch shortlink.Kind {
	case "workflow":
		if strings.HasSuffix(r.URL.Path, "/view") {
			tmpl, err := template.ParseFS(internal.Static(), "canvas.tmpl.html")
			if err != nil {
				api.InternalServerError(err, w)
				return
			}

			data := struct {
				Workflow     string
				WorkflowSlug string
			}{
				Workflow:     shortlink.Content,
				WorkflowSlug: slug,
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			err = tmpl.Execute(w, data)
			if err != nil {
				api.InternalServerError(err, w)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(shortlink.Content)); err != nil {
			api.Logger.Error(err)
		}
	case "url":
		http.Redirect(w, r, shortlink.Content, http.StatusMovedPermanently)
	default:
		api.BadRequest(errors.ErrKindUnsupported, w)
	}
}
