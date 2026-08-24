package console

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// RequestLog logs every request with method, path and duration.
func RequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// Recover turns panics in handlers into 500 responses.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic serving %s: %v\n%s", r.URL.Path, recovered, debug.Stack())
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
