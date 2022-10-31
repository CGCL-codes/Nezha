package graph

import (
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type intset map[int]struct{}

func linksTo(i []int) intset {
	if len(i) == 0 {
		return nil
	}
	s := make(intset)
	for _, v := range i {
		s[v] = struct{}{}
	}
	return s
}

func CreateJohnsonGraph(gSlice [][]int) []intset {
	var g = make([]intset, len(gSlice))

	for i, v := range gSlice {
		g[i] = linksTo(v)
	}

	return g
}

func FindAllCycles(g []intset) [][]int {
	newGraph := simple.NewDirectedGraph()
	newGraph.AddNode(simple.Node(-10))

	for u, e := range g {
		// Add nodes that are not defined by an edge.
		if newGraph.Node(int64(u)) == nil {
			newGraph.AddNode(simple.Node(u))
		}
		for v := range e {
			newGraph.SetEdge(simple.Edge{F: simple.Node(u), T: simple.Node(v)})
		}
	}

	cycles := topo.DirectedCyclesIn(newGraph)

	var got = make([][]int, len(cycles))

	if cycles != nil {
		for j, c := range cycles {
			ids := make([]int, len(c))
			for k, n := range c {
				ids[k] = int(n.ID())
			}
			got[j] = ids
		}
	}

	return got
}

