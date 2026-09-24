package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"portfolio-backend/internal/config"
)

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger logs incoming HTTP requests with method, path, response status, and duration
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := newResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		log.Printf("[HTTP] %s %s | %d | %s | %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start),
			r.RemoteAddr,
		)
	})
}

// Recovery catches any panics during request handling and returns a 500 Internal Server Error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] %v\nStack trace:\n%s", err, debug.Stack())
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"success":false,"error":"Internal Server Error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS handles Cross-Origin Resource Sharing headers and preflight OPTIONS requests
func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check allowed origins
			allowOrigin := ""
			if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
				allowOrigin = "*"
			} else if origin != "" {
				for _, allowed := range cfg.AllowedOrigins {
					if allowed == "*" || strings.EqualFold(allowed, origin) {
						allowOrigin = origin
						break
					}
				}
			}

			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
