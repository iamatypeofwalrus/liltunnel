package liltunnel

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// NewHTTPTunnler returns an HTTP only Tunneler. You can configure the HTTP
// tunneler to cache HTTP responses out to disk. Which can be _pretty_
// handy in the right scenario.
func NewHTTPTunnler(dialer Dialer, localPort string, remotePort string, log Logger) (Tunneler, error) {
	h := &httpTunnel{
		dialer:     dialer,
		localPort:  localPort,
		remotePort: remotePort,
		log:        log,
	}
	return h, nil
}

type httpTunnel struct {
	dialer     Dialer
	log        Logger
	localPort  string
	remotePort string
	cache      bool
}

func (h *httpTunnel) Tunnel() error {
	rp, err := h.constructProxy()
	if err != nil {
		return err
	}

	return http.ListenAndServe(":"+h.localPort, rp)
}

func (h *httpTunnel) constructProxy() (*httputil.ReverseProxy, error) {
	url, err := url.Parse(
		fmt.Sprintf("http://localhost:%v", h.remotePort),
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse url")
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

	return rp, nil
}
