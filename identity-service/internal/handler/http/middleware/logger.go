package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriterWrapper intercepts and stores the HTTP status code
// so the logger can read it after the request finishes.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader overrides the standard WriteHeader method to capture the status code
func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger intercepts every HTTP request, measures execution time, and logs the metrics
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Wrap the original response writer. Default to 200 OK
		// because Go implicitly sends 200 if WriteHeader is never called.
		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Let the request continue down the pipeline to your use cases/handlers
		next.ServeHTTP(wrapper, r)

		// Calculate how long the execution took
		duration := time.Since(startTime)

		// Log the structured transaction details
		log.Printf(
			"HTTP REQ | %s | %s %s | Status: %d | Duration: %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			wrapper.statusCode,
			duration,
		)
	})
}
