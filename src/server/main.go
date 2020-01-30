package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kevlabs/eq-golang-server/src/server/lib/counters"
	mware "github.com/kevlabs/eq-golang-server/src/server/lib/middleware"
	"github.com/kevlabs/eq-golang-server/src/server/routes"
)

var (
	c       = counters.NewContentCounters()
	content = []string{"sports", "entertainment", "business", "education"}
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func uploadCounters(c *counters.ContentCounters, done chan bool) {

	// wait for 15 seconds
	// <-time.After(time.Second * 10)
	// fmt.Println("15 seconds in")

	// every 15s
	ticker := time.NewTicker(time.Second * 10)

	for {
		select {
		// exit process if done
		case <-done:
			return

		// await ticker
		case <-ticker.C:
			fmt.Println("15 seconds have elapsed")

			// create channel
			incomingCounters := make(chan counters.KeyCounters)
			go c.Download(incomingCounters, true)

			for data := range incomingCounters {
				fmt.Println("data received", data)
			}
		}
	}
}

func printCounters(w http.ResponseWriter, r *http.Request, next func()) {
	// do nothing
	next()
}

func router() mware.Handler {

	var mux *http.ServeMux = http.NewServeMux()
	mux.HandleFunc("/", welcomeHandler)
	mux.Handle("/view/", routes.ViewHandler(c, content))
	mux.Handle("/stats/", routes.StatsHandler(c, content))

	return func(w http.ResponseWriter, r *http.Request, next func()) {
		mux.ServeHTTP(w, r)
	}
}

func listenHandler(port int) http.Handler {
	log.Printf("Server listening on port %v\n", port)
	return http.DefaultServeMux
}

func main() {

	stopUpload := make(chan bool)
	go uploadCounters(c, stopUpload)

	// register middleware
	http.Handle("/", mware.UseHandlers(mware.Logger, printCounters, router()))

	// start server
	log.Fatal(http.ListenAndServe(":8080", listenHandler(8080)))
}

// mutex blocks so update operation needs to happen in a go routine so it doesn;t block main thread
