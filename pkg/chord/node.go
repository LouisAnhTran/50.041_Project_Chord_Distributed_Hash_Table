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

func (n *Node) UpdateData(data map[int]string) {
	defer n.RWLock.Unlock()

	n.RWLock.Lock()
	for key, value := range data {
		n.Data[key] = value
	}
}

func GetLocalNode() *Node {
	return local_node
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
		DepartingNodeID:   n.ID,
		Keys:              n.Data,
		SuccessorListNode: lastNodeInSuccessorList,
		NewSuccessor:      n.Successor,
		NewPredecessor:    n.Predecessor,
	}
}
