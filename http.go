package liltunnel

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const (
	flushPeriod = 10 * time.Second
)

// NewHTTPTunnler returns an HTTP only Tunneler. You can configure the HTTP
// tunneler to cache HTTP responses out to disk. Which can be _pretty_
// handy in the right scenario.
func NewHTTPTunnler(dialer Dialer, localPort string, remotePort string, cache bool, ttl time.Duration, serveStale bool, log logger) (Tunneler, error) {
	h := &httpTunnel{
		dialer:          dialer,
		localPort:       localPort,
		remotePort:      remotePort,
		log:             log,
		cache:           cache,
		cacheTTL:        ttl,
		cacheServeStale: serveStale,
	}
	return h, nil
}

type httpTunnel struct {
	dialer          Dialer
	log             logger
	localPort       string
	remotePort      string
	cache           bool
	cacheTTL        time.Duration
	cacheServeStale bool
}

func (h *httpTunnel) Tunnel() error {
	url, err := url.Parse(
		fmt.Sprintf("http://localhost%v", h.localPort),
	)
	if err != nil {
		return errors.Wrap(err, "could not parse url")
	}
	rp := httputil.NewSingleHostReverseProxy(url)
	t := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext:           h.dialer.DialContext,
	}

	rp.Transport = t
	// Cast ReverseProxy as an HTTP handler so we can wrap it with the cache
	handler := http.Handler(rp)
	if h.cache {
		h.log.Println("initializing http cache")
		c, err := newHTTPCache(
			"",
			h.cacheTTL,
			h.cacheServeStale,
			h.log,
		)
		if err != nil {
			return err
		}

		handler = c.Handle(handler)
	}

	return http.ListenAndServe(h.localPort, handler)
}
