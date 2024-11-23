package api

import (
	"fmt"
	"net/http"
)

// HandleGetHealth returns the health status of the API.
func (api *API) HandleGetHealth(w http.ResponseWriter, _ *http.Request) {
	health := fmt.Sprintf(
		`{"status": "ok", "environment": %q, "version": %q}`,
		api.Config.Env,
		api.Config.Build.CommitSha,
	)

	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write([]byte(health)); err != nil {
		api.Logger.Error(err)
	}
}
