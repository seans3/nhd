package middleware

import (
	"net/http"
	"time"
)

// Timeout is a middleware that sets a timeout for each request.
// If the handler takes longer than the duration, it will send a
// 503 Service Unavailable response and the context of the request will be cancelled.
func Timeout(next http.Handler, duration time.Duration) http.Handler {
	return http.TimeoutHandler(next, duration, "request timed out")
}
