package chord

import (
	// "fmt"
	"os"
	"sync"

	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/models"
)

var local_node = &Node{
	ID:            0,
	Address:       os.Getenv("NODE_ADDRESS"),
	Successor:     0,
	Predecessor:   0,
	FingerTable:   []map[int]int{},
	SuccessorList: []int{},
	Data:          map[int]string{},
}

// Node represents a single node in the Chord network.
type Node struct {
	ID            int    // Unique identifier for the node
	Address       string // Network address of the node
	Successor     int    // Pointer to the successor node in the network
	Predecessor   int    // Pointer to the predecessor node (optional, but can be useful)
	FingerTable   []map[int]int
	SuccessorList []int
	Data          map[int]string
	RWLock        sync.RWMutex
}

func (n *Node) NewLeaveRingMessage() *models.LeaveRingMessage {
	succListLen := len(n.SuccessorList)
	lastNodeInSuccessorList := n.SuccessorList[succListLen-1]
	return &models.LeaveRingMessage{
		DepartingNodeID: n.ID,
		Keys:            n.Data,
		NewSuccessor:    lastNodeInSuccessorList,
		NewPredecessor:  n.Predecessor,
	}
}
