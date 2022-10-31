package core

import (
	"fmt"
	"reflect"
)

type RWNode struct {
	RWSet     RWSet
	TransInfo TransInfo
	Label 	  string
	Sequence  int32
	isAssigned bool
}

type TransInfo struct {
	ID		   string
	Timestamp  uint32
}

type RWSet struct {
	Key []byte
	Value []byte
}

func CreateRWNode(id string, time uint32, rAddr [][]byte, rValue [][]byte, wAddr [][]byte, wValue [][]byte) []*RWNode {
	var rwNodes []*RWNode
	// transInfo := TransInfo{ConvertByte2String(transaction.ID), transaction.Header.Timestamp}
	transInfo := TransInfo{id, time}

	// TODO: obtain read&write set of transaction
	for i:=0; i<len(rAddr); i++ {
		rSet := RWSet{rAddr[i], rValue[i]}
		rNode := RWNode{rSet, transInfo, "r", 0, false}
		rwNodes = append(rwNodes, &rNode)
	}

	for j:=0; j<len(wAddr); j++ {
		wSet := RWSet{wAddr[j], wValue[j]}
		wNode := RWNode{wSet, transInfo, "w", 0, false}
		rwNodes = append(rwNodes, &wNode)
	}

	return rwNodes
}

func (rw *RWNode) assignSequence(edge []*RWNode) {
	for _, e := range edge {
		if reflect.DeepEqual(rw, e) {
			continue
		}
		e.Sequence = rw.Sequence
	}
}

func ConvertByte2String(bytes []byte) string {
	newString := fmt.Sprintf("%x", bytes)
	return newString
}
