package chord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/config"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/models"
	"github.com/gin-gonic/gin"
)

// JoinNetwork is an exported function (capitalized) that allows a node to join the network.
func JoinNetwork(nodeID string) {
	fmt.Printf("Node %s is joining the network.\n", nodeID)
}

func InitChordRingStructure() {
	// go routine sleep for 3 seconds, giving sufficient time for all nodes to set up routes
	time.Sleep(3 * time.Second) // Pauses for 2 seconds

	for i := range config.NodeAddresses {
		node_address := config.NodeAddresses[i]

		if node_address != os.Getenv("NODE_ADDRESS") {
			url := fmt.Sprintf("http://%s/node_identifier", node_address)

			resp, err := http.Get(url)

			if err != nil {
				log.Fatalf("Failed to send request to %s: %v", node_address, err)
				continue
			}

			// Read and print the response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("Failed to read response: %v", err)
				continue
			}

			fmt.Printf("Response from %s: %s \n", node_address, string(body))

			// Unmarshal JSON into the Response struct
			var response models.ResponseNodeIdentifier

			if err := json.Unmarshal(body, &response); err != nil {
				log.Fatalf("Failed to parse JSON: %v", err)
				continue
			}

			// Now we can access the Data field directly
			fmt.Printf("Response from %s: %d\n", node_address, response.Data)

			// Append data to slice of node map
			config.AllNodeMap[response.Data] = node_address
		}
	}

	// Append data to slice of node map
	config.AllNodeMap[HashToRange(os.Getenv("NODE_ADDRESS"))] = os.Getenv("NODE_ADDRESS")

	// Populate all node ID
	populate_all_node_id_and_sort()

	// Set node property
	local_node.ID = HashToRange(os.Getenv("NODE_ADDRESS"))

	fmt.Println("print all sorted node ids: ", config.AllNodeID)

	// determine successor
	var successor = FindSuccessorForNode()
	local_node.Successor = successor
	fmt.Println("my id ", local_node.ID, " my successor is: ", local_node.Successor)

	// determine predecessor
	var predecessor = FindPredecessorForNode()
	local_node.Predecessor = predecessor
	fmt.Println("my id ", local_node.ID, " my predecessor is: ", local_node.Predecessor)

	// Populate Finger Table
	PopulateFingerTable()
	fmt.Println("my all node id: ", config.AllNodeID, " || my id is: ", local_node.ID, " || my finger table: ", local_node.FingerTable, " || my map: ", config.AllNodeMap, " || my successor: ", local_node.Successor, " || predecessor: ", local_node.Predecessor)
}

func HandleStoreData(req models.StoreDataRequest, c *gin.Context) {
	convert_data_to_identifier := HashToRange(req.Data)

	fmt.Println("the key of this data is: ", convert_data_to_identifier)

	// we need to find the successor for this key first
	// create new reques
	find_successor_request := models.FindSuccessorRequest{
		Key: convert_data_to_identifier,
	}

	// prepare request
	jsonData, err := json.Marshal(find_successor_request)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error - Error marshaling JSON",
			Error:   err.Error(),
		})
		return
	}

	closest_preceding_node := find_closest_preceding_node(convert_data_to_identifier)

	fmt.Println("closest preceding node: ", closest_preceding_node)

	address_closest_proceed_node := config.AllNodeMap[closest_preceding_node]

	url := fmt.Sprintf("http://%s/find_successor", address_closest_proceed_node)

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
	var responseBody models.FindSuccessorSuccessResponse
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error - Error decoding response body",
			Error:   err.Error(),
		})
		return
	}

	fmt.Println("Response from server:", responseBody)

	node_to_store_data := responseBody.Successor
	fmt.Println("id of node to store data, successor of data: ", node_to_store_data)

	node_to_store_data_address := config.AllNodeMap[node_to_store_data]
	fmt.Println("the addresss of node to store data: ", node_to_store_data_address)

	// now we send request requesting this node to store data
	// TO DO
	send_request_to_successor_for_storing_data(node_to_store_data_address, req.Data, convert_data_to_identifier, c)

}

func HandleFindSuccessor(req models.FindSuccessorRequest, c *gin.Context) {
	// this normal case where the node does not have the highest id
	if req.Key > local_node.ID && req.Key <= local_node.Successor {
		fmt.Println("I found successor: ", local_node.Successor)
		c.JSON(http.StatusOK, models.FindSuccessorSuccessResponse{
			Successor: local_node.Successor,
			Message:   "Successfully found successor",
		})
		return
	}

	// this special case if the node has the highest id
	if req.Key > config.AllNodeID[len(config.AllNodeID)-1] || req.Key < config.AllNodeID[0] {
		fmt.Println("I found successor: ", config.AllNodeID[0])
		c.JSON(http.StatusOK, models.FindSuccessorSuccessResponse{
			Successor: config.AllNodeID[0],
			Message:   "Successfully found successor"})
		return
	}

	closest_preceding_node := find_closest_preceding_node(req.Key)

	fmt.Println("closest preceding node: ", closest_preceding_node)

	address_closest_proceed_node := config.AllNodeMap[closest_preceding_node]

	// prepare request
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error - Error marshaling JSON",
			Error:   err.Error(),
		})
		return
	}

	url := fmt.Sprintf("http://%s/find_successor", address_closest_proceed_node)

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
	var responseBody models.FindSuccessorSuccessResponse
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error - Error decoding response body",
			Error:   err.Error(),
		})
		return
	}

	fmt.Println("Response from server:", responseBody)

	c.JSON(http.StatusOK, models.FindSuccessorSuccessResponse{
		Message:   "Successfully find successor",
		Successor: responseBody.Successor,
	})

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

