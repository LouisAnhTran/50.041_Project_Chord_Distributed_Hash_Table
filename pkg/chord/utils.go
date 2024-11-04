package chord

import (
    "crypto/sha1"
    "encoding/hex"
    "math/big"
    "sort"
    "github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/config"
    "math"

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
    all_node_id:=[]int{}

    for key := range config.AllNodeMap {
        all_node_id=append(all_node_id, key)
    }
    
    sort.Ints(all_node_id)

    config.AllNodeID=all_node_id

}

func FindSuccessorForNode() int {
    // Initialize index to -1 to indicate not found
    index_node:=0

    for i := range config.AllNodeID {
        if config.AllNodeID[i] == local_node.ID {
            index_node=i
            break
        }
    }

    return config.AllNodeID[(index_node+1) % len(config.AllNodeID)]
}

func FindPredecessorForNode() int {
    // Initialize index to -1 to indicate not found
    index_node:=0

    for i := range config.AllNodeID {
        if config.AllNodeID[i] == local_node.ID {
            index_node=i
            break
        }
    }

    if index_node==0 {
        return config.AllNodeID[len(config.AllNodeID)-1]
    }
    return config.AllNodeID[index_node-1]
}

func PopulateFingerTable() {
    // Initialize index to -1 to indicate not found
   

    for i:=1;i<=config.FingerTableEntry;i++ {
        adjusted_index := (local_node.ID + int(math.Pow(2, float64(i-1)))) % config.HashRange
        successor:=find_successor_for_fingertable_entry(adjusted_index)
        map_adjusted_index:=map[int]int{adjusted_index:successor}
        local_node.FingerTable=append(local_node.FingerTable,map_adjusted_index)
    }
}

func find_successor_for_fingertable_entry(entry_index int) int {
    found_index:=-1

    for i:=1;i<len(config.AllNodeID);i++ {
        if entry_index > config.AllNodeID[i-1] && entry_index <= config.AllNodeID[i] {
            found_index=i
            break
        }
    }

    if found_index == -1 {
        found_index=0
    } 

    return config.AllNodeID[found_index]
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
