package graph

type JohnsonCE interface {
	Run() (int, []bool)
	Unblock(v int, blocked []bool, blockedMap *[][]int)
	FindCycles(component *SCC) ([][]int, [][]int, []int, int)
	FindCyclesRecur(component *SCC, explore []bool, startV, currentV int, blocked []bool, stack *[]int, blockedMap *[][]int, cycles *[][]int, cyclesMap *[][]int, sumArray []int, sum *int) bool
	BreakCycles(component *SCC) []bool
}

type johnsonce struct {
	sccs            []SCC
	graph           *[][]int
	invalidVertices []bool
	nvertices       int
	removeSetSize   int
}

func NewJohnsonCE(graph *[][]int) JohnsonCE {
	return &johnsonce{
		sccs:            nil,
		graph:           graph,
		invalidVertices: make([]bool, len(*graph)),
		nvertices:       len(*graph),
		removeSetSize:   0,
	}
}

func (jce *johnsonce) Run() (int, []bool) {
	invalidVertices := make([]bool, jce.nvertices)

	sccGen := NewTarjanSCC(jce.graph)
	sccGen.SCC()

	for _, scc := range sccGen.GetSCCs() {
		inv := jce.BreakCycles(&scc)
		for _, vertex := range scc.Vertices {
			invalidVertices[vertex] = inv[vertex]
		}
	}

	return jce.removeSetSize, invalidVertices
}

// Run Johnson's algorithm to find all cycles within this strongly connected component
// Return a matrix
// |  |v1|v2| ... |vn|
// |c1| 0| 1| ... | 1|
// |c2| 1| 0| ... | 0|
// Emits one row per cycle indicating whether a given vertex vk is a part of this cycle
//
//
// Also returns a array containing a number (sk) corresponding to the number of cycles which
// contains the given vertex (vk)
// |  |v1|v2| ... |vn|
// |  |s1|s2| ... |sk|
func (jce *johnsonce) FindCycles(component *SCC) ([][]int, [][]int, []int, int) {
	if len((*component).Vertices) == 1 {
		return nil, nil, nil, int(0)
	}

	explore := (*component).Member

	sum := int(0)
	cycles := make([][]int, 0, 1024)
	cyclesMap := make([][]int, 0, 20)
	sumArray := make([]int, jce.nvertices)

	if len((*component).Vertices) == 2 {
		// SCC has only two vertices
		// Must contain a single cycle
		sum += 2
		cycle := make([]int, 2)
		cycleBool := make([]int, jce.nvertices)
		cycle[0], cycle[1] = (*component).Vertices[0], (*component).Vertices[1]
		cycleBool[(*component).Vertices[0]], cycleBool[(*component).Vertices[1]] = 1, 1
		sumArray[(*component).Vertices[0]], sumArray[(*component).Vertices[1]] = 1, 1
		cycles = append(cycles, cycle)
		cyclesMap = append(cyclesMap, cycleBool)
	} else {
		for _, v := range (*component).Vertices {

			stack := make([]int, 0, len((*component).Vertices))
			blocked := make([]bool, jce.nvertices)
			blockedMap := make([][]int, jce.nvertices)

			for i := 0; i < jce.nvertices; i++ {
				blockedMap[i] = make([]int, 0, len((*component).Vertices))
			}

			jce.FindCyclesRecur(component, explore, v, v, blocked, &stack, &blockedMap, &cycles, &cyclesMap, sumArray, &sum)
			explore[v] = false
		}

	}

	return cycles, cyclesMap, sumArray, sum
}

func (jce *johnsonce) FindCyclesRecur(component *SCC, explore []bool, startV, currentV int, blocked []bool, stack *[]int, blockedMap *[][]int, cycles *[][]int, cyclesMap *[][]int, sumArray []int, sum *int) bool {
	foundCycle := false
	*stack = append(*stack, currentV)
	blocked[currentV] = true

	for _, n := range (*(jce.graph))[currentV] {
		if explore[n] == false {
			continue
		} else if n == startV {
			// found a cycle
			foundCycle = true
			cycle := make([]int, 0, len(*stack))
			cycleBool := make([]int, jce.nvertices)
			for _, iter := range *stack {
				(*sum) += 1
				cycleBool[iter] = 1
				sumArray[iter] += 1
				cycle = append(cycle, iter)
			}
			*cycles = append(*cycles, cycle)
			*cyclesMap = append(*cyclesMap, cycleBool)
		} else if blocked[n] == false {
			ret := jce.FindCyclesRecur(component, explore, startV, n, blocked, stack, blockedMap, cycles, cyclesMap, sumArray, sum)
			foundCycle = foundCycle || ret
		}
	}

	if foundCycle {
		// recursive unblock currentV
		jce.Unblock(currentV, blocked, blockedMap)

	} else {
		for _, v := range (*(jce.graph))[currentV] {
			if explore[v] {
				(*blockedMap)[v] = append((*blockedMap)[v], currentV)
			}
		}
	}

	// stack pop()
	*stack = (*stack)[:len(*stack)-1]
	return foundCycle

}

func (jce *johnsonce) Unblock(v int, blocked []bool, blockedMap *[][]int) {
	blocked[v] = false
	for i := 0; i < len((*blockedMap)[v]); i++ {
		n := (*blockedMap)[v][i]
		if blocked[n] {
			jce.Unblock(n, blocked, blockedMap)
		}
	}
	(*blockedMap)[v] = nil
}

func (jce *johnsonce) BreakCycles(component *SCC) []bool {
	invalidVertices := make([]bool, jce.nvertices)
	_, boolMap, sumArray, sum := jce.FindCycles(component)

	// fmt.Println(cycles)

	for sum != int(0) {
		max := func() int {
			max := 0
			for i := 0; i < jce.nvertices; i++ {
				if sumArray[i] > sumArray[max] {
					max = i
				}
			}
			return max
		}()

		for j := 0; j < len(boolMap); j++ {
			if boolMap[j][max] == int(1) {
				for k := 0; k < jce.nvertices; k++ {
					sum -= boolMap[j][k]
					sumArray[k] -= boolMap[j][k]
					boolMap[j][k] = 0
				}
			}
		}
		jce.removeSetSize += 1
		invalidVertices[max] = true
	}
	return invalidVertices
}
