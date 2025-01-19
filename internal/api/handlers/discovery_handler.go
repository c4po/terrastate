package handlers

import (
	"encoding/json"
	"net/http"
)

type DiscoveryHandler struct{}

func NewDiscoveryHandler() *DiscoveryHandler {
	return &DiscoveryHandler{}
}

func (h *DiscoveryHandler) GetDiscovery(w http.ResponseWriter, r *http.Request) {
	discovery := map[string]interface{}{
		"modules.v1":  "/api/registry/v1/modules/",
		"state.v2":    "/api/v2/",
		"tfe.v2":      "/api/v2/",
		"tfe.v2.1":    "/api/v2/",
		"tfe.v2.2":    "/api/v2/",
		"versions.v1": "/api/versions/",
		"service-discovery": map[string]string{
			"providers.v1": "/api/v1/providers/",
			"state.v2":     "/api/v2/",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(discovery)
}
