package api

import (
	"encoding/json"
	"net/http"

	"github.com/satetsu888/agentrace/server/internal/version"
)

func HandleGetVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": version.Version,
	})
}
