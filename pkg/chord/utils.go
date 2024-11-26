package chord

import (
	"crypto/sha1"
	"encoding/hex"
	"math"
	"math/big"
	"sort"

	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/config"
)

// HashID generates a consistent SHA-1 hash for a given ID, useful for key lookups in the Chord ring.
func HashID(id string) string {
	hasher := sha1.New()
	hasher.Write([]byte(id))
	return hex.EncodeToString(hasher.Sum(nil))
}

func SortByKey(nodeList *[]map[int]string) {
	sort.Slice(*nodeList, func(i, j int) bool {
		// Assuming you want to sort by the integer key (ID)
		return extractKey((*nodeList)[i]) < extractKey((*nodeList)[j])
	})
}

func extractKey(node map[int]string) int {
	for key := range node {
		return key
	}
	return 0
}

func populate_all_node_id_and_sort() {
	all_node_id := []int{}

	for key := range config.AllNodeMap {
		all_node_id = append(all_node_id, key)
	}

	sort.Ints(all_node_id)

	config.AllNodeID = all_node_id

}

func FindSuccessorForNode() int {
	// Initialize index to -1 to indicate not found
	index_node := 0

	for i := range config.AllNodeID {
		if config.AllNodeID[i] == local_node.ID {
			index_node = i
			break
		}
	}

	return config.AllNodeID[(index_node+1)%len(config.AllNodeID)]
}

func FindPredecessorForNode() int {
	// Initialize index to -1 to indicate not found
	index_node := 0

	for i := range config.AllNodeID {
		if config.AllNodeID[i] == local_node.ID {
			index_node = i
			break
		}
	}

	if index_node == 0 {
		return config.AllNodeID[len(config.AllNodeID)-1]
	}
	return config.AllNodeID[index_node-1]
}

func PopulateFingerTable() {
	// Initialize index to -1 to indicate not found

	for i := 1; i <= config.FingerTableEntry; i++ {
		adjusted_index := (local_node.ID + int(math.Pow(2, float64(i-1)))) % config.HashRange
		successor := find_successor_for_fingertable_entry(adjusted_index)
		map_adjusted_index := map[int]int{adjusted_index: successor}
		local_node.FingerTable = append(local_node.FingerTable, map_adjusted_index)
	}
}

func find_successor_for_fingertable_entry(entry_index int) int {
	found_index := -1

	for i := 1; i < len(config.AllNodeID); i++ {
		if entry_index > config.AllNodeID[i-1] && entry_index <= config.AllNodeID[i] {
			found_index = i
			break
		}
	}

	if found_index == -1 {
		found_index = 0
	}

	return config.AllNodeID[found_index]
}

// function to update finger table; used in stabilize function
func fixFingerTable() {
	// TODO
}

// Populates successor list with size log(N)
func PopulateSuccessorList() {
	numOfNodes := len(config.AllNodeID)
	listLength := int(math.Log2(float64(numOfNodes)))
	currentPos := 0

	// find position of local_node in the ring structure (represented by AllNodeID)
	for i, v := range config.AllNodeID {
		if v == local_node.ID {
			currentPos = i
		}
	}

	for i := 0; i < listLength; i++ {
		if currentPos+1 < numOfNodes {
			currentPos += 1
		} else {
			// wrap-around
			currentPos = 0
		}
		successorId := config.AllNodeID[currentPos]
		local_node.SuccessorList = append(local_node.SuccessorList, successorId)
	}
}

// getter and setter using RWMutex since more reads are expected than writes
func getSuccessorList() []int {
	local_node.RWLock.RLock()
	defer local_node.RWLock.RUnlock()

	return local_node.SuccessorList
}

// function to update successor list; used in stabilize function
func updateSuccessorList() {
	local_node.RWLock.Lock()
	defer local_node.RWLock.Unlock()

	// TODO: reconcile successor list with successor and prepend successor to list
	// Note: this is useless for node voluntary leave case (for current design)
}

// safely appends new successor id to successor list
func addToSuccessorList(nodeId int) {
	local_node.RWLock.Lock()
	defer local_node.RWLock.Unlock()

	local_node.SuccessorList = append(local_node.SuccessorList, nodeId)
}

// safely modifies successor list that excludes deleted node id
func deleteFromSuccessorList(nodeId int) {
	local_node.RWLock.Lock()
	defer local_node.RWLock.Unlock()

	for i, id := range local_node.SuccessorList {
		if id == nodeId {
			local_node.SuccessorList = append(local_node.SuccessorList[:i], local_node.SuccessorList[i+1:]...)
			break
		}
	}
}

// safely deletes node from global node list (AllNodeID & AllNodeMap)
func deleteNodeEntry(nodeId int) {
	local_node.RWLock.Lock()
	defer local_node.RWLock.Unlock()

	// delete from AllNodeID list
	for i, id := range config.AllNodeID {
		if id == nodeId {
			config.AllNodeID = append(config.AllNodeID[:i], config.AllNodeID[i+1:]...)
		}
	}

	// delete from AllNodeMap map
	delete(config.AllNodeMap, nodeId)
}

// function called after node leave/join to maintain consistency across all nodes
func stabilize(allNodeID []int, allNodeMap map[int]string) {
	// TODO
	updateSuccessorList()
	fixFingerTable()
	// propagate updated AllNodeID and AllNodeMap through successors
}

// InRange checks if a target ID is in the range (start, end) on the Chord ring.
func InRange(target, start, end string) bool {
	// Convert strings to big integers for easier comparison
	targetInt := new(big.Int)
	startInt := new(big.Int)
	endInt := new(big.Int)

	targetInt.SetString(target, 16)
	startInt.SetString(start, 16)
	endInt.SetString(end, 16)

	if startInt.Cmp(endInt) < 0 {
		// Standard case where start < end
		return targetInt.Cmp(startInt) > 0 && targetInt.Cmp(endInt) <= 0
	}
	// Wraparound case (end < start)
	return targetInt.Cmp(startInt) > 0 || targetInt.Cmp(endInt) <= 0
}
