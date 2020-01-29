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

type ContentCounters struct {
	sync.Mutex
	data map[string]*Counters
}

func NewContentCounters() *ContentCounters {
	c := &ContentCounters{}
	c.Clear()
	return c
}

func (c *ContentCounters) Clear() {
	c.data = make(map[string]*Counters)
}

func (c *ContentCounters) getCounters(content string) *Counters {
	key := fmt.Sprintf("%v:%v", content, time.Now().Format("2006-01-02 15:04"))

	// init counters if not found
	counters, ok := c.data[key]
	if !ok {
		counters = &Counters{}
		c.data[key] = counters
	}

	return counters
}

func (c *ContentCounters) AddView(content string) {
	counters := c.getCounters(content)

	counters.Lock()
	counters.view++
	counters.Unlock()

}

func (c *ContentCounters) AddClick(content string) {
	counters := c.getCounters(content)

	counters.Lock()
	counters.click++
	counters.Unlock()

}
