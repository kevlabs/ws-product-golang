/*
Find in this file logic related to the rate limit middleware
Implements the token bucket algorithm (https://en.wikipedia.org/wiki/Token_bucket) with go's channels
*/

package middleware

import (
	"net"
	"net/http"
	"strconv"
	"time"
)

func RateLimit(limit int, burst int, intervalS int) Handler {

	// instantiate store to save IP/counter pairs
	store := NewIPStore(intervalS)

	// middleware
	return func(w http.ResponseWriter, r *http.Request, next func()) {

		currentTime := time.Now()
		ip, err := getIPAddress(r)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		// retrieve counter and/or reset/init if invalid
		var counter *IPBucket
		if store.Has(ip) && store.Get(ip).Expiry.After(currentTime) {
			counter = store.Get(ip)
		} else {
			// counter = &IPBucket{Expiry: time.Now().Add(time.Duration(periodMs) * time.Millisecond)}
			counter = NewIPBucket(limit, burst, intervalS)
		}

		// set headers
		w.Header().Set("RateLimit-Limit", strconv.Itoa(limit))
		// w.Header().Set("RateLimit-Remaining", strconv.Itoa(int(math.Max(float64(limit-counter.Connections-1), 0.0))))
		// w.Header().Set("RateLimit-Reset", strconv.Itoa(int(int64(counter.Expiry.Sub(currentTime))/1000000000)))

		// send 429 if too many requests
		// if counter.Connections >= limit {
		// 	http.Error(w, "Connection limit exceeded. Please try again later.", http.StatusTooManyRequests)
		// 	return
		// }
		errToken := counter.RemoveToken()
		if errToken != nil {
			http.Error(w, "Connection limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// // increment counter
		// counter.Connections++

		// upload counter to store
		store.Set(ip, counter)

		next()

	}
}

func getIPAddress(r *http.Request) (string, error) {
	var address string
	// check forwarded header in case user behind proxy
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		address = forwarded
	} else {
		address = r.RemoteAddr
	}

	IPAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return "", err
	}

	return IPAddress, nil
}
