package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/url"
)


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

func fetchLinks(client *Client, link *url.URL, out chan<- *url.URL) {

	fmt.Println("Fetching: ", link)

	// Fetch the link
	resp, err := client.Get(link)
	if err != nil {
		fmt.Println(err)
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
			out <- absUrl

		}
	}
}