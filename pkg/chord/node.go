package chord

import (
    // "fmt"
    "os"
)

var local_node = &Node{
    ID: 0,
    Address: os.Getenv("NODE_ADDRESS"),
    Successor: 0,
    Predecessor: 0,
    FingerTable: []map[int]int{},
}

// Node represents a single node in the Chord network.
type Node struct {
    ID       int  // Unique identifier for the node
    Address  string  // Network address of the node
    Successor int  // Pointer to the successor node in the network
    Predecessor int // Pointer to the predecessor node (optional, but can be useful)
    FingerTable []map[int]int
}

// NewNode initializes and returns a new node with the given ID and address.
// func NewNode(id int, address string) *Node {
//     return &Node{
//         ID:      id,
//         Address: address,
//         Successor: nil,   // Set successor to nil until joined
//         Predecessor: nil, // Set predecessor to nil until joined
//     }
// }

// // SetSuccessor sets the successor node for the current node.
// func (n *Node) SetSuccessor(successor *Node) {
//     n.Successor = successor
//     fmt.Printf("Node %d has set its successor to %d \n", n.ID, successor.ID)
// }

// // Join joins the Chord network by connecting to an existing node (if any).
// func (n *Node) Join(existingNode *Node) {
//     if existingNode != nil {
//         n.Successor = existingNode.FindSuccessor(n.ID)
//         fmt.Printf("Node %d joined the network and set successor to %d \n", n.ID, n.Successor.ID)
//     } else {
//         // If no existing node, the node is its own successor (it's the first node in the ring)
//         n.Successor = n
//         fmt.Printf("Node %d started a new network\n", n.ID)
//     }
// }

// // FindSuccessor finds the successor of a given key.
// func (n *Node) FindSuccessor(key int) *Node {
//     // Placeholder implementation for simplicity
//     fmt.Printf("Finding successor for key %d from Node %d \n", key, n.ID)
//     return n.Successor // This would typically involve more Chord logic
// }
