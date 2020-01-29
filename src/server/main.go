package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	counters "github.com/kevlabs/eq-golang-server/src/server/lib/counters"
	mware "github.com/kevlabs/eq-golang-server/src/server/lib/middleware"
)

var (
	c       = counters.NewContentCounters()
	content = []string{"sports", "entertainment", "business", "education"}
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	// randomly pick content type
	contentType := content[rand.Intn(len(content))]

	// increment view counter
	c.AddView(contentType)

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	// simulate random click call - 50% odds
	if rand.Intn(100) < 50 {
		processClick(contentType)
	}
}

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(contentType string) error {
	// increment click counter
	c.AddClick(contentType)
	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("IN STATS")
	if !isAllowed() {
		w.WriteHeader(429)
		return
	}
}

func isAllowed() bool {
	return true
}

func uploadCounters() error {
	return nil
}

func printCounters(w http.ResponseWriter, r *http.Request, next func()) {
	log.Println(c["sports"])
	next()
}

func router() mware.Handler {

	var mux *http.ServeMux = http.NewServeMux()
	mux.HandleFunc("/", welcomeHandler)
	mux.HandleFunc("/view/", viewHandler)
	// mux.HandleFunc("/stats/", statsHandler)
	mux.Handle("/stats/", mware.UseHandlers(mware.WrapHttpHandler(statsHandler)))

	return func(w http.ResponseWriter, r *http.Request, next func()) {
		mux.ServeHTTP(w, r)
	}
}

func listenHandler(port int) http.Handler {
	log.Printf("Server listening on port %v\n", port)
	return http.DefaultServeMux
}

func main() {

	// register middleware
	http.Handle("/", mware.UseHandlers(mware.Logger, printCounters, router()))

	//
	// dummy := newContentCounters("sports", "entertainment", "business", "education")
	// dummy["sports"].view++
	// log.Println(dummy["sports"].view)
	// // log.Println(dummy)
	// dummy.reset()
	// log.Println(dummy["sports"].view)
	// dummy["sports"] = counters{}
	// val, ok := dummy["sports"]
	// log.Println(dummy)
	// log.Println(val, ok)

	// start server
	log.Fatal(http.ListenAndServe(":8080", listenHandler(8080)))
}

// counter interface
// type counters struct {
// 	sync.Mutex
// 	view  int
// 	click int
// }

// var (
// 	c = counters{}

// 	content = []string{"sports", "entertainment", "business", "education"}
// )

// mutex blocks so update operation needs to happen in a go routine so it doesn;t block main thread
