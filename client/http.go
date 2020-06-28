package client

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/progrhyme/dlx"
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

func newHttpGetRequest(url string, headers map[string]string) (req *http.Request, err error) {
	req, _err := http.NewRequest(http.MethodGet, url, nil)
	if _err != nil {
		return req, errorwf(_err, "Failed to create HTTP request")
	}
	req.Header.Set("User-Agent", fmt.Sprintf("dlx/%s", dlx.Version))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}
