/* ExpirableStore struct
 * self-cleaning based on specified lifespan
 * implemented using two maps (constant time lookup), each with a different expiry - See InnerStore struct below
 */

package middleware

import (
	"sync"
	"time"
)

type ExpirableStore struct {
	sync.Mutex
	lifespan     time.Duration
	currentStore *InnerExpirableStore
	nextStore    *InnerExpirableStore
}

type Expirable interface {
	Expiry() time.Time
	IsExpired() bool
}

func NewExpirableStore(lifespan time.Duration) *ExpirableStore {
	s := &ExpirableStore{lifespan: lifespan}
	s.reset()
	return s
}

// PUBLIC API

func (s *ExpirableStore) Has(key string) bool {
	return s.nextStore.Has(key) || s.currentStore.Has(key)
}

// returns zero-value Expirable if not set
// nextStore always takes precedence over currentStore in data retrieval
func (s *ExpirableStore) Get(key string) Expirable {
	if s.nextStore.Has(key) {
		return s.nextStore.Get(key)
	}
	return s.currentStore.Get(key)
}

func (s *ExpirableStore) Set(key string, value Expirable) *ExpirableStore {
	// determine in which store to save key/value pair
	store := s.getStore(value)

	// delete key in nextStore if for whatever reason the expiry has been curtailed
	// and now the value needs to be stored in current store
	// no need to worry about currentStore as nextStore always takes precedence in data retrieval
	if s.nextStore.Has(key) && store != s.nextStore {
		s.nextStore.Delete(key)
	}

	store.Set(key, value)
	return s
}

func (s *ExpirableStore) Delete(key string) *ExpirableStore {
	s.currentStore.Delete(key)
	s.nextStore.Delete(key)
	return s
}

// UNEXPORTED/PRIVATE METHODS

func (s *ExpirableStore) createStore(expiry time.Time) *InnerExpirableStore {
	return NewInnerExpirableStore(expiry)
}

// init or roll stores over
func (s *ExpirableStore) reset() *ExpirableStore {

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

func (s *ExpirableStore) cleanup() {
	ticker := time.NewTicker(s.lifespan)

	for {
		// await ticker
		<-ticker.C
		s.reset()
	}
}

// returns most current store that can accomodate expirable's expiry, or next store if key does not exist
func (s *ExpirableStore) getStore(value Expirable) *InnerExpirableStore {
	s.Lock()
	defer s.Unlock()

	if value.Expiry().Before(s.currentStore.expiry) {
		return s.currentStore
	}

	return s.nextStore
}

// store implemented using a map (constant-time lookup)
type InnerExpirableStore struct {
	sync.Mutex
	expiry time.Time
	data   map[string]Expirable
}

func NewInnerExpirableStore(expiry time.Time) *InnerExpirableStore {
	s := &InnerExpirableStore{expiry: expiry}
	s.data = make(map[string]Expirable)
	return s
}

func (s *InnerExpirableStore) Has(key string) bool {
	s.Lock()
	_, ok := s.data[key]
	s.Unlock()
	return ok
}

func (s *InnerExpirableStore) Get(key string) Expirable {
	s.Lock()
	data := s.data[key]
	s.Unlock()
	return data
}

func (s *InnerExpirableStore) Set(key string, value Expirable) *InnerExpirableStore {
	s.Lock()
	s.data[key] = value
	s.Unlock()
	return s
}

func (s *InnerExpirableStore) Delete(key string) *InnerExpirableStore {
	s.Lock()
	delete(s.data, key)
	s.Unlock()
	return s
}

func (s *InnerExpirableStore) IsExpired() bool {
	return !time.Now().Before(s.expiry)
}
