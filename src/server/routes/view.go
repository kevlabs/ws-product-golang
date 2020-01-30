package routes

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	counters "github.com/kevlabs/eq-golang-server/src/server/lib/counters"
	mware "github.com/kevlabs/eq-golang-server/src/server/lib/middleware"
)

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(c *counters.ContentCounters, contentType string) error {
	// increment click counter
	go c.AddClick(contentType)
	return nil
}

func ViewHandler(c *counters.ContentCounters, content []string) *mware.HandlerSeries {

	viewHandler := func(w http.ResponseWriter, r *http.Request) {
		// randomly pick content type
		contentType := content[rand.Intn(len(content))]

		// increment view counter
		go c.AddView(contentType)

		// simulate delay
		err := processRequest(r)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(400)
			return
		}

		// simulate random click call - 50% odds
		if rand.Intn(100) < 50 {
			err := processClick(c, contentType)
			if err != nil {
				return
			}
		}
	}

	return mware.UseHandlers(mware.WrapHttpHandler(viewHandler))
}
