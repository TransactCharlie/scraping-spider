package main

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	httpClient   *http.Client
	baseHostName string
}

func (c *Client) Get(url *url.URL) (resp *http.Response, err error) {
	return c.httpClient.Get(url.String())
}

func (c *Client) handleRedirects(req *http.Request, via []*http.Request) error {

	if req.URL.Hostname() != c.baseHostName {
		return http.ErrUseLastResponse
	}
	return nil
}

func newClient(initialURL *url.URL) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(time.Second * 10),
	}
	c := &Client{
		httpClient:   client,
		baseHostName: initialURL.Hostname(),
	}
	c.httpClient.CheckRedirect = c.handleRedirects
	return c
}
