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
	"time"
)

var (
	cmdURL      = flag.String("U", "https://monzo.com", "Initial URL")
	cmdPoolSize = flag.Int("C", 25, "Max number of concurrent fetches")

	// Counters to keep track of in-flight workers and urls
	fetchers      = 0
	urlsToProcess = 0

	// Counters for statistics
	fetchedPages = 0
	linksSeen = 0
	linksDiscarded = 0

	// Debug report ticker
	ticker = time.Tick(time.Millisecond * 100)
)

func debugReport() {
	fmt.Fprintf(os.Stderr, "\r Fetched:%d  Discarded:%d  LinksSeen:%d", fetchedPages, linksDiscarded, linksSeen)
}

func main() {
	log := logging.New(os.Stderr, "", 0)
	flag.Parse()

	var (
		initialURL, _  = url.Parse(*cmdURL)
		httpClient     = httpclient.NewClient(initialURL)
		connectionPool = pool.NewPool(*cmdPoolSize)
		results        = []*page{}

		// Communication Channels
		filteredLinks    = make(chan *url.URL)
		discardedLinks   = make(chan *url.URL)
		candidateURLS    = make(chan *url.URL, 1) // We buffer this so we can inject the start URL
		fetchResults     = make(chan *page)

		// filterCandidates are the links returned by workers.
		// buffered to 4 times the number of max in-flight workers to try and balance channel size vs
		// goroutines - if we fill the buffer we spawn goroutines to write the values eventually.
		filterCandidates = make(chan *url.URL, 4 * *cmdPoolSize)

		// linkFilter is in charge of filtering out potential bad links or ones we've visited before
		linkFilter = filter.NewFilter(initialURL, filterCandidates,
			discardedLinks, filteredLinks)


	)

	// Filter
	linkFilter.Run()

	// Initial Fetch
	candidateURLS <- initialURL


	log.Println("Starting Event Loop...")
	for {
		select {

		// New URL to fetch and parse
		case l := <-filteredLinks:
			fetchers++
			urlsToProcess--
			go func(link *url.URL) {
				connectionPool.Claim()
				defer connectionPool.Release()
				fetchLinks(httpClient, link, candidateURLS, fetchResults)
			}(l)

		// Fetcher has finished and returned a page
		case p := <-fetchResults:
			fetchers--
			fetchedPages++
			results = append(results, p)
			if fetchers == 0 && urlsToProcess == 0 {
				goto EXIT
			}

		// Filter discarded a candidate URL
		case _ = <-discardedLinks:
			// fmt.Fprintf(os.Stderr, "D")
			urlsToProcess--
			linksDiscarded++
			if urlsToProcess == 0 && fetchers == 0 {
				goto EXIT
			}

		// Fetcher emitted a candidate URL
		case r := <-candidateURLS:
			urlsToProcess++
			linksSeen++
			// We need to write to the filterCandidates channel -- this has a buffer set to the size of max
			// inflight concurrent fetches. However a fetch can return many links as candidates
			// to get round this we can schedule a goroutine instead if we can't write to the channel.
			// However this does end up spawning a lot of goroutines
			// Eventually, either way we'd end up running out of memory and if we had some exponential
			// explosion of links we'd eventually have to either just fail and run out of memory or drop
			// messages.
			select {
				case filterCandidates <- r:
				default:
					// we weren't able to write to filterCandidates.
					// lets schedule a function to do it when it can...
					go func(candidate *url.URL) {
						filterCandidates <- candidate
					}(r)
			}

		// Debug Report ticker
		case <- ticker:
			debugReport()
		}
	}
EXIT:
	debugReport()
	log.Println()
	log.Println("Finished Collection")
	linkFilter.Stop()
	fmt.Println(generateGraph(results))
}
