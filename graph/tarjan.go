package graph

// Connections creates a slice where each item is a slice of strongly connected vertices.
//
// If a slice item contains only one vertex there are no loops. A loop on the
// vertex itself is also a connected group.
//
// The example shows the same graph as in the Wikipedia article.
func Connections(graph map[int][]int) [][]int {
	g := &data{
		graph: graph,
		nodes: make([]node, 0, len(graph)),
		index: make(map[int]int, len(graph)),
	}
	for v := range g.graph {
		if _, ok := g.index[v]; !ok { // node that has not been visited
			g.strongConnect(v)
		}
	}
	return g.output
}

// data contains all common data for a single operation.
type data struct {
	graph  map[int][]int
	nodes  []node
	stack  []int
	index  map[int]int
	output [][]int
}

// node stores data for a single vertex in the connection process.
type node struct {
	lowLink int
	stacked bool
}

// strongConnect runs Tarjan's algorithm recursively and outputs a grouping of
// strongly connected vertices.
func (data *data) strongConnect(v int) *node {
	index := len(data.nodes)
	data.index[v] = index
	data.stack = append(data.stack, v)
	data.nodes = append(data.nodes, node{lowLink: index, stacked: true})
	node := &data.nodes[index]

	for _, w := range data.graph[v] {
		i, seen := data.index[w]
		if !seen {
			n := data.strongConnect(w)
			if n.lowLink < node.lowLink {
				node.lowLink = n.lowLink
			}
		} else if data.nodes[i].stacked {
			if i < node.lowLink {
				node.lowLink = i
			}
		}
	}

	// finish backtracking
	if node.lowLink == index {
		var vertices []int
		i := len(data.stack) - 1
		for {
			w := data.stack[i]
			stackIndex := data.index[w]
			data.nodes[stackIndex].stacked = false
			vertices = append(vertices, w)
			if stackIndex == index {
				break
			}
			i--
		}
		data.stack = data.stack[:i]
		data.output = append(data.output, vertices)
	}

	return node
}