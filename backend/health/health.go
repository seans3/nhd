package health

import (
	"net/http"

	"github.com/seans3/nhd/backend/interfaces"
)

// HealthzHandler is a simple liveness probe.
// It returns 200 OK if the server is running.
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// ReadyzHandler is a readiness probe that checks dependencies.
type ReadyzHandler struct {
	DS interfaces.Datastore
}

// ServeHTTP checks if the datastore client is available.
func (h *ReadyzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// In a real application, you might perform a more thorough check,
	// like pinging the database. For now, we'll just check if the client exists.
	if h.DS == nil {
		http.Error(w, "datastore is not available", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
