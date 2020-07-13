// Package client implements HTTP client functionality of binq.
package client

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/binqry/binq/client/http"
	"github.com/binqry/binq/internal/erron"
	"github.com/binqry/binq/internal/urls"
	"github.com/binqry/binq/schema"
	"github.com/binqry/binq/schema/item"
	"github.com/progrhyme/go-lv"
)

var errIndexDataNotFound = errors.New("Index Data is Not Found on given URL")

type Client struct {
	ServerURL *url.URL
	logger    lv.Standard
}

func NewClient(svr *url.URL, logger lv.Standard) (c *Client) {
	return &Client{ServerURL: svr, logger: logger}
}

func (c *Client) GetItemInfo(name string) (tgt *item.Item, err error) {
	tgt, _err := c.GetItemInfoByPath(name)
	switch _err {
	case nil:
		// OK
	case errIndexDataNotFound:
		// Retry
		return c.getItemInfoByIndex(name)
	default:
		return tgt, _err
	}
	return tgt, nil
}

func (c *Client) GetIndex() (index *schema.Index, err error) {
	index, _err := c.getIndex(c.ServerURL.String())
	switch _err {
	case nil:
		// OK
	case errIndexDataNotFound:
		jsonAddr, _err := urls.Join(c.ServerURL.String(), "index.json")
		if _err != nil {
			// Usually unexpected
			return nil, erron.Errorwf(_err, "Can't get index data from server: %s", c.ServerURL)
		}
		if index, _err = c.getIndex(jsonAddr); _err != nil {
			msg := fmt.Sprintf("Retry failed. Can't get index data from server: %s", c.ServerURL)
			return nil, erron.Errorwf(_err, msg)
		}
		return index, nil
	default:
		return nil, _err
	}

	return index, nil
}

func (c *Client) GetItemInfoByPath(pth string) (tgt *item.Item, err error) {
	addr, _err := urls.Join(c.ServerURL.String(), pth)
	if _err != nil {
		// Unexpected case
		return tgt, erron.Errorwf(_err, "Failed to parse server URL: %v", c.ServerURL)
	}

	c.logger.Infof("GET %s", addr)
	res, _err := http.FetchIndex(addr)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to execute HTTP request")
		return tgt, err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		// OK
	case 404:
		c.logger.Debugf("Index Item Data is Not Found: %s", addr)
		return tgt, errIndexDataNotFound
	default:
		err = fmt.Errorf("HTTP response is not OK. Code: %d, URL: %s", res.StatusCode, addr)
		return tgt, err
	}

	var b strings.Builder
	_, _err = io.Copy(&b, res.Body)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to read HTTP response")
		return tgt, err
	}

	tgt, err = item.DecodeItemJSON([]byte(b.String()))
	if err != nil {
		return tgt, err
	}
	c.logger.Debugf("Decoded JSON: %s", tgt)

	return tgt, nil
}

func (c *Client) getIndex(addr string) (index *schema.Index, err error) {
	c.logger.Infof("GET %s", addr)

	res, _err := http.FetchIndex(addr)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to execute HTTP request")
		return nil, err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		// OK
	case 404:
		c.logger.Debugf("Index Data is Not Found: %s", addr)
		return nil, errIndexDataNotFound
	default:
		err = fmt.Errorf("HTTP response is not OK. Code: %d, URL: %s", res.StatusCode, addr)
		return nil, err
	}

	var b strings.Builder
	_, _err = io.Copy(&b, res.Body)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to read HTTP response")
		return index, err
	}

	index, err = schema.DecodeIndexJSON([]byte(b.String()))
	if err != nil {
		return index, err
	}
	c.logger.Debugf("Decoded JSON: %s", index)

	return index, nil
}

func (c *Client) getItemInfoByIndex(name string) (tgt *item.Item, err error) {
	index, err := c.GetIndex()
	if err != nil {
		return nil, err
	}

	pth := index.FindPath(name)
	switch pth {
	case "":
		err = fmt.Errorf("Can't find item in index: %s", c.ServerURL)
		return tgt, err
	case name:
		err = fmt.Errorf(
			"Found path equals to specified name. Won't retry. name: %s, server: %s", name, c.ServerURL)
		return tgt, err
	default:
		// OK
	}

	tgt, _err := c.GetItemInfoByPath(pth)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to get Item Data on path: %s", pth)
		return tgt, err
	}

	return tgt, nil
}
