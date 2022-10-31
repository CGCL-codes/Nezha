package core

import (
	"Nezha/graph"
	"math"
	"reflect"
)

type ArcNode struct {
	adj  int
	next *ArcNode
	isAborted bool
}

type VNode struct {
	data  int
	first *ArcNode
	isSorted bool
}

type AlGraph struct {
	vertices       []*VNode
	vexNum, arcNum int
	outDegree 	   []int
	inDegree 	   []int
}

func (al *AlGraph) Init(gSlice [][]int) {
	// init the adjacency list
	for i, v := range gSlice {
		degreeNum := 0
		node := &VNode{data: i}
		arcNode := new(ArcNode)
		al.vexNum++
		for j, vv := range v {
			if j == 0 {
				arcNode = &ArcNode{adj: vv}
				node.first = arcNode
			} else {
				arcNode.next = &ArcNode{adj: vv}
				arcNode = arcNode.next
			}
			al.arcNum++
			degreeNum++
		}
		al.outDegree = append(al.outDegree, degreeNum)
		al.vertices = append(al.vertices, node)
	}

	// init the graph in-degree
	var inDegree = make(map[int]int)

	for _, ver := range al.vertices {
		tmpVer := ver.first
		for {
			if tmpVer == nil {
				break
			}
			inDegree[tmpVer.adj]++
			tmpVer = tmpVer.next
		}
	}

	for i:=0; i<al.vexNum; i++ {
		al.inDegree = append(al.inDegree, inDegree[i])
	}
}

// RebuildGraph remove all vertices and edges of aborted transactions
func (al *AlGraph) RebuildGraph(abortedTx []int) {
	for _, v := range al.vertices {
		data := v.data
		if isExistForInt(abortedTx, data) {
			farc := v.first
			for {
				if farc == nil {
					break
				}
				al.inDegree[farc.adj]--
				farc = farc.next
			}
		} else {
			farc := v.first
			for {
				if farc == nil {
					break
				}
				if isExistForInt(abortedTx, farc.adj) {
					farc.isAborted = true
				}
				farc = farc.next
			}
		}
	}
}

func (al *AlGraph) BasicTopologicalSort() []int {
	var emptyInDegreeVex []int
	var topologicalOrder []int

	for i, v := range al.inDegree {
		if v == 0 {
			emptyInDegreeVex = append(emptyInDegreeVex, i)
		}
	}

	count := 0

	for len(emptyInDegreeVex) > 0 {
		top := emptyInDegreeVex[0]
		topologicalOrder = append(topologicalOrder, top)

		farc := al.vertices[top].first
		for {
			if farc == nil {
				break
			}
			if farc.isAborted {
				farc = farc.next
				continue
			}
			al.inDegree[farc.adj]--
			if al.inDegree[farc.adj] == 0 {
				emptyInDegreeVex = append(emptyInDegreeVex, farc.adj)
			}
			farc = farc.next
		}
		count++
		emptyInDegreeVex = emptyInDegreeVex[1:]
	}

	return topologicalOrder
}

// AdvancedTopologicalSort determine the unique sort order among conflict queues
func (al *AlGraph) AdvancedTopologicalSort() []int {
	var zeroDegreeVex []int
	var topologicalOrder []int
	var top int
	var min = math.MaxInt64

	for i, v := range al.inDegree {
		if v == 0 {
			zeroDegreeVex = append(zeroDegreeVex, i)
		}
		// find the minimum in-degree
		if v < min {
			min = v
		}
	}

	if len(zeroDegreeVex) == 0 {
		top = al.findMaxNode(min)
		zeroDegreeVex = append(zeroDegreeVex, top)
	}

	for len(zeroDegreeVex) > 0 {
		top = zeroDegreeVex[0]
		topologicalOrder = append(topologicalOrder, top)
		al.vertices[top].isSorted = true

		farc := al.vertices[top].first
		for {
			if farc == nil {
				break
			}
			if al.vertices[farc.adj].isSorted {
				farc = farc.next
				continue
			}
			al.inDegree[farc.adj]--
			if al.inDegree[farc.adj] == 0 {
				zeroDegreeVex = append(zeroDegreeVex, farc.adj)
			}
			farc = farc.next
		}

		zeroDegreeVex = zeroDegreeVex[1:]

		// not yet finished
		if len(topologicalOrder) < len(al.vertices) {
			min = math.MaxInt64
			for i, v := range al.inDegree {
				if al.vertices[i].isSorted {
					continue
				}
				if v < min {
					min = v
				}
			}
		}

		if min > 0 {
			top = al.findMaxNode(min)
			zeroDegreeVex = append(zeroDegreeVex, top)
		}
	}

	return topologicalOrder
}

