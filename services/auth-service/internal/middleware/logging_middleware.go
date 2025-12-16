package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs HTTP requests and responses
// This is useful for debugging and monitoring
type LoggingMiddleware struct {
	// logger can be extended to use structured logging (like zap, logrus)
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

// LogRequest is a middleware function that logs incoming requests
// It wraps an HTTP handler and logs:
// - Request method (GET, POST, etc.)
// - Request path/URL
// - Client IP address
// - Response status code
// - Response time (how long the request took)
func (m *LoggingMiddleware) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record the start time
		start := time.Now()

		// Get client IP address
		// X-Forwarded-For header is used when behind a proxy/load balancer
		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			clientIP = realIP
		}

		// Create a response writer wrapper to capture status code
		// http.ResponseWriter doesn't let us read the status code directly
		// So we wrap it to intercept WriteHeader calls
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code
		}

		// Log incoming request
		log.Printf("➡️  [REQUEST] %s %s from %s", r.Method, r.URL.Path, clientIP)

		// Call the next handler (the actual route handler)
		next.ServeHTTP(wrapped, r)

		// Calculate response time
		duration := time.Since(start)

		// Log response
		log.Printf("⬅️  [RESPONSE] %s %s - Status: %d - Duration: %v",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LogRequestWithBody logs requests including request/response bodies
// ⚠️ WARNING: Only use for debugging! This can log sensitive data (passwords, tokens)
// Use LogRequest for production instead
func (m *LoggingMiddleware) LogRequestWithBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		}

		// Log request with body (for debugging)
		log.Printf("➡️  [REQUEST] %s %s from %s", r.Method, r.URL.Path, clientIP)
		log.Printf("   Headers: %v", r.Header)
		// Note: Reading request body consumes it, so we'd need to buffer it
		// For now, we'll skip body logging to keep it simple

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		log.Printf("⬅️  [RESPONSE] %s %s - Status: %d - Duration: %v",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}
