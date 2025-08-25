package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging wraps an HTTP handler to log request details.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("--> %s %s", r.Method, r.RequestURI)

		// Serve the next handler in the chain.
		next.ServeHTTP(w, r)

		log.Printf("<-- %s %s (%v)", r.Method, r.RequestURI, time.Since(start))
	})
}
