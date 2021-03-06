package middleware

import (
	"net/http"
)

// middleware handler type
type Handler func(http.ResponseWriter, *http.Request, func())

type HandlerSeries struct {
	handlers []Handler
}

// Handler implements the http.Handler interface
func (s *HandlerSeries) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nextHandlerIndex := 0

	// declare next function to keep track or position in handler series and call next handler
	var next func()
	next = func() {
		if nextHandlerIndex < len(s.handlers) {
			handler := s.handlers[nextHandlerIndex]
			nextHandlerIndex++
			handler(w, r, next)
		}
	}

	// call first handler
	next()
}

// register all handlers (e.g. at route or app level)
func UseHandlers(handlers ...Handler) *HandlerSeries {
	return &HandlerSeries{handlers}
}

// wraps http handler in middleware handler. Should be placed last in middleware chain as next will not be called
func WrapHttpHandler(httpHandler func(w http.ResponseWriter, r *http.Request)) Handler {
	return func(w http.ResponseWriter, r *http.Request, next func()) {
		httpHandler(w, r)
	}
}
