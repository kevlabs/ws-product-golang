/*
Find in this file logic related to the rate limit middleware
Implements the token bucket algorithm (https://en.wikipedia.org/wiki/Token_bucket) with go's channels
*/

package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// interval also sets the rate of the ExpirableStore cleanup. As a result, avoid using an extremely small value.
func RateLimit(limit int, burst int, interval time.Duration) Handler {

	// instantiate store to save IP/bucket pairs
	store := NewExpirableStore(interval)

	// approximate limit policy
	var policyQuota, policyWindow int
	// if tokensPerSecond := 1000 * limit / intervalMs; tokensPerSecond == 0 {
	if intervalS := float64(interval) / float64(time.Second); intervalS < 1.0 {
		// secondsPerToken := intervalMs / (1000 * limit)
		policyQuota = int(float64(limit) / intervalS)
		policyWindow = 1
	} else {
		policyQuota = limit
		policyWindow = int(intervalS)
	}
	policyHeader := fmt.Sprintf("%v, %v;window=%v; burst=%v;policy=\"leaky bucket\"", limit, policyQuota, policyWindow, burst)

	// middleware
	return func(w http.ResponseWriter, r *http.Request, next func()) {

		ip := getIPAddress(r)

		// retrieve bucket and/or reset/init if invalid
		var (
			bucket *IPBucket
			ok     bool
		)
		if store.Has(ip) && !store.Get(ip).IsExpired() {
			bucket, ok = store.Get(ip).(*IPBucket)
			if !ok {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		} else {
			bucket = NewIPBucket(limit, burst, interval)
		}

		// set headers
		w.Header().Set("RateLimit-Limit", policyHeader)

		// send 429 if too many requests
		errToken := bucket.RemoveToken()
		if errToken != nil {
			http.Error(w, "Connection limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// upload bucket to store
		store.Set(ip, bucket)

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
