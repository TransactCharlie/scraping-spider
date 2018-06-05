package main

import (
	"flag"
	"fmt"
	"github.com/transactcharlie/scraping-spider/filter"
	"github.com/transactcharlie/scraping-spider/pool"
	"net/url"
)

var (
	cmdURL      = flag.String("U", "https://monzo.com", "Initial URL")
	cmdPoolSize = flag.Int("C", 25, "Max number of concurrent fetches")
)

func main() {
	flag.Parse()

	var (
		filterCandidates = make(chan *url.URL, 1)
		filteredLinks    = make(chan *url.URL, 10)
		discardedLinks   = make(chan *url.URL, 10)
		initialURL, _    = url.Parse(*cmdURL)
		linkFilter       = filter.NewFilter(initialURL, filterCandidates, discardedLinks, filteredLinks)
		httpClient       = newClient(initialURL)
		connectionPool   = pool.NewPool(*cmdPoolSize)
		results          = []*page{}
		candidateURLS    = make(chan *url.URL, 1)
		fetchResults     = make(chan *page, 1)
		fetchers         = 0
		urlsToProcess    = 0
	)

	// Filter
	linkFilter.Run()

	// Initial Fetch
	candidateURLS <- initialURL

	for {
		select {

		// New URL to fetch and parse
		case l := <-filteredLinks:
			fetchers++
			urlsToProcess--
			go func(link *url.URL) {
				connectionPool.Claim()
				fetchLinks(httpClient, link, candidateURLS, fetchResults)
				connectionPool.Release()
			}(l)

		// Fetcher has finished and returned a page
		case p := <-fetchResults:
			fetchers--
			results = append(results, p)
			if fetchers == 0 && urlsToProcess == 0 {
				goto EXIT
			}

		// Filter discarded a candidate URL
		case <-discardedLinks:
			urlsToProcess--
			if urlsToProcess == 0 && fetchers == 0 {
				goto EXIT
			}

		// Fetcher emitted a candidate URL
		case r := <-candidateURLS:
			urlsToProcess++
			filterCandidates <- r

		}
	}
EXIT:
	linkFilter.Stop()
	fmt.Println(generateGraph(results))
}
