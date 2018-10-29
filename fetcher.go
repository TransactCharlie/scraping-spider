package main

import (
	"golang.org/x/net/html"
	"net/url"
	"github.com/transactcharlie/scraping-spider/httpclient"
	"strings"
)

type page struct {
	url   *url.URL
	title string
	links []*url.URL
}

func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
			return
		}
	}
	// If we got here return not OK and empty href
	return
}

func fetchLinks(client *httpclient.Client, link *url.URL, out chan<- *url.URL, finished chan<- *page) {

	page := &page{url: link}
	// Defer a write for the final result so that it always happens
	defer func() { finished <- page }()

	// Fetch the link
	resp, err := client.Get(link)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		switch {

		case tt == html.ErrorToken:
			// End of the document, we're done
			return

		case tt == html.StartTagToken:
			t := z.Token()

			// Title
			isTitle := t.Data == "title"
			if isTitle {
				z.Next()
				title := z.Token()
				page.title = title.String()
				continue
			}
			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, u := getHref(t)
			if !ok {
				continue
			}

			// Parse the url into a url.URL
			candidateUrl, err := url.Parse(u)
			if err != nil {
				continue
			}

			// Handle relative vs absolute links
			absUrl := link.ResolveReference(candidateUrl)

			// Strip any fragments or query strings
			absUrl.RawQuery = ""
			absUrl.Fragment = ""

			// Remove any redundant trailing '/' '///' etc
			absUrl.Path = strings.TrimRight(absUrl.Path, "/")

			page.links = append(page.links, absUrl)

			// Emit the link to be handled
			out <- absUrl
		}
	}
}
