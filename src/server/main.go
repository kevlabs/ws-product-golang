package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kevlabs/eq-golang-server/src/server/lib/counters"
	mware "github.com/kevlabs/eq-golang-server/src/server/lib/middleware"
	"github.com/kevlabs/eq-golang-server/src/server/routes"
)

var (
	liveCounters  = counters.NewContentCounters()
	countersStore = counters.NewCountersStore()
	content       = []string{"sports", "entertainment", "business", "education"}
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func printCounters(w http.ResponseWriter, r *http.Request, next func()) {
	// do nothing
	next()
}

func router() mware.Handler {

	var mux *http.ServeMux = http.NewServeMux()
	mux.HandleFunc("/", welcomeHandler)
	mux.Handle("/view/", routes.ViewHandler(liveCounters, content))
	mux.Handle("/stats/", routes.StatsHandler(liveCounters, countersStore, content))

	return func(w http.ResponseWriter, r *http.Request, next func()) {
		mux.ServeHTTP(w, r)
	}
}

func main() {

	stopUpload := make(chan bool)
	// upload to store every 5 seconds
	go counters.UploadCounters(liveCounters, countersStore, 5000, stopUpload)

	// register app-level middleware
	http.Handle("/", mware.UseHandlers(mware.Logger, printCounters, router()))

	// start server
	log.Fatal(http.ListenAndServe(":8080", mware.ListenHandler(8080)))
}
