package liltunnel

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

type cache struct {
	flushFile string
	cache     map[string]response
	verbose   bool
	Mutex     sync.Mutex
}

func (c *cache) get(key string) (response, bool) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	val, ok := c.cache[key]
	return val, ok
}

func (c *cache) put(key string, val response) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.cache[key] = val
}

func (c *cache) flush() {
	c.Mutex.Lock()
	out, err := json.Marshal(c)
	c.Mutex.Unlock()
	if c.verbose {
		log.Println("could not marshal json:", err)
		return
	}

	err = ioutil.WriteFile(c.flushFile, out, 0777)
	if err != nil {
		log.Println("could not dump cache to disk:", err)
	}
}

func flushAfterPeriod(c *cache, sleep time.Duration) {
	t := time.After(sleep)
	select {
	case <-t:
		c.flush()
	}
}

type response struct {
	Body string `json:"body"`
}
