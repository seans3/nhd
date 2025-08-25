package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
)

// statusCodeRecorder is a wrapper around http.ResponseWriter to capture the status code.
type statusCodeRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing it.
func (rec *statusCodeRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

// MetricsHandler holds the metrics data and provides the middleware and handler.
type MetricsHandler struct {
	mu          sync.RWMutex
	statusCodes map[int]int
}

// NewMetricsHandler creates a new MetricsHandler.
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		statusCodes: make(map[int]int),
	}
}

// Middleware is the middleware function to record metrics.
func (mh *MetricsHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't record metrics for the metrics endpoint itself.
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		recorder := &statusCodeRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)

		mh.mu.Lock()
		defer mh.mu.Unlock()
		mh.statusCodes[recorder.statusCode]++
	})
}

// Handler is the HTTP handler to serve the metrics.
func (mh *MetricsHandler) Handler(w http.ResponseWriter, r *http.Request) {
	mh.mu.RLock()
	defer mh.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(mh.statusCodes); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
	}
}
