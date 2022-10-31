package core

import (
	"math"
	"reflect"
	"sort"
	"strings"
)

type QueueGraph struct {
	Queues map[string]*Queue
	Edges  map[string]*Edge
}

type Queue struct {
	rSlice   []*RWNode
	wSlice   []*RWNode
	maxRead  int32 // record the maximum read sequence number
	maxWrite int32 // record the maximum write sequence number
}

type Edge struct {
	set       []*RWNode
	isAborted bool
}

const initialSequence = 10

func CreateGraph(rwNodes [][]*RWNode) *QueueGraph {
	var edges = make(map[string]*Edge)
	var queueArray = make(map[string][]*RWNode)
	var queues = make(map[string]*Queue)

	for _, rw := range rwNodes {
		// build edges
		id := rw[0].TransInfo.ID
		edge := &Edge{rw, false}
		edges[id] = edge

		// create vertices
		for _, n := range rw {
			key := n.RWSet.Key
			newKey := ConvertByte2String(key)
			queueArray[newKey] = append(queueArray[newKey], n)
		}
	}

	for key := range queueArray {
		rSlice, wSlice := initialSorting(queueArray[key])
		newQueue := &Queue{rSlice, wSlice, 0, 0}
		queues[key] = newQueue
	}

	return &QueueGraph{queues, edges}
}

// initialSorting put all the write operations behind the read operations
func initialSorting(queue []*RWNode) ([]*RWNode, []*RWNode) {
	var rSlice []*RWNode
	var wSlice []*RWNode

	for _, rw := range queue {
		if strings.Compare(rw.Label, "r") == 0 {
			rSlice = append(rSlice, rw)
		} else {
			wSlice = append(wSlice, rw)
		}
	}

	return rSlice, wSlice
}

// queuesSort determine the sort order between different queues
func (cq *QueueGraph) QueuesSort() []string {
	/*
			sort each queue:
			1. determine the relative order by examining queues of write operations
		       on which all the read operations depend
			2. if two queues maintain the same order, then compare their addresses
	*/

	// ================ topological sorting version ================ //

	// first sort queues by individual address
	var sortedQueues = make(map[string]int)
	var sortedStrings []string

	for k := range cq.Queues {
		sortedStrings = append(sortedStrings, k)
	}

	sort.Strings(sortedStrings)

	for i, s := range sortedStrings {
		sortedQueues[s] = i
	}

	// find the dependent queue for each queue
	var depQueue = make(map[int][]int)

	for key := range cq.Queues {
		qIndex := sortedQueues[key]
		temp := map[string]struct{}{}

		for _, w := range cq.Queues[key].wSlice {
			rwKey := w.RWSet.Key
			id := w.TransInfo.ID

			for _, n := range cq.Edges[id].set {
				rwKey2 := n.RWSet.Key
				if reflect.DeepEqual(n.Label, "r") && !reflect.DeepEqual(rwKey, rwKey2) {
					newKey := ConvertByte2String(rwKey2)
					temp[newKey] = struct{}{}
				}
			}
		}

		if len(temp) > 0 {
			for kk := range temp {
				index := sortedQueues[kk]
				depQueue[qIndex] = append(depQueue[qIndex], index)
			}
		} else {
			depQueue[qIndex] = []int{}
		}
	}

	var deps [][]int

	for i := 0; i < len(depQueue); i++ {
		deps = append(deps, depQueue[i])
	}

	// generate a graph based on address queue dependencies
	var al AlGraph
	var sequence []string

	al.Init(deps)
	order := al.AdvancedTopologicalSort()

	for _, o := range order {
		if len(sortedStrings) > 0 {
			str := sortedStrings[o]
			sequence = append(sequence, str)
		}
	}

	return sequence
}

// DeSS obtain a deterministic total order
func (cq *QueueGraph) DeSS(sequence []string) map[int32][][]*RWNode {
	var commitOrder = make(map[int32][][]*RWNode)

	for _, s := range sequence {
		queue := cq.Queues[s]
		// sort each queue
		cq.sortInQueue(queue)
	}

	// obtain the final commit order
	for t := range cq.Edges {
		if cq.Edges[t].isAborted {
			continue
		}

		edge := cq.Edges[t].set
		var wNodes []*RWNode

		for _, e := range edge {
			if e.Label == "w" {
				wNodes = append(wNodes, e)
			}
		}

		seq := edge[0].Sequence
		commitOrder[seq] = append(commitOrder[seq], wNodes)
	}

	return commitOrder
}

