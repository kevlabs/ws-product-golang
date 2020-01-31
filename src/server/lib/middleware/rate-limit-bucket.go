/* IPBucket
 * bucket refills based on limit and interval (in s)
 * burst determines maximum bucket capacity at any time
 * bucket is full at onset
 */

package middleware

import (
	"errors"
	"sync"
	"time"
)

type IPBucket struct {
	sync.Mutex
	Expiry    time.Time
	limit     int
	intervalS int
	bucket    chan time.Time
	done      chan bool
}

func NewIPBucket(limit int, burst int, intervalS int) *IPBucket {
	b := &IPBucket{limit: limit, intervalS: intervalS}
	b.bucket = make(chan time.Time, burst)
	b.done = make(chan bool)
	b.setExpiry()
	b.Fill()

	// start refill process
	go b.Start()

	return b
}

// fill to capacity
func (b *IPBucket) Fill() *IPBucket {
	notFull := true
	for notFull {
		err := b.AddToken()
		if err != nil {
			notFull = false
		}
	}
	return b
}

// start auto-refilling process
func (b *IPBucket) Start() {
	ticker := time.NewTicker(time.Duration(b.intervalS/b.limit) * time.Second)
	for {
		select {
		case <-b.done:
			ticker.Stop()
			close(b.bucket)

		case <-ticker.C:
			// add token
			err := b.AddToken()
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
	b.Expiry = time.Now().Add(time.Duration(b.intervalS) * time.Second)
	b.Unlock()
	return b
}

func (b *IPBucket) AddToken() error {
	b.setExpiry()
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
