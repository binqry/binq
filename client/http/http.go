// Package http wraps net/http to suit binq use cases
package http

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/binqry/binq"
	"github.com/binqry/binq/internal/erron"
)

// Fetch is a shorthand function to execute HTTP GET request primarily to download items.
func Fetch(addr string) (res *http.Response, err error) {
	hc := newDefaultClient()
	req, err := newGetRequest(addr, map[string]string{})
	if err != nil {
		return nil, err
	}
	return hc.Do(req)
}

// FetchIndex is a shorthand function to send HTTP GET request to Binq Index Server.
func FetchIndex(addr string) (res *http.Response, err error) {
	hc := newClientForIndex()
	headers := make(map[string]string)
	headers["Accept"] = "application/json"
	req, err := newGetRequest(addr, headers)
	if err != nil {
		return nil, err
	}
	return hc.Do(req)
}

// Functions to create http.Client & http.Request

var (
	// defaultTimeout represents default timeout period of HTTP requests.
	defaultTimeout      time.Duration = 300 * time.Second
	timeoutToQueryIndex time.Duration = 5 * time.Second
)

const undefinedHTTPTimeout = -1

func newClient(timeout time.Duration) (hc *http.Client) {
	if timeout == undefinedHTTPTimeout {
		timeout = defaultTimeout
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

func newDefaultClient() (hc *http.Client) {
	return newClient(defaultTimeout)
}

func newClientForIndex() (hc *http.Client) {
	return newClient(timeoutToQueryIndex)
}

func newGetRequest(url string, headers map[string]string) (req *http.Request, err error) {
	req, _err := http.NewRequest(http.MethodGet, url, nil)
	if _err != nil {
		return req, erron.Errorwf(_err, "Failed to create HTTP request")
	}
	req.Header.Set("User-Agent", fmt.Sprintf("binq/%s", binq.Version))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}
