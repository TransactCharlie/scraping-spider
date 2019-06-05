package httpclient

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

// Checks for redirects and won't follow them off domain.
func (c *Client) handleRedirects(req *http.Request, via []*http.Request) error {

	if req.URL.Hostname() != c.baseHostName {
		return http.ErrUseLastResponse
	}
	return nil
}

// Initialises a new client with a base URL
func NewClient(initialURL *url.URL) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 500 * time.Millisecond,
		}).DialContext,
		MaxIdleConns:          0,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		// Add a timeout to abort a slow connection (if the site is still streaming us stuff for example)
		Timeout: time.Duration(time.Second * 60),
	}
	c := &Client{
		httpClient:   client,
		baseHostName: initialURL.Hostname(),
	}
	c.httpClient.CheckRedirect = c.handleRedirects
	return c
}
