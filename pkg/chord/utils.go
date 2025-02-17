package chord

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"sort"
	"strings"

	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/config"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/models"
	"github.com/gin-gonic/gin"
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
	for i := 1; i <= config.FingerTableEntry; i++ {
		adjusted_index := (local_node.ID + int(math.Pow(2, float64(i-1)))) % config.HashRange
		successor := find_successor_for_fingertable_entry(adjusted_index)
		map_adjusted_index := map[int]int{adjusted_index: successor}
		local_node.FingerTable = append(local_node.FingerTable, map_adjusted_index)
	}
}

func UpdatePopulateFingerTable() {
	// invalidate current outdated finger table
	local_node.FingerTable = []map[int]int{}

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

func find_node_address_matching_id(node_id int) string {
	for _, node_addr := range config.NodeAddresses {
		if HashToRange(node_addr) == node_id {
			return node_addr
		}
	}
	return "NOT VALID"
}

func send_request_to_successor_for_retrieving_data(node_address string, data_identifier int, c *gin.Context) {
	fmt.Println("create ")
	url := fmt.Sprintf("http://%s/internal_retrieve_data/%d", node_address, data_identifier)

	resp, err := http.Get(url)

	if err != nil {
		log.Fatalf("Failed to send request to %s: %v", node_address, err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error ",
			Error:   err.Error(),
		})
		return
	}

	// Read and print the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error",
			Error:   err.Error(),
		})
		return
	}

	fmt.Printf("Response from %s: %s \n", node_address, string(body))

	// Unmarshal JSON into the Response struct
	var response models.InternalRetrieveDataResponse

	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error",
			Error:   err.Error(),
		})
		return
	}

	// Now we can access the Data field directly
	fmt.Println("Response from ", node_address, ": ", response)

	c.JSON(http.StatusOK, models.InternalRetrieveDataResponse{
		Message: response.Message,
		Data:    response.Data,
	})

}

func GenerateUrl(addr string, endpoint string) string {
	return "http://" + addr + "/" + strings.ReplaceAll(endpoint, "/", "")
}

func send_request_to_successor_for_storing_data(node_address string, data_to_be_store string, data_identifier int, c *gin.Context) {
	url := fmt.Sprintf("http://%s/internal_store_data", node_address)

	// prepare request
	internal_store_data_request := models.InternalStoreDataRequest{
		Data: data_to_be_store,
		Key:  data_identifier,
	}

	jsonData, err := json.Marshal(internal_store_data_request)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error - Error marshaling JSON",
			Error:   err.Error(),
		})
		return
	}

	// Create a new request
	new_req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error - Error creating request",
			Error:   err.Error(),
		})
		return
	}

	// Send the request using the http.Client
	client := &http.Client{}
	response, err := client.Do(new_req)
	if err != nil {
		fmt.Println("Error making POST request", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error - Error making POST request",
			Error:   err.Error(),
		})
		return
	}
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		fmt.Println("Error: received non-200 response status:", response.Status)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error",
		})
		return
	}

	// Read and print the response body
	var responseBody models.InternalStoreDataResponse
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error - Error decoding response body",
			Error:   err.Error(),
		})
		return
	}

	fmt.Println("Response from server:", responseBody)

	c.JSON(http.StatusOK, models.StoreDataResponsee{
		Message: responseBody.Message,
		Key:     data_identifier,
	})
}
