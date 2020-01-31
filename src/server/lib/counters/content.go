/*
Find in this file logic related to the in-memory storage of counters
*/

package counters

import (
	"fmt"
	"sync"
	"time"
)

type Counters struct {
	sync.Mutex
	view  int
	click int
}

type CountersValue struct {
	view  int
	click int
}

type KeyCounters struct {
	key      string
	counters CountersValue
}

type ContentCounters struct {
	sync.Mutex
	data map[string]*Counters
}

func NewContentCounters() *ContentCounters {
	c := &ContentCounters{}
	c.Clear()
	return c
}

// clear does not lock map to prevent deadlocks (caller should take care of locking)
func (c *ContentCounters) clear() *ContentCounters {
	c.data = make(map[string]*Counters)
	return c
}

// reset data field to an empty map
func (c *ContentCounters) Clear() *ContentCounters {
	c.Lock()
	c.clear()
	c.Unlock()
	return c
}

// iterate over data and send it over supplied channel
func (c *ContentCounters) Download(channel chan KeyCounters, clear bool) *ContentCounters {
	c.Lock()
	for key, counters := range c.data {
		channel <- KeyCounters{key, CountersValue{counters.view, counters.click}}
	}
	close(channel)

	if clear {
		c.clear()
	}

	c.Unlock()
	return c
}

// retrieve counters based on content type and time
// create counters if does not exist
func (c *ContentCounters) getCounters(content string) *Counters {
	key := fmt.Sprintf("%v:%v", content, time.Now().Format("2006-01-02 15:04"))

	// init counters if not found
	c.Lock()
	counters, ok := c.data[key]
	if !ok {
		counters = &Counters{}
		c.data[key] = counters
	}
	c.Unlock()

	return counters
}

func (c *ContentCounters) AddView(content string) *ContentCounters {
	counters := c.getCounters(content)

	counters.Lock()
	counters.view++
	counters.Unlock()

	return c
}

func (c *ContentCounters) AddClick(content string) *ContentCounters {
	counters := c.getCounters(content)

	counters.Lock()
	counters.click++
	counters.Unlock()

	return c
}
