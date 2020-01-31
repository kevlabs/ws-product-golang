package routes

import (
	"net/http"
	"time"

	"github.com/kevlabs/eq-golang-server/src/server/lib/counters"
	mware "github.com/kevlabs/eq-golang-server/src/server/lib/middleware"
)

func StatsHandler(c *counters.ContentCounters, s *counters.CountersStore, content []string) *mware.HandlerSeries {

	statsHandler := func(w http.ResponseWriter, r *http.Request) {
		// show store content
		s.WriteAll(w)
	}

	// use rate-limit middleware
	// bucket burst: 5, refill rate: 0.25/s (for reference Shopify's API has a burst of 40 and a refill rate of 2/s)
	// rate is voluntarily low for dev purposes
	return mware.UseHandlers(mware.RateLimit(75, 5, 5*time.Minute), mware.WrapHttpHandler(statsHandler))
}