func HandleInternalStoreData(request models.InternalStoreDataRequest, c *gin.Context) {
	// simply store data to to this machine
	local_node.Data[request.Key] = request.Data

	// machine data storage after storing the data
	fmt.Println("hash table after storing key-value pair: ", local_node.Data)
	fmt.Println()

	c.JSON(http.StatusOK, models.InternalStoreDataResponse{
		Message: "Stored data successfully",
	})
}

func HandleRetrieveData(key int, c *gin.Context) {
	// we need to find the successor for this key first
	// create new reques
	find_successor_request := models.FindSuccessorRequest{
		Key: key,
	}

	// prepare request
	jsonData, err := json.Marshal(find_successor_request)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error - Error marshaling JSON",
			Error:   err.Error(),
		})
		return
	}

	closest_preceding_node := find_closest_preceding_node(key)

	fmt.Println("closest preceding node: ", closest_preceding_node)

	address_closest_proceed_node := config.AllNodeMap[closest_preceding_node]

	url := fmt.Sprintf("http://%s/find_successor", address_closest_proceed_node)

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
	var responseBody models.FindSuccessorSuccessResponse
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body", err)
		c.JSON(http.StatusInternalServerError, models.FindSuccessorErrorResponse{
			Message: "Server error - Error decoding response body",
			Error:   err.Error(),
		})
		return
	}

	fmt.Println("Response from server:", responseBody)

	node_to_retrieve_data := responseBody.Successor
	fmt.Println("id of node to store data, successor of data: ", node_to_retrieve_data)

	node_to_retrieve_data_address := config.AllNodeMap[node_to_retrieve_data]
	fmt.Println("the addresss of node to store data: ", node_to_retrieve_data_address)

	send_request_to_successor_for_retrieving_data(node_to_retrieve_data_address, key, c)
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

func HandleInternalRetrieveData(key int, c *gin.Context) {
	value, exists := local_node.Data[key]
	if exists {
		// machine data storage after storing the data
		fmt.Println("value of the key ", key, " is: ", value)
		fmt.Println()

		c.JSON(http.StatusOK, models.InternalRetrieveDataResponse{
			Message: "Retrieved data successfully",
			Data:    value,
		})
	} else {
		c.JSON(http.StatusNotFound, models.InternalRetrieveDataResponse{
			Message: "Can not find any value with this key",
		})
	}

}

// functions for handling node leave
// message contains departing node's ID, keys, successor and predecessor
func HandleNodeVoluntaryLeave(message models.LeaveRingMessage, c *gin.Context) {
	if message.DepartingNodeID == local_node.Predecessor {
		// store departing node's (predecessor) keys
		if len(message.Keys) > 0 {
			for key, data := range message.Keys {
				local_node.Data[key] = data
			}
		}

		// set new predecessor
		local_node.Predecessor = message.NewPredecessor
	}

	if message.DepartingNodeID == local_node.Successor {
		// remove departing node (successor) from successor list
		deleteFromSuccessorList(message.DepartingNodeID)
		// add new successor (last node in departing node's SuccessorList) to the list
		addToSuccessorList(message.NewSuccessor)
		// update AllNodeID and AllNodeMap by deleting the respective entries
		deleteNodeEntry(message.DepartingNodeID)

		// set new successor
		newSuccessor := local_node.SuccessorList[0]
		local_node.Successor = newSuccessor

		// call stabilization function
		stabilize(config.AllNodeID, config.AllNodeMap)
	}
}

func HandleNodeInvoluntaryLeave() {
	for _, id := range local_node.SuccessorList {
		addr := config.AllNodeMap[id]
		client := &http.Client{Timeout: 3 * time.Second} // timeout for request
		// call health check on each node
		url := fmt.Sprintf("http://%s/health_check", addr)
		resp, err := client.Get(url)

		if err != nil {
			log.Fatalf("Node %d (%s) is unresponsive: %v\n", id, addr, err)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Decode JSON response into a map
			var body map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				log.Fatalf("Node %d (%s): failed to decode response body: %v\n", id, addr, err)
				continue
			}

			// Check the "message" field
			if message, ok := body["message"].(string); ok && message == "good health" {
				fmt.Printf("Node %d (%s) is healthy.\n", id, addr)
				// set this live node as new successor
				local_node.Successor = id
				// call stabilize
				stabilize(config.AllNodeID, config.AllNodeMap)
				break
			} else {
				fmt.Printf("Node %d (%s) returned unexpected message: %v\n", id, addr, body)
			}
		} else {
			fmt.Printf("Node %d (%s) returned unhealthy status: %d\n", id, addr, resp.StatusCode)
		}
	}

	// TODO:
	// Modified closest_preceding_node (in fig 5) searches not only the finger table but also the successor list for the most immediate predecessor of id.
	// If a node fails during the find_successor, the lookup proceeds, after a timeout, by trying the next best predecessor (of key k) among the nodes in the finger table and the successor list (only choose from finger table if no node better precedes successor list's best predecessor than from finger table's).
}

func find_closest_preceding_node(node_id int) int {
	for i := len(local_node.FingerTable) - 1; i >= 0; i-- {
		for k, v := range local_node.FingerTable[i] {
			fmt.Println("k is: ", k, " v: ", v)
			if v < node_id {
				return v
			}
		}
	}
	return local_node.Successor
}

// FindSuccessor is another exported function to find the successor of a given key.
func FindSuccessor(key string) string {
	fmt.Printf("Finding successor for key %s...\n", key)
	// Placeholder return value
	return "SuccessorNode"
}

func Test_pack() {
	fmt.Println("test package successfully !")
}
