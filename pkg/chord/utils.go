package chord

import (
    "crypto/sha1"
    "encoding/hex"
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

func FindSuccessorForNode(id *int) int {
    // Initialize index to -1 to indicate not found
    index_node:=0

    // Iterate through the slice of maps
    for i, nodeMap := range config.SliceOfNodeMap {
        // Check if the key exists in the map
        if _, exists := nodeMap[*id]; exists {
            index_node = i // Set the index if the key exists
            break     // Exit the loop since we found the key
        }
    }

    return (index_node+1) % len(config.SliceOfNodeMap)
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
