package splashmerkle

type Root struct {
	leaf *Node
}

func (root *Root) Bytes() []byte {
	return root.leaf.H
}
