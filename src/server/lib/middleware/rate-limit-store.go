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
	lifespan     time.Duration
	currentStore *InnerIPStore
	nextStore    *InnerIPStore
}

func NewIPStore(lifespan time.Duration) *IPStore {
	s := &IPStore{lifespan: lifespan}
	s.reset()
	return s
}

// PUBLIC API

func (s *IPStore) Has(key string) bool {
	return s.nextStore.Has(key) || s.currentStore.Has(key)
}

// returns zero-value IPBucket if not set
// nextStore always takes precedence over currentStore in data retrieval
func (s *IPStore) Get(key string) *IPBucket {
	if s.nextStore.Has(key) {
		return s.nextStore.Get(key)
	}
	return s.currentStore.Get(key)
}

func (s *IPStore) Set(key string, bucket *IPBucket) *IPStore {
	// determine in which store to save key/bucket pair
	store := s.getStore(bucket)

	// delete key in nextStore if for whatever reason the expiry has been curtailed
	// and now the value needs to be stored in current store
	// no need to worry about currentStore as nextStore always takes precedence in data retrieval
	if s.nextStore.Has(key) && store != s.nextStore {
		s.nextStore.Delete(key)
	}

	store.Set(key, bucket)
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
		currentExpiry := time.Now().Add(s.lifespan)
		nextExpiry := currentExpiry.Add(s.lifespan)

		s.currentStore = s.createStore(currentExpiry)
		s.nextStore = s.createStore(nextExpiry)

		s.Unlock()

		// set up cleaning process
		go s.cleanup()

		return s
	}

	// rollover
	nextExpiry := time.Now().Add(2 * s.lifespan)

	s.currentStore = s.nextStore
	s.nextStore = s.createStore(nextExpiry)

	s.Unlock()

	return s
}

func (s *IPStore) cleanup() {
	ticker := time.NewTicker(s.lifespan)

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
	return !time.Now().Before(s.expiry)
}
