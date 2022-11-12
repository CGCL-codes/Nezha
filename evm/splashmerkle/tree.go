package splashmerkle

import (
	"encoding/hex"
	"math"
	"strconv"

	"Nezha/evm/splashmerkle/utils"
)

type Tree struct {
	Root   Root
	Nodes  []Node
	Inputs int
}

var blankNode = Node{H: []byte{}}

func (tree *Tree) constructBase(input [][]byte) []Node {
	inputLength := len(input)

	// prepare base of tree
	tree.Nodes = make([]Node, inputLength)

	// populate base of tree
	for i := 0; i < inputLength; i++ {
		tree.Nodes[i] = Node{H: input[i]}
		// If index is even then setup relationship with
		// sibling.
		if i%2 != 0 && i > 0 {
			tree.Nodes[i].SiblingLeft = &tree.Nodes[i-1]
			tree.Nodes[i-1].SiblingRight = &tree.Nodes[i]
		} else {
			tree.Nodes[i].isLeaf = true
		}

	}

	tree.Inputs = inputLength
	return tree.Nodes
}

func (tree *Tree) generateLayerFrom(inNodes []Node) []Node {
	inNodesCount := len(inNodes)
	outNodesCount := int(math.Round(float64(inNodesCount) / 2.0))
	outNodes := make([]Node, outNodesCount)
	var hash []byte
	//Loop all inputs
	for i := 0; i < inNodesCount; i++ {

		if i%2 == 0 {
			left := inNodes[i]
			right := left.GetSibling()

			if right == nil {
				buf := utils.Hash(left.H)
				hash = buf[:]
			} else {
				buf := utils.Hash(append(left.H, right.H...))
				hash = buf[:]
			}

			// calculate outNode index
			out := int(i / 2)

			outNodes[out].H = hash
			if out+1 < outNodesCount && out%2 == 0 {
				outNodes[out].isLeaf = true
				outNodes[out].SiblingRight = &outNodes[out+1]
				outNodes[out+1].SiblingLeft = &outNodes[out]

			}

		}
	}

	tree.Nodes = append(tree.Nodes, outNodes...)
	//fmt.Println(tree.ToString())
	return outNodes
}

// ConstructTree will create a merkle tree from the inputs,
// The inputs must be [:32]byte will be the base of the tree
func (tree *Tree) ConstructTree(input [][]byte) {
	//fmt.Println("INPUT LEN", len(input))
	layer := tree.constructBase(input)

	rootFound := false

	for !rootFound {
		newlayer := tree.generateLayerFrom(layer)
		//fmt.Println("LAYER LEN", len(newlayer))
		if len(newlayer) == 1 {
			tree.Root = Root{&newlayer[0]}
			rootFound = true
		}
		if len(newlayer) == 0 {
			tree.Root = Root{&blankNode}
			rootFound = true
		}
		layer = newlayer
	}
}

func (tree *Tree) IncludesInput(input []byte) bool {
	for i := 0; i < tree.Inputs; i++ {
		if utils.CheckByteEq(tree.Nodes[i].H, input) {
			return true
		}
	}
	return false
}

func (tree *Tree) Includes(input []byte) bool {
	for i := 0; i < len(tree.Nodes); i++ {
		if utils.CheckByteEq(tree.Nodes[i].H, input) {
			return true
		}
	}
	return false
}

func (tree *Tree) ToString() string {
	s := "[TREE]\nInputs:" + strconv.Itoa(tree.Inputs) + "\n"

	for i := 0; i < len(tree.Nodes); i++ {
		s += "🍁 Node" + strconv.Itoa(i) + " - " + hex.EncodeToString(tree.Nodes[i].H)
		if tree.Nodes[i].GetSibling() == nil {
			s += " NOSIB "
		}
		s += "\n"
	}

	return s
}
