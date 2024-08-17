package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ivov/n8n-shortlink/internal"
	"github.com/ivov/n8n-shortlink/internal/db/entities"
	"github.com/ivov/n8n-shortlink/internal/errors"
	"github.com/ivov/n8n-shortlink/internal/log"
)

// HandleGetProtectedSlug handles a GET /p/{slug} request by resolving a password-protected shortlink.
func (api *API) HandleGetProtectedSlug(w http.ResponseWriter, r *http.Request, slug string, shortlink *entities.Shortlink) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		http.ServeFileFS(w, r, internal.Static(), "challenge.html")
		return
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		api.Unauthorized(errors.ErrAuthHeaderMalformed, "MALFORMED_AUTHORIZATION_HEADER", w)
		return
	}

	encodedPassword := strings.TrimPrefix(authHeader, "Basic ")
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedPassword)
	if err != nil {
		api.Unauthorized(errors.ErrAuthHeaderMalformed, "MALFORMED_AUTHORIZATION_HEADER", w)
		return
	}

	decodedPassword := string(decodedBytes)
	if !api.ShortlinkService.VerifyPassword(shortlink.Password, decodedPassword) {
		api.Unauthorized(errors.ErrPasswordInvalid, "INVALID_PASSWORD", w)
		return
	}

	api.Logger.Info("password verified", log.Str("slug", slug))

	err = api.VisitService.SaveVisit(slug, shortlink.Kind, r.Referer(), r.UserAgent())
	if err != nil {
		api.Logger.Error(err) // log and move on
	}

	switch shortlink.Kind {
	case "workflow":
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(shortlink.Content))
	case "url":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"url": shortlink.Content})
	default:
		api.BadRequest(errors.ErrKindUnsupported, w)
	}
}
