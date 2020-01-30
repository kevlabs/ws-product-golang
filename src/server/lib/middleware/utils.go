package middleware

import (
	"log"
	"net/http"
	"time"
)

// logs all incoming requests (method, route, processing time)
func Logger(w http.ResponseWriter, r *http.Request, next func()) {
	start := time.Now()
	next()
	log.Printf("\x1b[34m%s \x1b[32m%s \x1b[0m%v", r.Method, r.URL.Path, time.Since(start))
}

// logs port when server starts and returns default mux
func ListenHandler(port int) http.Handler {
	log.Printf("Server listening on port %v\n", port)
	return http.DefaultServeMux
}