// sortInQueue sort each read/write unit in each queue
func (cq *QueueGraph) sortInQueue(queue *Queue) {
	var tmpRQueue []*RWNode
	var tmpWQueue []*RWNode
	var tmpWQueue2 = make(map[int32][]*RWNode)

	// first check if the rNode in the queue has a determined order
	for _, r := range queue.rSlice {
		if r.Sequence != 0 {
			id := r.TransInfo.ID
			if cq.Edges[id].isAborted {
				continue
			}
			tmpRQueue = append(tmpRQueue, r)
		}
	}

	// do not have write dependency or the first queue to sort
	if len(tmpRQueue) == 0 {
		for _, r := range queue.rSlice {
			r.Sequence = initialSequence
			r.isAssigned = true
			id := r.TransInfo.ID
			edge := cq.Edges[id].set
			r.assignSequence(edge)
			queue.maxRead = r.Sequence
		}
	} else {
		min := math.MaxInt32

		for _, tr := range tmpRQueue {
			tr.isAssigned = true
			if int(tr.Sequence) < min {
				min = int(tr.Sequence)
			}
			if tr.Sequence > queue.maxRead {
				queue.maxRead = tr.Sequence
			}
		}

		for _, r := range queue.rSlice {
			if r.Sequence != 0 {
				continue
			}
			r.Sequence = int32(min)
			r.isAssigned = true
			id := r.TransInfo.ID
			edge := cq.Edges[id].set
			r.assignSequence(edge)
		}
	}

	// then check if the wNode in the queue has a determined order (for aborting transactions)
	for _, w := range queue.wSlice {
		if w.Sequence != 0 {
			id := w.TransInfo.ID
			if cq.Edges[id].isAborted {
				continue
			}
			if w.Sequence <= queue.maxRead {
				edge := cq.Edges[id].set
				isSame := false
				isBefore := false

				for _, rw := range edge {
					if rw.Label == "r" && rw.isAssigned {
						if reflect.DeepEqual(w.RWSet.Key, rw.RWSet.Key) {
							// has related rNode at the same address
							isSame = true
							break
						} else {
							// read unit is assigned before write unit (r-w)
							isBefore = true
						}
					}
				}

				if isSame {
					tmpWQueue = append(tmpWQueue, w)
				} else if isBefore {
					cq.Edges[id].isAborted = true
				} else {
					// managing wNodes whose serial numbers need to be adjusted
					tmpWQueue2[w.Sequence] = append(tmpWQueue2[w.Sequence], w)
				}
			} else {
				tmpWQueue2[w.Sequence] = append(tmpWQueue2[w.Sequence], w)
			}
		}
	}

	var keys []int

	for key := range tmpWQueue2 {
		keys = append(keys, int(key))
	}

	sort.Ints(keys)

	if queue.maxRead == 0 {
		queue.maxWrite = initialSequence - 1
	} else {
		queue.maxWrite = queue.maxRead
	}

	// first adjust the sequence number of transaction whose rNode and wNode at the same address
	for i, w := range tmpWQueue {
		id := w.TransInfo.ID
		if i == 0 {
			w.Sequence = queue.maxWrite + 1
			queue.maxWrite += 1
			queue.maxRead = queue.maxWrite
			w.isAssigned = true
			edge := cq.Edges[id].set
			w.assignSequence(edge)
		} else {
			cq.Edges[id].isAborted = true
		}
	}

	// assign sequence numbers to wNodes whose sequence numbers are not assigned
	for _, w := range queue.wSlice {
		if w.Sequence != 0 {
			continue
		}
		for i := queue.maxWrite + 1; i < math.MaxInt32; i++ {
			if !isExist(keys, int(i)) {
				w.Sequence = i
				queue.maxWrite = w.Sequence
				w.isAssigned = true
				id := w.TransInfo.ID
				edge := cq.Edges[id].set
				w.assignSequence(edge)
				break
			}
		}
	}

	// update the maximum wNode sequence number
	if len(keys) > 0 {
		maxSeq := keys[len(keys)-1]
		if queue.maxWrite < int32(maxSeq) {
			queue.maxWrite = int32(maxSeq)
		}
	}

	// adjust the remaining wNodes
	for _, n := range keys {
		for i, ww := range tmpWQueue2[int32(n)] {
			if int32(n) > queue.maxRead && i == 0 {
				continue
			}
			ww.Sequence = queue.maxWrite + 1
			queue.maxWrite += 1
			ww.isAssigned = true
			id := ww.TransInfo.ID
			edge := cq.Edges[id].set
			ww.assignSequence(edge)
		}
	}
}

func (cq *QueueGraph) GetRWNums() int {
	sumRW := 0

	for k := range cq.Queues {
		lenR := len(cq.Queues[k].rSlice)
		lenW := len(cq.Queues[k].wSlice)
		sum := lenR + lenW
		sumRW += sum
	}

	return sumRW
}

func (cq *QueueGraph) GetAbortedNums() int {
	count := 0

	edges := cq.Edges
	for k := range edges {
		if edges[k].isAborted {
			count++
		}
	}

	return count
}

func isExist(slice []int, n int) bool {
	for _, e := range slice {
		if reflect.DeepEqual(n, e) {
			return true
		} else {
			continue
		}
	}
	return false
}
