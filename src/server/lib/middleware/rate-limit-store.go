/* IPStore struct
 * self-cleaning based on specified lifespan
 * implemented using two maps (constant time lookup), each with a different expiry - See InnerStore struct below
 */

package middleware

import (
	"sync"
	"time"
)

type IPStore struct {
	sync.Mutex
	lifespanMs   int
	currentStore *InnerIPStore
	nextStore    *InnerIPStore
}

func NewIPStore(lifespanMs int) *IPStore {
	s := &IPStore{lifespanMs: lifespanMs}
	s.reset()
	return s
}

// PUBLIC API

func (s *IPStore) Has(key string) bool {
	return s.nextStore.Has(key) || s.currentStore.Has(key)
}

// returns zero-value IPBucket if not set
func (s *IPStore) Get(key string) *IPBucket {
	if s.nextStore.Has(key) {
		return s.nextStore.Get(key)
	}
	return s.currentStore.Get(key)
}

func (s *IPStore) Set(key string, bucket *IPBucket) *IPStore {
	// determines in which store to save key/bucket pair
	s.getStore(bucket).Set(key, bucket)
	return s
}

func (s *IPStore) Delete(key string) *IPStore {
	s.currentStore.Delete(key)
	s.nextStore.Delete(key)
	return s
}

// UNEXPORTED/PRIVATE METHODS

func (s *IPStore) createStore(expiry time.Time) *InnerIPStore {
	return NewInnerIPStore(expiry)
}

// init or roll stores over
func (s *IPStore) reset() *IPStore {

	s.Lock()

	// init
	if s.currentStore == nil {
		currentExpiry := time.Now().Add(time.Duration(s.lifespanMs) * time.Millisecond)
		nextExpiry := currentExpiry.Add(time.Duration(s.lifespanMs) * time.Millisecond)

		s.currentStore = s.createStore(currentExpiry)
		s.nextStore = s.createStore(nextExpiry)

		s.Unlock()

		// set up cleaning process
		go s.cleanup()

		return s
	}

	// rollover
	nextExpiry := time.Now().Add(time.Duration(2*s.lifespanMs) * time.Millisecond)

	s.currentStore = s.nextStore
	s.nextStore = s.createStore(nextExpiry)

	s.Unlock()

	return s
}

func (s *IPStore) cleanup() {
	ticker := time.NewTicker(time.Duration(s.lifespanMs) * time.Millisecond)

	for {
		// await ticker
		<-ticker.C
		s.reset()
	}
}

// returns most current store that can accomodate bucket's expiry, or next store if key does not exist
func (s *IPStore) getStore(bucket *IPBucket) *InnerIPStore {
	s.Lock()
	defer s.Unlock()

	if bucket.Expiry.Before(s.currentStore.expiry) {
		return s.currentStore
	}

	return s.nextStore
}

type IPRecord struct {
	sync.Mutex
	Expiry      time.Time
	Connections int
}

// store implemented using a map (constant-time lookup)
type InnerIPStore struct {
	sync.Mutex
	expiry time.Time
	data   map[string]*IPBucket
}

func NewInnerIPStore(expiry time.Time) *InnerIPStore {
	s := &InnerIPStore{expiry: expiry}
	s.data = make(map[string]*IPBucket)
	return s
}

func (s *InnerIPStore) Has(key string) bool {
	s.Lock()
	_, ok := s.data[key]
	s.Unlock()
	return ok
}

func (s *InnerIPStore) Get(key string) *IPBucket {
	s.Lock()
	data := s.data[key]
	s.Unlock()
	return data
}

func (s *InnerIPStore) Set(key string, bucket *IPBucket) *InnerIPStore {
	s.Lock()
	s.data[key] = bucket
	s.Unlock()
	return s
}

func (s *InnerIPStore) Delete(key string) *InnerIPStore {
	s.Lock()
	delete(s.data, key)
	s.Unlock()
	return s
}

func (s *InnerIPStore) IsExpired() bool {
	return time.Now().After(s.expiry) || time.Now().Equal(s.expiry)
}
