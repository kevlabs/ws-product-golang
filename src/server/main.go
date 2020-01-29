package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/kevlabs/eq-golang-server/src/server/lib/middleware"
)

type counters struct {
	sync.Mutex
	view  int
	click int
}

var (
	c = counters{}

	content = []string{"sports", "entertainment", "business", "education"}
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	// randomly pick content type
	data := content[rand.Intn(len(content))]
	log.Println("content", data)

	// increment view counter
	// CURRENTLY NOT MAPPED TO CONTENT TYPE
	c.Lock()
	c.view++
	c.Unlock()

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	// simulate random click call - 50% odds
	if rand.Intn(100) < 50 {
		processClick(data)
	}
}

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(data string) error {
	// DATA IS CURRENTLY NOT USED
	// increment click counter
	c.Lock()
	c.click++
	c.Unlock()

	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
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

func printReq(w http.ResponseWriter, r *http.Request, next func()) {
	log.Println("MIDDLEWARE 2")
	next()
}

func router() middleware.Handler {

	var mux *http.ServeMux = http.NewServeMux()
	mux.HandleFunc("/", welcomeHandler)
	mux.HandleFunc("/view/", viewHandler)
	mux.HandleFunc("/stats/", statsHandler)

	return func(w http.ResponseWriter, r *http.Request, next func()) {
		mux.ServeHTTP(w, r)
	}
}

func listenHandler(port int) http.Handler {
	log.Printf("Server listening on port %v\n", port)

	return middleware.UseHandlers(middleware.Logger, printReq, router())
}

func main() {

	log.Fatal(http.ListenAndServe(":8080", listenHandler(8080)))
}
