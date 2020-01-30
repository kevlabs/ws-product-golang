package counters

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type CountersStore struct {
	sync.Mutex
	data []KeyCounters
}

func NewCountersStore() *CountersStore {
	s := &CountersStore{}
	// s.data = make([]KeyCounters)
	return s
}

func (s *CountersStore) Upload(data KeyCounters) {
	s.Lock()
	s.data = append(s.data, data)
	s.Unlock()
}

func (s *CountersStore) PrintAll() {
	var allCounters strings.Builder

	// build data string
	s.Lock()
	for _, data := range s.data {
		fmt.Fprintf(&allCounters, "Key: %v, Views: %v, Clicks: %v\n", data.key, data.counters.view, data.counters.click)
	}
	s.Unlock()

	// print data string when ready
	fmt.Print(allCounters.String())
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
			fmt.Printf("%v seconds have elapsed\n", intervalS)

			// create channel
			incomingCounters := make(chan KeyCounters)
			go c.Download(incomingCounters, true)

			for data := range incomingCounters {
				fmt.Println("data received", data)

				// upload to store
				s.Upload(data)

				// print store content
				s.PrintAll()
			}
		}
	}
}
