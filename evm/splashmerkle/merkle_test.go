package splashmerkle

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"Nezha/evm/splashmerkle/utils"
)

func TestMerkleRoot(t *testing.T) {
	randSize := 6 //12 + rand.Intn(20)
	testSet := make([][]byte, randSize)
	for i := 0; i < randSize; i++ {
		data := make([]byte, 10)
		if _, err := rand.Read(data); err != nil {
			t.Error("failed to create random data set")
		}
		h := utils.Hash(data)
		testSet[i] = h[:]
	}

	tree := Tree{}
	tree.ConstructTree(testSet)

	if len(tree.Nodes) <= tree.Inputs {
		t.Error("Tree contains too few Nodes", len(tree.Nodes), tree.Inputs)
	}

	if tree.Root.leaf == nil {
		t.Error("Failed to generate merkle root")
	}

	for i := 0; i < randSize; i++ {
		if !tree.IncludesInput(testSet[i]) {
			t.Error("Tree doesn't include test set")
		}
	}

	leafA := tree.Nodes[0]
	leafB := tree.Nodes[1]
	hash := utils.Hash(append(leafA.H, leafB.H...))

	if !utils.CheckByteEq(hash[:], tree.Nodes[randSize].H) {
		for index := 0; index < len(tree.Nodes); index++ {
			t.Log("hash", index, hex.EncodeToString(tree.Nodes[index].H))
		}
		t.Log("target", hex.EncodeToString(hash[:]))
		t.Log("leafA.Sib", hex.EncodeToString(leafA.GetSibling().H))
		t.Log("leafB    ", hex.EncodeToString(leafB.H))
		t.Error("first leaf in layer 1 not equal to the hash of first two inputs.")
	}

	proof, err := tree.GenerateProofFor(testSet[0]) //4
	if err != nil {
		t.Error("Error generatign merkletree proof", err)
	}

	if !proof.Verify(tree.Root.Bytes()) {
		t.Error("Merkletree proof failed verification")
	}

}
