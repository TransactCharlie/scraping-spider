package main

import (
	"fmt"
	"github.com/awalterschulze/gographviz"
	"strings"
)

func escape(s string) string {
	return "\"" + s + "\""
}

func basePath(u *page) string {
	paths := strings.Split(u.url.Path, "/")
	if len(paths) > 1 {
		return paths[0]
	}
	return ""
}

func generateGraph(results []*page) string {

	graph := gographviz.NewGraph()
	_ = graph.SetName("Sitemap")
	_ = graph.SetDir(true)
	_ = graph.SetStrict(true)

	subGraphCache := map[string]struct{}{}
	nodeCache := map[string]struct{}{}
	edgeCache := map[string]struct{}{}

	// We will do this in two passes to populate the nodes we correctly found
	for _, p := range results {
		subgraph := basePath(p)
		fmt.Print(subgraph)
		_, found := subGraphCache[subgraph]
		if !found {
			err := graph.AddSubGraph("Sitemap", subgraph, nil)
			if err != nil {
				panic(err)
			}
			subGraphCache[subgraph] = struct{}{}
		}
		graph.AddNode(subgraph, escape(p.url.String()), map[string]string{"label": escape(p.url.Path)})
		nodeCache[p.url.String()] = struct{}{}
	}

	// 2nd Pass for edges -- any edges that point to nodes not in the cache will be discarded.
	for _, p := range results {
		for _, l := range p.links {
			if _, ok := nodeCache[l.String()]; ok {
				edgeName := p.url.String() + "-" + escape(l.String())

				// bypass any backwards links
				if len(strings.Split(p.url.String(), "/")) > len(strings.Split(l.String(), "/")) {
					continue
				}

				// Do we have this edge already?
				if _, ok := edgeCache[edgeName]; !ok {
					graph.AddEdge(escape(p.url.String()), escape(l.String()), true, nil)
					edgeCache[edgeName] = struct{}{}
				}
			}
		}
	}
	return graph.String()
}
