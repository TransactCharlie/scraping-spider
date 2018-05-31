package filter

import (
	"net/url"
	"sync"
)

type filter struct {
	Once        sync.Once
	initialHost string
	seenLinks   map[string]struct{}
	candidates  chan *url.URL
	discards    chan *url.URL
	filtered    chan *url.URL
	running     sync.WaitGroup
}

func NewFilter(initialURL *url.URL, candidates, discards, filtered chan *url.URL) *filter {
	return &filter{
		Once:        sync.Once{},
		initialHost: initialURL.Hostname(),
		seenLinks:   make(map[string]struct{}),
		candidates:  candidates,
		discards:    discards,
		filtered:    filtered,
		running:     sync.WaitGroup{},
	}
}

// Starts the filter running in a goroutine.
// You can only call this one time successfully
func (f *filter) Run() {
	go func() {
		f.Once.Do(func() {
			f.running.Add(1)
			defer f.running.Done()
			f.filterLoop()
			close(f.discards)
			close(f.filtered)
		})
	}()
}

// Stops the filter by closing the candidates channel
func (f *filter) Stop() {
	close(f.candidates)
	f.running.Wait()
}

// Loops over f.candidates while the candidates channel is open
func (f *filter) filterLoop() {
	for l := range f.candidates {

		// Is this link part of the same domain?
		if l.Hostname() != f.initialHost {
			f.discards <- l
			continue
		}

		// Have we seen it before
		if _, ok := f.seenLinks[l.Path]; ok {
			f.discards <- l
			continue
		}

		// This url passed our filtering process
		f.seenLinks[l.Path] = struct{}{}
		f.filtered <- l
	}
}
