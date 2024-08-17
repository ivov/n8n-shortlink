package api

import (
	"fmt"
	"net/http"
)

// HandleGetHealth returns the health status of the API.
func (api *API) HandleGetHealth(w http.ResponseWriter, r *http.Request) {
	health := fmt.Sprintf(
		`{"status": "ok", "environment": %q, "version": %q}`,
		api.Config.Env,
		api.Config.Build.CommitSha,
	)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(health))
}
