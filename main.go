package main

import (
	"flag"
	"fmt"
	"github.com/transactcharlie/scraping-spider/filter"
	"github.com/transactcharlie/scraping-spider/pool"
	"github.com/transactcharlie/scraping-spider/httpclient"
	logging "log"
	"net/url"
	"os"
)

var (
	cmdURL      = flag.String("U", "https://monzo.com", "Initial URL")
	cmdPoolSize = flag.Int("C", 25, "Max number of concurrent fetches")
)

func main() {
	log := logging.New(os.Stderr, "", 0)
	flag.Parse()

	var (
		initialURL, _  = url.Parse(*cmdURL)
		httpClient     = httpclient.NewClient(initialURL)
		connectionPool = pool.NewPool(*cmdPoolSize)
		results        = []*page{}

		// Communication Channels
		filterCandidates = make(chan *url.URL)
		filteredLinks    = make(chan *url.URL)
		discardedLinks   = make(chan *url.URL)
		candidateURLS    = make(chan *url.URL, 1) // We buffer this so we can inject the start URL
		fetchResults     = make(chan *page)

		// linkFilter is in charge of filtering out potential bad links or ones we've visited before
		linkFilter = filter.NewFilter(initialURL, filterCandidates,
			discardedLinks, filteredLinks)

		// Counters to keep track of in-flight workers and urls
		fetchers      = 0
		urlsToProcess = 0
	)

	// Filter
	linkFilter.Run()

	// Initial Fetch
	candidateURLS <- initialURL

	log.Println("Starting Event Loop")
	for {
		select {

		// New URL to fetch and parse
		case l := <-filteredLinks:
			fmt.Fprintf(os.Stderr, ".")
			fetchers++
			urlsToProcess--
			go func(link *url.URL) {
				connectionPool.Claim()
				fetchLinks(httpClient, link, candidateURLS, fetchResults)
				connectionPool.Release()
			}(l)

		// Fetcher has finished and returned a page
		case p := <-fetchResults:
			fmt.Fprintf(os.Stderr, "P")
			fetchers--
			results = append(results, p)
			if fetchers == 0 && urlsToProcess == 0 {
				goto EXIT
			}

		// Filter discarded a candidate URL
		case _ = <-discardedLinks:
			fmt.Fprintf(os.Stderr, "D")
			urlsToProcess--
			if urlsToProcess == 0 && fetchers == 0 {
				goto EXIT
			}

		// Fetcher emitted a candidate URL
		case r := <-candidateURLS:
			fmt.Fprintf(os.Stderr, "+")
			urlsToProcess++
			// We need to run this in a goroutine so this loop is *always*
			// available to consume new events.
			go func() {
				filterCandidates <- r
			}()
		}
	}
EXIT:
	log.Println()
	log.Println("Finished")
	linkFilter.Stop()
	fmt.Println(generateGraph(results))
}
