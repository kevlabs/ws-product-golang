/*
Find in this file logic related to the rate limit middleware
Implements the token bucket algorithm (https://en.wikipedia.org/wiki/Token_bucket) with go's channels
*/

package middleware

import (
	"fmt"
	"net/http"
)

// intervalMs also sets the rate of the IPStore cleanup. As a result, avoid using an extremely small value.
func RateLimit(limit int, burst int, intervalMs int) Handler {

	// instantiate store to save IP/counter pairs
	store := NewIPStore(intervalMs)

	// approximate limit policy
	var policyQuota, policyWindow int
	if tokensPerSecond := 1000 * limit / intervalMs; tokensPerSecond == 0 {
		secondsPerToken := intervalMs / (1000 * limit)
		policyQuota = 1
		policyWindow = secondsPerToken
	} else {
		policyQuota = tokensPerSecond
		policyWindow = 1
	}
	policyHeader := fmt.Sprintf("%v, %v;window=%v; burst=%v;policy=\"leaky bucket\"", limit, policyQuota, policyWindow, burst)

	// middleware
	return func(w http.ResponseWriter, r *http.Request, next func()) {

		ip := getIPAddress(r)

		// retrieve counter and/or reset/init if invalid
		var counter *IPBucket
		if store.Has(ip) && !store.Get(ip).IsExpired() {
			counter = store.Get(ip)
		} else {
			counter = NewIPBucket(limit, burst, intervalMs)
		}

		// set headers
		w.Header().Set("RateLimit-Limit", policyHeader)

		// send 429 if too many requests
		errToken := counter.RemoveToken()
		if errToken != nil {
			http.Error(w, "Connection limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// upload counter to store
		store.Set(ip, counter)

		next()
	}
}

func getIPAddress(r *http.Request) string {
	// check forwarded header in case user behind proxy
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}

	return r.RemoteAddr
}
