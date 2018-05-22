package main

import (
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseHostName string
}

func (c *Client ) Get(url *url.URL) (resp *http.Response, err error) {
	return c.httpClient.Get(url.String())
}

func (c *Client ) handleRedirects(req *http.Request, via []*http.Request) error {

	if req.URL.Hostname() != c.baseHostName {
		return http.ErrUseLastResponse
	}
	return nil
}

func newClient(baseHostName string) *Client {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    10 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{
		Transport: tr,
	}
	c := &Client{
		httpClient: client,
		baseHostName: baseHostName,
		}
	c.httpClient.CheckRedirect = c.handleRedirects
	return c
}