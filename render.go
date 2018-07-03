package main

import (
	"github.com/awalterschulze/gographviz"
)

func escape(s string) string {
	return "\"" + s + "\""
}

func generateGraph(results []*page) string {

	INDOMAIN := "cluster_InDomain"
	EXDOMAIN := "cluster_ExDomain"
	graph := gographviz.NewGraph()
	_ = graph.SetName("Sitemap")
	_ = graph.SetDir(true)
	_ = graph.AddSubGraph("Sitemap", INDOMAIN, nil)
	_ = graph.AddSubGraph("Sitemap", EXDOMAIN, nil)

	nodeCache := map[string]struct{}{}
	// extraNodeCache := map[string]struct{}{}
	edgeCache := map[string]struct{}{}

	// We will do this in two passes to populate the nodes we correctly found
	for _, p := range results {
		graph.AddNode(INDOMAIN, escape(p.url.String()), map[string]string{"label": escape(p.url.Path)})
		nodeCache[p.url.String()] = struct{}{}
	}
	graph.AddNode(EXDOMAIN, "OutsideContextProblem", nil)

	// 2nd Pass for edges -- any edges that point to nodes not in the cache will
	// needs nodes created for them
	for _, p := range results {
		for _, l := range p.links {
			if _, ok := nodeCache[l.String()]; ok {
				edgeName := p.url.String() + "-" + escape(l.String())
				if _, ok := edgeCache[edgeName]; !ok {
					graph.AddEdge(escape(p.url.String()), escape(l.String()), true, nil)
					edgeCache[edgeName] = struct{}{}
				}
			} else {
				edgeName := p.url.String() + "-exdomain"
				if _, ok := edgeCache[edgeName]; !ok {
					graph.AddEdge(escape(p.url.String()), "OutsideContextProblem", true, nil)
					edgeCache[edgeName] = struct{}{}
				}

			}
		}
	}
	output := graph.String()

	return output
}
