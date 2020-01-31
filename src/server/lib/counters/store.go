/*
Find in this file logic related to the mock store including upload of data to it
*/

package counters

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// implemented with slice data structure
type CountersStore struct {
	sync.Mutex
	data []KeyCounters
}

func NewCountersStore() *CountersStore {
	s := &CountersStore{}
	// s.data = make([]KeyCounters)
	return s
}

func (s *CountersStore) Upload(data KeyCounters) *CountersStore {
	s.Lock()
	s.data = append(s.data, data)
	s.Unlock()
	return s
}

func (s *CountersStore) WriteAll(w io.Writer) *CountersStore {
	s.Lock()
	// build data string
	for _, data := range s.data {
		fmt.Fprintf(w, "Key: %v, Views: %v, Clicks: %v\n", data.key, data.counters.view, data.counters.click)
	}
	s.Unlock()
	return s
}

func (s *CountersStore) PrintAll() *CountersStore {
	var allCounters strings.Builder
	s.WriteAll(&allCounters)
	fmt.Print(allCounters.String())
	return s
}

func UploadCounters(c *ContentCounters, s *CountersStore, intervalS int, done chan bool) {

	ticker := time.NewTicker(time.Duration(intervalS) * time.Second)

	for {
		select {
		// exit process if done
		case <-done:
			return

		// await ticker
		case <-ticker.C:
			// fmt.Printf("%v seconds have elapsed\n", intervalS)

			// create channel
			incomingCounters := make(chan KeyCounters)
			go c.Download(incomingCounters, true)

			for data := range incomingCounters {
				// upload chunk to store
				s.Upload(data)
			}

			// print store content - for demo purposes only
			// fmt.Println("PRINTING STORE CONTENT")
			// s.PrintAll()
		}
	}
}
