package liltunnel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/user"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultCacheFile = "liltunnel.httpcache.json"
	defaultTTL       = 12 * time.Hour // Cache for a workday
	headerCache      = "X-Liltunnel-Cache"
	cacheHit         = "hit"
	cacheMiss        = "miss"
)

func newHTTPCache(cacheFile string, l logger) (*cache, error) {
	if cacheFile == "" {
		curr, err := user.Current()
		if err != nil {
			return nil, errors.Wrap(err, "newHTTPCache")
		}

		cacheFile = path.Join(
			curr.HomeDir,
			fmt.Sprintf(".%v", defaultCacheFile),
		)
	}

	var responseCache map[string]response
	f, err := ioutil.ReadFile(cacheFile)
	if err == nil {
		if jsonErr := json.Unmarshal(f, &responseCache); jsonErr != nil {
			return nil, errors.Wrapf(jsonErr, "could not unmarshal %v into JSON", cacheFile)
		}
	} else {
		responseCache = make(map[string]response)
	}

	cc := &cache{
		flushFile:     cacheFile,
		responseCache: responseCache,
		log:           l,
		ttl:           defaultTTL,
		m:             new(sync.Mutex),
	}

	return cc, nil
}

type cache struct {
	flushFile  string
	serveStale bool
	ttl        time.Duration
	log        logger

	m             *sync.Mutex
	responseCache map[string]response
}

type response struct {
	Body         string      `json:"body"`
	ResponseCode int         `json:"response_code"`
	ExpiresAt    time.Time   `json:"expires_at"`
	Headers      http.Header `json:"headers"`
}

func (c *cache) Handle(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			cachedResponse, ok := c.get(req.URL.String())
			if ok {
				copyCachedResponse(cachedResponse, w)
				return
			}
		}

		rw := &httpResponseWriter{orig: w}
		next.ServeHTTP(rw, req)

		if canCache(rw.code) {
			c.put(req.URL.String(), rw)
		}

		go c.flush()
	}
	return http.HandlerFunc(fn)
}

func (c *cache) get(key string) (response, bool) {
	c.m.Lock()
	defer c.m.Unlock()
	val, ok := c.responseCache[key]
	if !ok {
		c.log.Println("cache miss")
		return response{}, false
	}

	if c.serveStale {
		c.log.Println("serving stale cache read")
		return val, true
	}

	// still fresh!
	if time.Now().Before(val.ExpiresAt) {
		c.log.Println("serving fresh cache read")
		return val, true
	}

	// we had a value, but force a refresh
	c.log.Println("forcing cache refresh")
	return val, false
}

func (c *cache) put(key string, rw *httpResponseWriter) {
	resp := response{
		Body:         rw.body.String(),
		ResponseCode: rw.code,
		ExpiresAt:    time.Now().Add(c.ttl),
		Headers:      rw.Header(),
	}

	c.m.Lock()
	defer c.m.Unlock()
	c.responseCache[key] = resp
}

func (c *cache) flush() {
	c.log.Println("flushing cache to disk")
	c.m.Lock()
	out, err := json.Marshal(c.responseCache)
	c.m.Unlock()

	if err != nil {
		c.log.Println("could not marshal json:", err)
		return
	}

	err = ioutil.WriteFile(c.flushFile, out, 0777)
	if err != nil {
		c.log.Println("could not dump cache to disk:", err)
	}
}

func (c *cache) flushAsync(sleep time.Duration) {
	go func(s time.Duration) {
		for {
			t := time.After(s)
			select {
			case <-t:
				c.flush()
			}
		}
	}(sleep)
}

func canCache(code int) bool {
	switch code {
	case http.StatusOK, http.StatusAccepted, http.StatusCreated:
		return true
	default:
		return false
	}
}

func copyCachedResponse(resp response, w http.ResponseWriter) {
	w.Header().Add(headerCache, cacheHit)
	for k, v := range resp.Headers {
		if len(v) == 1 {
			w.Header().Set(k, v[0])
		} else if len(v) > 1 {
			w.Header().Set(k, strings.Join(v, ","))
		}
	}

	w.WriteHeader(resp.ResponseCode)
	w.Write([]byte(resp.Body))
}
