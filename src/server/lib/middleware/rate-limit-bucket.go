/* IPBucket
 * bucket refills based on limit and interval (in s)
 * burst determines maximum bucket capacity at any time
 * bucket is full at onset
 */

package middleware

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type IPBucket struct {
	sync.Mutex
	Expiry   time.Time
	limit    int
	interval time.Duration
	bucket   chan time.Time
	done     chan bool
}

func NewIPBucket(limit int, burst int, interval time.Duration) *IPBucket {
	b := &IPBucket{limit: limit, interval: interval}
	b.bucket = make(chan time.Time, burst)
	b.done = make(chan bool)
	b.setExpiry()
	b.Fill()

	// start refill process
	go b.Start()

	// REMOVE
	fmt.Println("BUCKET CREATED")

	return b
}

// fill to capacity
func (b *IPBucket) Fill() *IPBucket {
	notFull := true
	for notFull {
		err := b.addToken()
		if err != nil {
			notFull = false
		}
	}
	return b
}

// start auto-refilling process
func (b *IPBucket) Start() {
	ticker := time.NewTicker(b.interval / time.Duration(b.limit))
	for {

		// call stop if bucket has expired
		if b.IsExpired() {
			go b.Stop()
		}

		select {
		case <-b.done:
			ticker.Stop()
			close(b.bucket)

			// REMOVE
			fmt.Println("BUCKET CLOSED")
			return

		case <-ticker.C:
			// add token
			err := b.addToken()
			if err != nil {
				// continue if bucket full
				continue
			}
		}
	}
}

func (b *IPBucket) Stop() *IPBucket {
	b.done <- true
	return b
}

func (b *IPBucket) setExpiry() *IPBucket {
	b.Lock()
	b.Expiry = time.Now().Add(b.interval)
	b.Unlock()
	return b
}

func (b *IPBucket) IsExpired() bool {
	b.Lock()
	defer b.Unlock()
	return !b.Expiry.After(time.Now())
}

func (b *IPBucket) addToken() error {
	select {
	case b.bucket <- time.Now():
		// token added
		return nil

	default:
		// bucket is full
		return errors.New("Bucket is full")
	}
}

func (b *IPBucket) RemoveToken() error {
	b.setExpiry()
	select {
	case <-b.bucket:
		//enough in the bucket
		return nil

	default:
		//bucket is empty
		return errors.New("Bucket is empty")
	}
}
