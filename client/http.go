package client

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/binqry/binq"
	"github.com/binqry/binq/internal/erron"
)

var (
	// DefaultHTTPTimeout represents default timeout period of HTTP requests.
	// See init.go for default value
	DefaultHTTPTimeout      time.Duration
	httpTimeoutToQueryIndex time.Duration
)

const undefinedHTTPTimeout = -1

func NewHttpClient(timeout time.Duration) (hc *http.Client) {
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

func NewHttpGetRequest(url string, headers map[string]string) (req *http.Request, err error) {
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
