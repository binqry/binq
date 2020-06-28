package client

import (
	"net"
	"net/http"
	"time"
)

// DefaultHTTPTimeout represents default timeout period of HTTP requests.
// See init.go for default value
var DefaultHTTPTimeout time.Duration

const undefinedHTTPTimeout = -1

func newHttpClient(timeout time.Duration) (hc *http.Client) {
	if timeout == undefinedHTTPTimeout {
		timeout = DefaultHTTPTimeout
	}
	hc = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: timeout,
	}
	return hc
}
