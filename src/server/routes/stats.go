package routes

import (
	"log"
	"net/http"

	"github.com/kevlabs/eq-golang-server/src/server/lib/counters"
	mware "github.com/kevlabs/eq-golang-server/src/server/lib/middleware"
)

// TO DO: replace with middleware
func isAllowed() bool {
	return true
}

func StatsHandler(c *counters.ContentCounters, content []string) *mware.HandlerSeries {

	statsHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Println("IN STATS")
		if !isAllowed() {
			w.WriteHeader(429)
			return
		}
	}

	return mware.UseHandlers(mware.WrapHttpHandler(statsHandler))
}
