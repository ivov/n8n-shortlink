package api

import (
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ivov/n8n-shortlink/internal/db/entities"
	"github.com/ivov/n8n-shortlink/internal/errors"
	"github.com/tomasen/realip"
)

// HandlePostShortlink handles a POST /shortlink request by creating a shortlink.
func (api *API) HandlePostShortlink(w http.ResponseWriter, r *http.Request) {

	// check size limit

	const maxPayloadSize = 5 * 1024 * 1024 // 5 MB

	if r.ContentLength > maxPayloadSize {
		api.BadRequest(errors.ErrPayloadTooLarge, w)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadSize)

	var candidate entities.Shortlink

	err := json.NewDecoder(r.Body).Decode(&candidate)
	if err != nil {
		if err.Error() == "http: request body too large" {
			api.BadRequest(errors.ErrPayloadTooLarge, w)
		} else {
			api.InternalServerError(fmt.Errorf("failed to decode JSON: %w", err), w)
		}
		return
	}

	// check kind

	_, err = url.ParseRequestURI(candidate.Content)
	if err == nil {
		candidate.Kind = "url"
	} else {
		var unmarshaled interface{}
		err = json.Unmarshal([]byte(candidate.Content), &unmarshaled)
		if err == nil {
			candidate.Kind = "workflow"
		} else {
			api.BadRequest(errors.ErrContentMalformed, w)
			return
		}
	}

	if err := api.ShortlinkService.ValidateKind(candidate.Kind); err != nil {
		api.BadRequest(errors.ErrKindUnsupported, w)
		return
	}

	// check or generate slug

	if candidate.Slug != "" {
		err = api.ShortlinkService.ValidateUserSlug(candidate.Slug)
		if err != nil {
			switch {
			case stdErrors.Is(err, errors.ErrSlugTooShort):
				api.BadRequest(errors.ErrSlugTooShort, w)
			case stdErrors.Is(err, errors.ErrSlugTooLong):
				api.BadRequest(errors.ErrSlugTooLong, w)
			case stdErrors.Is(err, errors.ErrSlugMisformatted):
				api.BadRequest(errors.ErrSlugMisformatted, w)
			case stdErrors.Is(err, errors.ErrSlugTaken):
				api.BadRequest(errors.ErrSlugTaken, w)
			case stdErrors.Is(err, errors.ErrSlugReserved):
				api.BadRequest(errors.ErrSlugReserved, w)
			default:
				api.InternalServerError(err, w)
			}
			return
		}
	} else {
		generatedSlug, err := api.ShortlinkService.GenerateSlug()
		if err != nil {
			api.InternalServerError(err, w)
			return
		}
		candidate.Slug = generatedSlug
	}

	// check and hash password if provided

	if candidate.Password != "" {
		if err := api.ShortlinkService.ValidatePasswordLength(candidate.Password); err != nil {
			api.BadRequest(err, w)
			return
		}

		hash, err := api.ShortlinkService.HashPassword(candidate.Password)
		if err != nil {
			api.InternalServerError(err, w)
			return
		}
		candidate.Password = hash
	}

	candidate.AllowedVisits = -1 // unlimited, not yet implemented
	// candidate.ExpiresAt = &entities.CustomTime{Time: time.Now()} // not yet implemented

	candidate.CreatorIP = realip.FromRequest(r)

	shortlink, err := api.ShortlinkService.SaveShortlink(&candidate)
	if err != nil {
		api.InternalServerError(err, w)
		return
	}

	shortlink.Password = "" // do not return the password

	api.CreatedSuccesfully(w, shortlink)
}
