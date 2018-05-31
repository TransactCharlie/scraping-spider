package main

import (
	"github.com/awalterschulze/gographviz"
)

func escape(s string) string {
	return "\"" + s + "\""
}

func generateGraph(results []*page) string{

	graphAst, _ := gographviz.ParseString(`digraph Results {}`)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}
	nodeCache := map[string]struct{}{}
	extraNodeCache := map[string]struct{}{}

	// We will do this in two passes to populate the nodes we correctly found
	for _, p := range(results) {
		graph.AddNode("Results", escape(p.url.String()), nil)
		nodeCache[p.url.String()] = struct{}{}
	}

	// 2nd Pass for edges -- any edges that point to nodes not in the cache will
	// needs nodes created for them
	for _, p := range(results) {
		for _, l := range(p.links) {
			if _, ok := nodeCache[l.String()]; ok {
				graph.AddEdge(escape(p.url.String()), escape(l.String()), true, nil)
			} else {
				//TODO: differentiate these nodes
				// TO make it *possible* to render this we'll restrict the link to just the root domain
				hostName := l.Hostname()
				if _, ok := extraNodeCache[hostName]; !ok {
					graph.AddNode("Results", escape(hostName), nil)
				}
				graph.AddEdge(escape(p.url.String()), escape(hostName), true, nil)
			}
		}
	}
	output := graph.String()

	return output
}