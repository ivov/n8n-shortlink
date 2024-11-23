package api

import (
	"encoding/json"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/ivov/n8n-shortlink/internal/errors"
	"github.com/ivov/n8n-shortlink/internal/log"
)

func (api *API) jsonResponse(w http.ResponseWriter, status int, payload interface{}) {
	json, err := json.Marshal(payload)
	if err != nil {
		api.InternalServerError(err, w)
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(json); err != nil {
		api.Logger.Error(err)
	}
}

// CreatedSuccesfully responds with a 201.
func (api *API) CreatedSuccesfully(w http.ResponseWriter, payload interface{}) {
	api.jsonResponse(w, http.StatusCreated, SuccessResponse{Data: payload})
}

// InternalServerError responds with a 500.
func (api *API) InternalServerError(err error, w http.ResponseWriter) {
	api.Logger.Error(err)

	sentry.CaptureException(err)

	errorResponse := ErrorResponse{
		Error: ErrorField{
			Message: "The server encountered an unexpected issue. Please contact the administrator.",
			Code:    "INTERNAL_SERVER_ERROR",
			Doc:     "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/500",
			Trace:   err.Error(),
		},
	}

	api.jsonResponse(w, http.StatusInternalServerError, errorResponse)
}

// BadRequest responds with a 400.
func (api *API) BadRequest(err error, w http.ResponseWriter) {
	api.Logger.Info(err.Error())

	errorResponse := ErrorResponse{
		Error: ErrorField{
			Message: "Your request is invalid. Please correct the request and retry.",
			Code:    errors.ToCode[err],
			Doc:     "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
			Trace:   err.Error(),
		},
	}

	api.jsonResponse(w, http.StatusBadRequest, errorResponse)
}

// NotFound responds with a 404.
func (api *API) NotFound(w http.ResponseWriter) {
	errorResponse := ErrorResponse{
		Error: ErrorField{
			Message: "The requested resource could not be found.",
			Doc:     "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/404",
			Trace:   "n/a",
		},
	}

	api.jsonResponse(w, http.StatusNotFound, errorResponse)
}

// RateLimitExceeded responds with a 429.
func (api *API) RateLimitExceeded(w http.ResponseWriter, ip string) {
	api.Logger.Info("client exceeded rate limit", log.Str("ip", ip))

	errorResponse := ErrorResponse{
		Error: ErrorField{
			Message: "You have exceeded the rate limit. Please wait and retry later.",
			Doc:     "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/429",
			Trace:   "n/a",
		},
	}

	api.jsonResponse(w, http.StatusTooManyRequests, errorResponse)
}

// Unauthorized responds with a 401.
func (api *API) Unauthorized(err error, w http.ResponseWriter) {
	payload := ErrorResponse{
		Error: ErrorField{
			Message: "Missing valid authentication credentials.",
			Code:    errors.ToCode[err],
			Doc:     "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/401",
			Trace:   "n/a",
		},
	}

	api.jsonResponse(w, http.StatusUnauthorized, payload)
}

// ErrorResponse is a generic error response.
type ErrorResponse struct {
	Error ErrorField `json:"error"`
}

// SuccessResponse is a generic success response.
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// ErrorField contains error details.
type ErrorField struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Doc     string `json:"doc"`
	Trace   string `json:"trace"`
}
