package chord

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/config"
)

// HashToRange hashes the input string and returns an integer in the range [1, maxRange].
func HashToRange(node_address string) int {
	// Generate SHA-256 hash of the input string
	hash := sha256.Sum256([]byte(node_address))

	// Convert first 8 bytes of the hash to an integer (unsigned)
	hashInt := binary.BigEndian.Uint64(hash[:8])

	// Modulus operation to get a value within 0 to maxRange-1
	result := int(hashInt%uint64(config.HashRange)) + 1

	return result
}
