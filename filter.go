package main

import (
	"fmt"
	"net/url"
)

func filterLinks(initialURL *url.URL, in <-chan *url.URL, out chan<- *url.URL) {

	// Map to filter out seen paths.
	linksSeen := make(map[string]token)

	initialHost := initialURL.Hostname()

	for l := range(in) {

		fmt.Println("filtering: ", l)
		// Skip the link if it's outside our initial host url
		if l.Hostname() != initialHost {
			continue
		}

		// Skip the link if we saw it before
		if _, ok := linksSeen[l.Path] ; ok {
			continue
		}
		linksSeen[l.Path] = token{}

		// Pass it on to be crawled (or ditch it if we filled up the filteredLinks buffer)
		select {
		case out <- l:
		default:
			fmt.Println("Dropping %s",  l)
			continue
		}

	}
	close(out)
	fmt.Println("Filter Done")
}

// Starts a goroutine running the filter and adds to waitgroup wg
func StartFilter(baseURL *url.URL, in <-chan *url.URL, out chan<- *url.URL ) {
	go func() {
		filterLinks(baseURL, in, out)
	}()
}