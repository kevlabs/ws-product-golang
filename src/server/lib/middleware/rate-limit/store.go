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
	s.Lock()
	s.reset()
	s.Unlock()
	return s
}

// PUBLIC API

func (s *IPStore) Has(key string) bool {
	s.checkExpiry()
	return s.nextStore.Has(key) || s.currentStore.Has(key)
}

// returns zero-value IPRecord
func (s *IPStore) Get(key string) *IPRecord {
	s.checkExpiry()
	if s.nextStore.Has(key) {
		return s.nextStore.Get(key)
	}
	return s.currentStore.Get(key)
}

func (s *IPStore) Set(key string, value *IPRecord) *IPStore {
	s.checkExpiry()
	// determines in which store to save key/value pair
	s.getStore(value).Set(key, value)
	return s
}

func (s *IPStore) Delete(key string) *IPStore {
	s.checkExpiry()
	s.currentStore.Delete(key)
	s.nextStore.Delete(key)
	return s
}

// UNEXPORTED METHODS

func (s *IPStore) createStore(expiry time.Time) *InnerIPStore {
	return NewInnerIPStore(expiry)
}

// replaces current store with next store and creates new next store
func (s *IPStore) rollOver() *IPStore {
	currentStore := s.nextStore
	nextStore := s.createStore(currentStore.expiry.Add(time.Duration(s.lifespanMs) * time.Millisecond))

	s.currentStore = currentStore
	s.nextStore = nextStore

	return s
}

// called at init or if both stores have expired
func (s *IPStore) reset() *IPStore {
	currentExpiry := time.Now().Add(time.Duration(s.lifespanMs) * time.Millisecond)
	nextExpiry := currentExpiry.Add(time.Duration(s.lifespanMs) * time.Millisecond)

	s.currentStore = s.createStore(currentExpiry)
	s.nextStore = s.createStore(nextExpiry)

	return s
}

// instead of setting up a job to run every 'lifespanMs' to roll over store, cleanup action is triggered when accessing store data
func (s *IPStore) checkExpiry() *IPStore {
	// if both stores have expired, reset otherwise roll over.
	s.Lock()
	if s.currentStore.IsExpired() {
		if s.nextStore.IsExpired() {
			s.reset()
		} else {
			s.rollOver()
		}
	}
	s.Unlock()
	return s
}

// returns most current store that can accomodate expiry key in value object, or next store if key does not exist
func (s *IPStore) getStore(r *IPRecord) *InnerIPStore {
	s.Lock()
	defer s.Unlock()

	if r.Expiry.Before(s.currentStore.expiry) {
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
	data   map[string]*IPRecord
}

func NewInnerIPStore(expiry time.Time) *InnerIPStore {
	s := &InnerIPStore{expiry: expiry}
	s.data = make(map[string]*IPRecord)
	return s
}

func (s *InnerIPStore) Has(key string) bool {
	s.Lock()
	_, ok := s.data[key]
	s.Unlock()
	return ok
}

func (s *InnerIPStore) Get(key string) *IPRecord {
	s.Lock()
	data := s.data[key]
	s.Unlock()
	return data
}

func (s *InnerIPStore) Set(key string, value *IPRecord) *InnerIPStore {
	s.Lock()
	s.data[key] = value
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
