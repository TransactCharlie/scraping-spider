package main

import (
	"sync"
	"net/url"
	"fmt"
)


type token = struct{}

var (
	rawLinks       = make(chan *url.URL)
	filteredLinks  = make(chan *url.URL, 10000)
	connectionPool = make(chan struct{}, 50)
	workerGroup    = sync.WaitGroup{}
	stopObserver   = make(chan token)
	client         = newClient("monzo.com")
)

func main() {

	initialURL, _ := url.Parse("https://monzo.com/")

	// Filter (makes sure we don't visit the same link more than once)
	StartFilter(initialURL, rawLinks, filteredLinks)

	// Observer
	StartObserver(stopObserver, rawLinks, filteredLinks)

	// Fill Worker Pool Tokens
	for _, t := range(make([]token, 50)) {
		connectionPool <- t
	}

	// Initial Fetch
	filteredLinks <- initialURL

	for {
		select {
		case l := <-filteredLinks:
			workerGroup.Add(1)
			go func(link *url.URL) {
				<-connectionPool
				fetchLinks(client, link, rawLinks)
				connectionPool <- token{}
				workerGroup.Done()
			}(l)
		default:
			// We might have workers in flight even if we have no current work to do
			workerGroup.Wait()
			// If all workers have finished and there is nothing now in filteredLinks we are done
			if len(filteredLinks) == 0 {
				fmt.Println("oooo")
				goto END
			}
		}
	}
	END:
		fmt.Println("closing Down")
		stopObserver <- token{}


}