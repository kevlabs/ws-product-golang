package middleware

import (
	"net/http"
)

type Handler func(http.ResponseWriter, *http.Request, func())

type HandlerSeries struct {
	handlers []Handler
}

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

func UseHandlers(handlers ...Handler) *HandlerSeries {
	return &HandlerSeries{handlers}
}