// findMaxNode find the node with the least in-degrees and the most out-degrees (in the case of a cycle)
func (al *AlGraph) findMaxNode(min int) int {
	var inDegreeVex []int
	var maxNode int

	// if exists cycles
	for i, v := range al.inDegree {
		if al.vertices[i].isSorted {
			continue
		}
		if v == min {
			inDegreeVex = append(inDegreeVex, i)
		}
	}

	// find the maximum out-degree
	max := 0
	for _, v2 := range inDegreeVex {
		if al.outDegree[v2] > max {
			max = al.outDegree[v2]
		}
	}
	for _, v3 := range inDegreeVex {
		if al.outDegree[v3] == max {
			maxNode = v3
			break
		}
	}

	return maxNode
}

func (al *AlGraph) RemoveCycles(gSlice [][]int) []int {
	var abortedTx []int
	var appears = make(map[int][]int)

	g := graph.CreateJohnsonGraph(gSlice)
	cycles := graph.FindAllCycles(g)

	for i, c := range cycles {
		for _, v := range c {
			if !isExistForInt(appears[v], i) {
				appears[v] = append(appears[v], i)
			}
		}
	}

	// find the vertex contained in most cycles
	if len(cycles) > 0 {
		max := 0
		maxVertex := 0
		for v := range appears {
			length := len(appears[v])
			if length > max {
				max = length
			}
		}
		for k := range appears {
			length := len(appears[k])
			if length == max {
				maxVertex = k
				break
			}
		}

		abortedTx = append(abortedTx, maxVertex)
		// if remaining some cycles
		if len(appears[maxVertex]) < len(cycles) {
			for i := range cycles {
				if !isExistForInt(appears[maxVertex], i) &&
					isExistForInt(abortedTx, cycles[i][0]) {
					abortedTx = append(abortedTx, cycles[i][0])
				}
			}
		}
	}

	return abortedTx
}

func (al *AlGraph) GetAllCycles(conn []int) map[int][]int {
	var cycles = make(map[int][]int)
	var path = make([]int, len(conn)+1)
	var visit = make(map[int]int)
	pathNum := 0

	existCycle := func(start int, end int) bool {
		flag := 0
		length := end - start

		for i:=0; i<pathNum; i++ {
			if len(cycles[i]) == length {
				flag = 0
				for j:=start; j<end; j++ {
					e := path[j]
					for k:=0; k<len(cycles[i]); k++ {
						if cycles[i][k] == e {
							flag++
						}
					}
				}
				if flag == length {
					return true
				}
			}
		}
		return false
	}

	var findAllCycles func(v int, k int)

	findAllCycles = func(v int, k int) {
		var start int
		visit[v] = 1
		path[k] = v
		node := al.vertices[v]
		ver := node.first

		for {
			if ver == nil {
				break
			}
			if !isExistForInt(conn, ver.adj) {
				ver = ver.next
				continue
			}

			if visit[ver.adj] == 1 {
				for i:=0; i<k; i++ {
					if path[i] == ver.adj {
						start = i
					}
				}
				if existCycle(start, k+1) {
					for i:=start; i<=k; i++ {
						cycles[pathNum] = append(cycles[pathNum], path[i])
					}
					pathNum++
				}
			} else {
				findAllCycles(ver.adj, k+1)
			}
			ver = ver.next
		}

		visit[v] = 0
		path[k] = 0
	}

	for _, n := range conn {
		visit[n] = 0
	}

	for _, n := range conn {
		if visit[n] == 0 {
			findAllCycles(n, 0)
		}
	}

	return cycles
}

