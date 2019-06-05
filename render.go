package main

import (
	"github.com/awalterschulze/gographviz"
)

func escape(s string) string {
	return "\"" + s + "\""
}

func generateGraph(results []*page) string {

	graph := gographviz.NewGraph()
	_ = graph.SetName("Sitemap")
	_ = graph.SetDir(true)
	_ = graph.SetStrict(true)

	nodeCache := map[string]struct{}{}
	edgeCache := map[string]struct{}{}

	// We will do this in two passes to populate the nodes we correctly found
	for _, p := range results {
		graph.AddNode("", escape(p.url.String()), map[string]string{"label": escape(p.url.Path)})
		nodeCache[p.url.String()] = struct{}{}
	}

	// 2nd Pass for edges -- any edges that point to nodes not in the cache will be discarded.
	for _, p := range results {
		for _, l := range p.links {
			if _, ok := nodeCache[l.String()]; ok {
				edgeName := p.url.String() + "-" + escape(l.String())
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
