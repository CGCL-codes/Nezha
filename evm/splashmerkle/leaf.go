package splashmerkle

// Node - A merkle tree is constructed from nodes, some nodes may
// be leaves. Each Node contains a pointer to its sibling, accessible
// through `node.GetSibling()`.
type Node struct {
	// H - The hash of this nodes children
	H []byte
	// Left Nodes are leaves
	isLeaf bool
	// Siblings may be nil, for this reason its best to use GetSibling()
	SiblingLeft  *Node
	SiblingRight *Node
}

// GetSibling - Finds and returns a nodes sibling, will return nil
// is there is no registered sibling.
func (n *Node) GetSibling() *Node {
	if n.SiblingLeft != nil {
		return n.SiblingLeft
	}
	if n.SiblingRight != nil {
		return n.SiblingRight
	}
	return nil
}