func BuildConflictGraph(txs [][]*RWNode) [][]int {
	var gSlice = make([][]int, len(txs))

	for i, v := range txs {
		var rKeys [][]byte
		var wKeys [][]byte

		for _, rw := range v {
			if rw.Label == "r" {
				rKeys = append(rKeys, rw.RWSet.Key)
			} else {
				wKeys = append(wKeys, rw.RWSet.Key)
			}
		}

		for j:=i+1; j<len(txs); j++ {
			for _, rw := range txs[j] {
				if rw.Label == "w" && isExistForByte(rKeys, rw.RWSet.Key) {
					// build r-w dependency
					if !isExistForInt(gSlice[i], j) {
						gSlice[i] = append(gSlice[i], j)
					}
				} else if rw.Label == "w" && isExistForByte(wKeys, rw.RWSet.Key) {
					// build w-w dependency
					if !isExistForInt(gSlice[i], j) {
						gSlice[i] = append(gSlice[i], j)
					}
				}
			}
		}

		for k:=i-1; k>=0; k-- {
			for _, rw := range txs[k] {
				if rw.Label == "w" && isExistForByte(rKeys, rw.RWSet.Key) {
					// build r-w dependency
					if !isExistForInt(gSlice[i], k) {
						gSlice[i] = append(gSlice[i], k)
					}
				}
			}
		}
	}

	return gSlice
}

// light version of building conflict graph
func NewBuildConflictGraph(txs [][]*RWNode) [][]int {
	var gSlice = make([][]int, len(txs))
	var rSet = make(map[string][]int)
	var wSet = make(map[string][]int)

	for i, v := range txs {
		var rKeys []string
		var wKeys []string

		for _, rw := range v {
			if rw.Label == "r" {
				rKey := ConvertByte2String(rw.RWSet.Key)
				rKeys = append(rKeys, rKey)
				rSet[rKey] = append(rSet[rKey], i)
			} else {
				wKey := ConvertByte2String(rw.RWSet.Key)
				wKeys = append(wKeys, wKey)
				wSet[wKey] = append(wSet[wKey], i)
			}
		}

		for j:=i+1; j<len(txs); j++ {
			for _, rw := range txs[j] {
				if rw.Label == "w" {
					rwKey := ConvertByte2String(rw.RWSet.Key)
					if isExistForString(wKeys, rwKey) {
						// build w-w dependency
						gSlice[i] = append(gSlice[i], j)
					}
				}
			}
		}
	}

	for i, v := range txs {
		for _, rw := range v {
			if rw.Label == "r" {
				rwKey := ConvertByte2String(rw.RWSet.Key)
				for _, v := range wSet[rwKey] {
					// build r-w dependency
					if v != i && !isExistForInt(gSlice[i], v) {
						gSlice[i] = append(gSlice[i], v)
					}
				}
			}
		}
	}

	return gSlice
}

func isExistForInt(slice []int, n int) bool {
	for _, e := range slice {
		if reflect.DeepEqual(n, e) {
			return true
		} else {
			continue
		}
	}
	return false
}

func isExistForByte(slice [][]byte, n []byte) bool {
	for _, e := range slice {
		if reflect.DeepEqual(n, e) {
			return true
		} else {
			continue
		}
	}
	return false
}

func isExistForString(slice []string, n string) bool {
	for _, e := range slice {
		if reflect.DeepEqual(n, e) {
			return true
		} else {
			continue
		}
	}
	return false
}
