package chord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/config"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/models"
	"github.com/gin-gonic/gin"
)

// This funcion is called by a newly joined node
func NewNodeJoinNetwork() {
	// Algorithm for new node joing network - say node 11
	// 1. Select a random node and run find_succesor function based on node 11's ID, let say Successor is Ns
	// 2. Node 11 sets its successor=Ns and notify Ns about it existence by calling function notify
	// 3. Node 11's Succesor run stablization function, and request its successor to execute stabalization function as well

	// 1. FIND SUCCESSOR FOR NEWLY JOINED NODE
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between 1 and 10
	randomNumber := rand.Intn(len(config.NodeAddresses))

	random_node_to_send_request := config.NodeAddresses[randomNumber]

	fmt.Println("random node to send find successor request: ", random_node_to_send_request)

	fmt.Println("joining node id: ", HashToRange(os.Getenv("NODE_ADDRESS")))

	newly_join_node_id := HashToRange(os.Getenv("NODE_ADDRESS"))

	// find succesor for this newly joint node's  id
	// we need to find the successor for this key first
	// create new reques
	find_successor_request := models.FindSuccessorRequest{
		Key: newly_join_node_id,
	}

	// prepare request
	jsonData, err := json.Marshal(find_successor_request)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	url := fmt.Sprintf("http://%s/find_successor", random_node_to_send_request)

	// Create a new request
	new_req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Send the request using the http.Client
	client := &http.Client{}
	response, err := client.Do(new_req)
	if err != nil {
		fmt.Println("Error making POST request", err)
		return
	}
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		fmt.Println("Error: received non-200 response status:", response.Status)
		return
	}

	// Read and print the response body
	var responseBody models.FindSuccessorSuccessResponse
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response body", err)
		return
	}

	fmt.Println("Response from server:", responseBody)

	newly_join_node_successor_id := responseBody.Successor
	fmt.Println("joining node's succesor: ", newly_join_node_successor_id)

	successor_address_derived_from_id := find_node_address_matching_id(newly_join_node_successor_id)

	fmt.Println("joining node's succesor address: ", successor_address_derived_from_id)

	// 2. SET NEWLY JOINT NODE SUCCESSOR
	local_node.Successor = newly_join_node_successor_id

	// 3. NEWLY JOINT NODE NOTIFY ITS SUCCESSOR
	// / create new payload
	notify_request := models.NotifyRequest{
		Key:         newly_join_node_id,
		NodeAddress: os.Getenv("NODE_ADDRESS"),
	}

	// prepare request
	jsonData, err = json.Marshal(notify_request)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	url = fmt.Sprintf("http://%s/notify", successor_address_derived_from_id)

	// Create a new request
	new_req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Send the request using the http.Client
	client = &http.Client{}
	response, err = client.Do(new_req)
	if err != nil {
		fmt.Println("Error making POST request", err)
		return
	}
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		fmt.Println("Error: received non-200 response status:", response.Status)
		return
	}

	// Read and print the response body
	var newResponseBody models.NotifyResponse
	if err := json.NewDecoder(response.Body).Decode(&newResponseBody); err != nil {
		fmt.Println("Error decoding response body", err)
		return
	}

	fmt.Println("Response from server: ", newResponseBody.Message)

	// Once new node receive welcome onboarding response from succesor, it no need to do anything else

}

// This function is called once by all nodes in initial network to construct a Chord ring system
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
	// Populate Successor List
	PopulateSuccessorList()
	fmt.Println("my all node id: ", config.AllNodeID, " || my id is: ", local_node.ID, " || my finger table: ", local_node.FingerTable, " || my successor list: ", local_node.SuccessorList, " || my map: ", config.AllNodeMap, " || my successor: ", local_node.Successor, " || predecessor: ", local_node.Predecessor)
}

func HandleStoreData(req models.StoreDataRequest, c *gin.Context) {
	convert_data_to_identifier := HashToRange(req.Data)

	fmt.Println("the key of this data is: ", convert_data_to_identifier)

	// Convert convert_data_to_identifier to a string for HashToRange
	identifierAsString := fmt.Sprintf("%v", convert_data_to_identifier)
	convert_identifier_to_replica_key := HashToRange(identifierAsString)

	// tbh just put convert_data_to_identifier and convert_identifier_to_replica_key in a list, and for each req in the list do the below

	// Put both keys in a list
	keys := []int{
		convert_data_to_identifier,
		convert_identifier_to_replica_key,
	}

	var prev_preceding_note int
	var hash_threshold int = 3

	// Process each key in the list
	for i, key := range keys {

		closest_preceding_node := find_closest_preceding_node(key)
		
		if i == 0 {
			prev_preceding_note = closest_preceding_node
		} else {
			hash_counter := 0
			// in case we end up with the same preceding node, since then successor would be the same node also
			for prev_preceding_note == closest_preceding_node {
				// this will trigger at hash(hash(hash(hash(hash(original_data)))))
				// which is hash(hash(hash(hash(original_key))))
				// which is hash(hash(hash(initial_replica_key))) => hash counter sets how many rehashes are done here
				if hash_counter >= hash_threshold {
					break
				}
				// Simulate hash computation and update the key
				identifierAsString := fmt.Sprintf("%v", key)
				key = HashToRange(identifierAsString)
				closest_preceding_node = find_closest_preceding_node(key)
				hash_counter++
			}
			prev_preceding_note = closest_preceding_node
		}

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

		// if we succeed this should be two different nodes here
		fmt.Println("closest preceding node: ", closest_preceding_node)

		address_closest_proceed_node := config.AllNodeMap[closest_preceding_node]

		// it's called this
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

		// so from the result get the node
		node_to_store_data := responseBody.Successor
		fmt.Println("id of node to store data, successor of data: ", node_to_store_data)
		if i == 1 {
			fmt.Println("Starting replication.")
		}

		node_to_store_data_address := config.AllNodeMap[node_to_store_data]
		fmt.Println("the addresss of node to store data: ", node_to_store_data_address)

		// now we send request requesting this node to store data
		// that node will call internal store data function
		// if you search logs, you will find two different nodes calling internal_store_data for one request
		send_request_to_successor_for_storing_data(node_to_store_data_address, req.Data, key, c)

		// new todo
		ReplicaMap[convert_data_to_identifier] = convert_identifier_to_replica_key
	}

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

func HandleInternalStoreData(request models.InternalStoreDataRequest, c *gin.Context) {
	// simply store data to to this machine
	local_node.Data[request.Key] = request.Data

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

func HandleSuccessorNotification(request models.NotifyRequest, c *gin.Context) {
	// successor check if newly join node key greater than its predeccesor id
	if local_node.Predecessor > 0 && local_node.Predecessor < request.Key {
		// update Map node and address
		config.AllNodeMap[request.Key] = request.NodeAddress

		// update All node address
		config.AllNodeID = append(config.AllNodeID, request.Key)

		sort.Ints(config.AllNodeID)

		fmt.Println("Successor's updated all node id: ", config.AllNodeID)

		fmt.Println("Successor's updated all node map: ", config.AllNodeMap)

		old_predecessor := local_node.Predecessor

		// update predeccesor to newly joined node
		local_node.Predecessor = request.Key

		c.JSON(http.StatusOK, models.NotifyResponse{
			Message: "Welcome onboard to Chord Ring",
		})

		// after return response to newly join node, the successor must do something else
		// TO DO
		// 1. Successor request its's predeccessor to init stablization
		// 2. Update its own finger table
		// 3. Notify all other nodes in the system about the existence of newly joint node, by sending new node's ID and Address

		// 1. SUCCESSOR REQUEST ITS PREDECESSOR TO INITIALIZE STABALIZATION

		fmt.Println("old predeccesor is: ", old_predecessor)

		address_predecessor := config.AllNodeMap[old_predecessor]

		fmt.Println("old predeccesor's address: ", address_predecessor)

		// create new reques
		stabalization_request := models.StablizationSuccessorRequest{
			Message:     "A new node join our network, please kick off stabalization process",
			Key:         request.Key,
			NodeAddress: request.NodeAddress,
		}

		// prepare request
		jsonData, err := json.Marshal(stabalization_request)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		url := fmt.Sprintf("http://%s/start_stablization", address_predecessor)

		// Create a new request
		new_req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		// Send the request using the http.Client
		client := &http.Client{}
		response, err := client.Do(new_req)
		if err != nil {
			fmt.Println("Error making POST request", err)
			return
		}
		defer response.Body.Close()

		// Check the response status code
		if response.StatusCode != http.StatusOK {
			fmt.Println("Error: received non-200 response status:", response.Status)
			return
		}

		// Read and print the response body
		var responseBody models.StablizationSuccessorResponse
		if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
			fmt.Println("Error decoding response body for stablization", err)
			return
		}

		fmt.Println("Response from predeccesor:", responseBody)

		fmt.Println("My predeccesor already notify newly joint node")

		fmt.Println("At this stage, all nodes has correct S and P pointers, but AllNodeID, AllNodeMap and Finger tables are not consistent and updated")
		// AT THIS POINT, ALL NODES ALREADY HAVE CORRECT POINTERS TO SUCCESSORS AND PREDECCORS
		// HOWEVER, N-2 NODES ARE NOT AWARE OF THE EXISTENCE OF NEW NODE YET, HENCE I NEED TO SEND UPDATED ALL NODE MAP TO MALL

		// 2. UPDATE MY OWN FINGER TABLE
		fmt.Println("Joining's succesor is updating finger table now")
		UpdatePopulateFingerTable()
		fmt.Println("Joining node's Succesor updated Finger Table ", local_node.FingerTable)

		// 3. Notify all other nodes in the system about the existence of newly joint node, by sending new node's ID and Address
		JoiningNodeSuccessorUpdateAllNodesInRingAboutNewNode(request.Key, request.NodeAddress)

		fmt.Println("At this stage, all Nodes pointers have been corrected and all metadata are up to date")

	} else if local_node.Predecessor == 0 {
		// update Map node and address
		config.AllNodeMap[request.Key] = request.NodeAddress

		// update All node address
		config.AllNodeID = append(config.AllNodeID, request.Key)

		sort.Ints(config.AllNodeID)

		fmt.Println("Joining node's updated all node id: ", config.AllNodeID)

		fmt.Println("Joining node's updated all node map: ", config.AllNodeMap)

		// update predeccesor to newly joined node
		local_node.Predecessor = request.Key

		c.JSON(http.StatusOK, models.NotifyResponse{
			Message: "I am the joining node and I already joined the Ring and updated all my pointers",
		})

	} else {
		c.JSON(http.StatusBadRequest, models.NotifyResponse{
			Message: "you reached out to the wrong successor",
		})
	}
}

func JoiningNodeSuccessorUpdateAllNodesInRingAboutNewNode(new_node_key int, new_node_address string) {

	for node_key, node_address := range config.AllNodeMap {
		if node_key != local_node.ID {
			// create new reques
			stabalization_request := models.UpdateMetadataUponNewNodeJoinRequest{
				Key:         new_node_key,
				NodeAddress: new_node_address,
			}

			// prepare request
			jsonData, err := json.Marshal(stabalization_request)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			fmt.Println("I am sending update metadata request to Node key ", node_key, " - Node address ", node_address)

			url := fmt.Sprintf("http://%s/update_metadata", node_address)

			// Create a new request
			new_req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("Error creating request:", err)
				return
			}

			// Send the request using the http.Client
			client := &http.Client{}
			response, err := client.Do(new_req)
			if err != nil {
				fmt.Println("Error making POST request", err)
				return
			}
			defer response.Body.Close()

			// Check the response status code
			if response.StatusCode != http.StatusOK {
				fmt.Println("Error: received non-200 response status:", response.Status)
				return
			}

			// Read and print the response body
			var responseBody models.UpdateMetadataUponNewNodeJoinResponse
			if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
				fmt.Println("Error decoding response body for stablization", err)
				return
			}

			fmt.Println("Response from node for update metadata request:", responseBody)
		}
	}

}

func HandleStartStablization(request models.StablizationSuccessorRequest, c *gin.Context) {
	my_successor := local_node.Successor

	// if new node id less than my succesor, i will update my succesor
	if request.Key < my_successor && request.Key > local_node.ID {
		// update successor to joining node
		local_node.Successor = request.Key

		// update Map node and address
		config.AllNodeMap[request.Key] = request.NodeAddress

		// update All node address
		config.AllNodeID = append(config.AllNodeID, request.Key)

		sort.Ints(config.AllNodeID)

		fmt.Println("Stabalization - My updated all node id: ", config.AllNodeID)

		fmt.Println("Stabalization - My updated all node map: ", config.AllNodeMap)

		// predeccesor notify new join node

		fmt.Println("Predecessor notify newly joint node so the new node can update its predeccesor pointer")
		// / create new payload
		notify_request := models.NotifyRequest{
			Key:         local_node.ID,
			NodeAddress: local_node.Address,
		}

		// prepare request
		jsonData, err := json.Marshal(notify_request)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		url := fmt.Sprintf("http://%s/notify", request.NodeAddress)

		// Create a new request
		new_req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		// Send the request using the http.Client
		client := &http.Client{}
		response, err := client.Do(new_req)
		if err != nil {
			fmt.Println("Error making POST request", err)
			return
		}
		defer response.Body.Close()

		// Check the response status code
		if response.StatusCode != http.StatusOK {
			fmt.Println("Error: received non-200 response status:", response.Status)
			return
		}

		// Read and print the response body
		var newResponseBody models.NotifyResponse
		if err := json.NewDecoder(response.Body).Decode(&newResponseBody); err != nil {
			fmt.Println("Error decoding response body", err)
			return
		}

		fmt.Println("Response from server: ", newResponseBody.Message)

		c.JSON(http.StatusOK, models.StablizationSuccessorResponse{
			Message: "Complete local stabalization",
		})

	} else {
		c.JSON(http.StatusBadRequest, models.StablizationSuccessorResponse{
			Message: "you reached out to the wrong successor",
		})
	}
}

func HandleUpdateMetaData(request models.UpdateMetadataUponNewNodeJoinRequest, c *gin.Context) {
	// update Map node and address
	config.AllNodeMap[request.Key] = request.NodeAddress

	// update All node address
	config.AllNodeID = append(config.AllNodeID, request.Key)

	sort.Ints(config.AllNodeID)

	fmt.Println("Updated all node id upon new node joining: ", config.AllNodeID)

	fmt.Println("Updated all node map upon new node joining: ", config.AllNodeMap)

	// update finder table
	UpdatePopulateFingerTable()
	fmt.Println("Updated finger table upon new node joining: ", local_node.FingerTable)

	c.JSON(http.StatusOK, models.UpdateMetadataUponNewNodeJoinResponse{
		Message: "Successfully updated All Node Map, All Node ID and Finger table with joined node's ID and Address",
	})
}

// Starts voluntary leave sequence for a node
func HandleLeaveSequence() {
	defer os.Exit(0) // Kill node

	localNode := GetLocalNode()
	msg := *localNode.NewLeaveRingMessage()
	fmt.Println("[ Node", localNode.ID, "] LeaveRingMessage:", msg)

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("[ Node", localNode.ID, "] Error during LeaveRingMessage JSON conversion.")
	}

	successorAddr := config.AllNodeMap[localNode.Successor]
	predecessorAddr := config.AllNodeMap[localNode.Predecessor]
	successorUrl := GenerateUrl(successorAddr, "notify_leave")
	predecessorUrl := GenerateUrl(predecessorAddr, "notify_leave")

	// Send POST request to successor
	sRes, err := http.Post(successorUrl, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil {
		fmt.Println("[ Node", localNode.ID, "] Error sending leave request to successor at address", successorAddr)
		fmt.Println(err.Error())
	}
	defer sRes.Body.Close()

	// Send POST request to predecessor
	pRes, err := http.Post(predecessorUrl, "application/json", bytes.NewBuffer(jsonMsg))

	if err != nil {
		fmt.Println("[ Node", localNode.ID, "] Error sending leave request to predecessor at address", predecessorAddr)
		fmt.Println(err.Error())
	}
	defer pRes.Body.Close()

	// The node is assumed to have left at this point. Any faults happening here will be
	// dealt by the successors and predecessors.
}

// Handles leave sequence for a leaving node
func HandleNodeLeave(msg models.LeaveRingMessage) {
	localNode := GetLocalNode()
	if msg.DepartingNodeID == localNode.Successor {
		fmt.Println("[ Node", localNode.ID, "] Replacing successor with Node "+
			strconv.Itoa(msg.DepartingNodeID)+
			"'s successor Node", msg.NewSuccessor)

		localNode.Successor = msg.NewSuccessor
		fmt.Println("[ Node", localNode.ID, "] New successor:", localNode.Successor)

		// Update successor list and node entries
		deleteFromSuccessorList(msg.DepartingNodeID)
		addToSuccessorList(msg.SuccessorListNode)
		deleteNodeEntry(msg.DepartingNodeID)

		successorAddr := config.AllNodeMap[localNode.Successor]
		successorUrl := GenerateUrl(successorAddr, "/leave_data")

		dataMsg := models.NewDataUpdateMessage(msg.DepartingNodeID, localNode.ID, msg.Data)
		dataJsonMsg, err := json.Marshal(dataMsg)
		if err != nil {
			fmt.Println("[ Node", localNode.ID, "] Error during DataUpdateMessage JSON conversion.")
		}

		res, err := http.Post(successorUrl, "application/json", bytes.NewBuffer(dataJsonMsg))
		if err != nil {
			HandleInvoluntaryDeadNode(localNode.Successor)
		} else if res.StatusCode != http.StatusOK {
			fmt.Println("[ Node", localNode.ID, "] Non-200 response received during cycle check initiation from successor at address", successorAddr)
			fmt.Println("[ Node", localNode.ID, "] Aborting...")
			return
		}

		StartReconciliation()
	} else if msg.DepartingNodeID == localNode.Predecessor {
		fmt.Println("[ Node", localNode.ID, "] Replacing predecessor with Node"+
			strconv.Itoa(msg.DepartingNodeID)+
			"'s predecessor Node", msg.NewPredecessor)

		localNode.Predecessor = msg.NewPredecessor
		fmt.Println("[ Node", localNode.ID, "] New predecessor:", localNode.Predecessor)

		deleteNodeEntry(msg.DepartingNodeID)
		if len(msg.Data) > 0 {
			localNode.UpdateData(msg.Data)
		}
	} else {
		fmt.Println("[ Node", localNode.ID, "] LeaveRingMessage received from a Node that is neither the successor nor predecessor of this node.")
		fmt.Println("[ Node", localNode.ID, "]", msg)
		fmt.Println("[ Node", localNode.ID, "] Aborting...")
		return
	}
}

func HandleInvoluntaryLeaveSequence(deadNodeId int) {
	localNode := GetLocalNode()
	// predecessor will double check if broadcasted dead node is actually dead
	for _, id := range localNode.SuccessorList {
		successorAddr := config.AllNodeMap[id]
		// call health check on each node

		fmt.Println("[ Node", GetLocalNode().ID, "] Checking for node liveness: Node", id)
		url := fmt.Sprintf("http://%s/health_check", successorAddr)
		resp, err := http.Get(url)

		if err != nil {
			fmt.Printf("Node %d (%s) is unresponsive: %v\n", id, successorAddr, err)
			// double check passes, remove dead node from AllNodeID
			if id == deadNodeId {
				deleteNodeEntry(id)
			}
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Decode JSON response into a map
			var body map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				fmt.Println("Error decoding response body", err)
				continue
			}

			// Check the "message" field
			if message, ok := body["message"].(string); ok && message == "good health" {
				fmt.Printf("Node %d (%s) is healthy.\n", id, successorAddr)

				// if live node is current successor, end operation
				// else set this live node as new successor
				if localNode.Successor == id {
					fmt.Printf("Node %d: Current successor is still alive. Aborting operation...\n", localNode.ID)
					return
				} else {
					localNode.Successor = id
				}

				// notify new successor
				msg := models.InvoluntaryLeaveMessage{NewPredecessor: localNode.ID}

				jsonMsg, jsonErr := json.Marshal(msg)
				if jsonErr != nil {
					fmt.Println("[ Node", localNode.ID, "] JSON error during relink for involuntary node leave process.")
				}

				successorUrl := GenerateUrl(successorAddr, "notify_new_successor")

				res, err := http.Post(successorUrl, "application/json", bytes.NewBuffer(jsonMsg))

				if err != nil {
					// new successor might have suddenly died, skip to the next successor
					fmt.Printf("Successor %d suddenly died. Trying next successor...\n", id)
					continue
				} else if res.StatusCode != http.StatusOK {
					fmt.Println("[ Node", localNode.ID, "] Non-200 response received during relink to successor at address", successorAddr, "for involuntary node leave process.")
					fmt.Println("[ Node", localNode.ID, "] Aborting...")
					return
				}
				break
			} else {
				fmt.Printf("Node %d (%s) returned unexpected message: %v\n", id, successorAddr, body)
			}
		} else {
			fmt.Printf("Node %d (%s) returned unhealthy status: %d\n", id, successorAddr, resp.StatusCode)
		}
	}

	// TODO:
	// Modified closest_preceding_node (in fig 5) searches not only the finger table but also the successor list for the most immediate predecessor of id.
	// If a node fails during the find_successor, the lookup proceeds, after a timeout, by trying the next best predecessor (of key k) among the nodes in the finger table and the successor list (only choose from finger table if no node better precedes successor list's best predecessor than from finger table's).
}

func HandleNodeInvoluntaryLeave(msg models.InvoluntaryLeaveMessage) {
	newPredecessor := msg.NewPredecessor
	localNode := GetLocalNode()
	localNode.Predecessor = newPredecessor
	StartReconciliation()
}

// Contacting Node checks whether the dead node is its successor
func HandleInvoluntaryDeadNode(deadNodeId int) {
	localNode := GetLocalNode()
	if deadNodeId == localNode.Successor {
		HandleInvoluntaryLeaveSequence(deadNodeId)
	} else {
		// broadcast dead node id to let its corresponding predecessor to handle it
		BroadcastDeadNode(deadNodeId)
	}
}

func BroadcastDeadNode(deadNodeId int) {
	localNode := GetLocalNode()
	localBCQueue := []int{deadNodeId}
	// trackers to prevent duplication
	broadcastTracker := map[int]bool{} // track broadcasted nodes
	queueTracker := map[int]bool{}     // track queued nodes
	queueTracker[deadNodeId] = true

	for len(localBCQueue) > 0 {
		currentDeadNode := localBCQueue[0]
		localBCQueue = localBCQueue[1:]

		msg := models.BroadcastMessage{DeadNode: currentDeadNode}
		jsonMsg, jsonErr := json.Marshal(msg)
		if jsonErr != nil {
			fmt.Println("[ Node", localNode.ID, "] JSON error during broadcast for involuntary node leave process.")
		}

		for _, nodeId := range config.AllNodeID {
			if nodeId == currentDeadNode || broadcastTracker[nodeId] {
				// skip nodes that are known to be dead and broadcasted
				continue
			}
			nodeAddr := config.AllNodeMap[nodeId]
			nodeUrl := GenerateUrl(nodeAddr, "receive_broadcast_dead_node")

			res, err := http.Post(nodeUrl, "application/json", bytes.NewBuffer(jsonMsg))

			if err != nil || res.StatusCode/500 >= 1 {
				// detected target node is dead, new broadcast
				if !queueTracker[nodeId] {
					fmt.Println("[ Node", localNode.ID, "] Detected dead node during broadcast:", nodeId)
					localBCQueue = append(localBCQueue, nodeId)
					queueTracker[nodeId] = true
				}
				broadcastTracker[currentDeadNode] = true
				continue
			} else if res.StatusCode != http.StatusOK {
				fmt.Println("[ Node", localNode.ID, "] Non-200 response received during broadcast to node at address", nodeAddr)
				fmt.Println("[ Node", localNode.ID, "] Aborting...")
				return
			}

			broadcastTracker[currentDeadNode] = true
			defer res.Body.Close()
		}
	}

}

func HandleReceiveBroadcast(msg models.BroadcastMessage) {
	localNode := GetLocalNode()
	deadNodeId := msg.DeadNode

	// check if dead node is its successor
	if deadNodeId == localNode.Successor {
		HandleInvoluntaryLeaveSequence(deadNodeId)
	} else {
		// do nothing
		return
	}
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

func HandleCycleCheckStart() {
	localNode := GetLocalNode()
	msg := *models.NewCycleCheckMessage()
	msg.Initiator = localNode.ID
	msg.Nodes = append(msg.Nodes, localNode.ID)

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("[ Node", localNode.ID, "] JSON error during cycle check initiation.")
		return
	}

	successorAddr := config.AllNodeMap[localNode.Successor]
	successorUrl := GenerateUrl(successorAddr, "cycle_check")

	res, err := http.Post(successorUrl, "application/json", bytes.NewBuffer(jsonMsg))

	if err != nil || res.StatusCode/500 >= 1 {
		fmt.Println("[ Node", localNode.ID, "] Error sending cycle check initiation request to successor at address", successorAddr)
		if err != nil {
			fmt.Println("[ Node", localNode.ID, "]", err.Error())
		}
		fmt.Println("[ Node", localNode.ID, "] Aborting...")
		// handle involuntary node leave.
		HandleInvoluntaryDeadNode(localNode.Successor)
		return
	} else if res.StatusCode != http.StatusOK {
		fmt.Println("[ Node", localNode.ID, "] Non-200 response received during cycle check initiation from successor at address", successorAddr)
		fmt.Println("[ Node", localNode.ID, "] Aborting...")
		return
	}
	defer res.Body.Close()
}

func HandleCycleCheck(msg models.CycleCheckMessage) {
	localNode := GetLocalNode()
	if slices.Contains(msg.Nodes, localNode.ID) {
		nodeStructString := "<- " + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(msg.Nodes)), " <-> "), "[]") + " ->"
		fmt.Println("[ Node", localNode.ID, "] Cycle check finished. ")
		fmt.Println("[ Node", localNode.ID, "] Initiator:", msg.Initiator)
		fmt.Println("[ Node", localNode.ID, "] Ring structure:", nodeStructString)
		return
	}

	fmt.Println("[ Node", localNode.ID, "] Forwarding cycle check...")
	msg.Nodes = append(msg.Nodes, localNode.ID)
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("[ Node", localNode.ID, "] JSON error during cycle check.")
	}

	successorAddr := config.AllNodeMap[localNode.Successor]
	successorUrl := GenerateUrl(successorAddr, "cycle_check")

	res, err := http.Post(successorUrl, "application/json", bytes.NewBuffer(jsonMsg))

	if err != nil || res.StatusCode/500 >= 1 {
		fmt.Println("[ Node", localNode.ID, "] Error sending cycle check request to successor at address", successorAddr)
		if err != nil {
			fmt.Println("[ Node", localNode.ID, "]", err.Error())
		}
		fmt.Println("[ Node", localNode.ID, "] Aborting...")
		HandleInvoluntaryDeadNode(localNode.Successor)
		return
	} else if res.StatusCode != http.StatusOK {
		fmt.Println("[ Node", localNode.ID, "] Non-200 response received during cycle check from successor at address", successorAddr)
		fmt.Println("[ Node", localNode.ID, "] Aborting...")
		return
	}
	fmt.Printf("Node %d's successorList: %v\n", localNode.ID, localNode.SuccessorList)
	defer res.Body.Close()
}

// Starts the stabilization function for leaving/dying nodes
func StartReconciliation() {
	localNode := GetLocalNode()
	UpdatePopulateFingerTable()
	fmt.Println("[ Node", localNode.ID, "] Starting reconciliation process..")
	sList := localNode.SuccessorList
	msg := models.NewReconcileMessage(localNode.ID, config.AllNodeID, config.AllNodeMap).
		SetSuccessorList(sList).
		Sign(localNode.ID)

	jsonMsg, jsonErr := json.Marshal(msg)
	if jsonErr != nil {
		fmt.Println("[ Node", localNode.ID, "] JSON error during reconciliation start process.")
	}

	predecessorAddr := config.AllNodeMap[localNode.Predecessor]
	predecessorUrl := GenerateUrl(predecessorAddr, "reconcile")

	res, err := http.Post(predecessorUrl, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil || res.StatusCode/500 >= 1 {
		fmt.Println("[ Node", localNode.ID, "] Error sending reconciliation message to predecessor at address", predecessorAddr)
		if err != nil {
			fmt.Println("[ Node", localNode.ID, "]", err.Error())
		}
		fmt.Println("[ Node", localNode.ID, "] Aborting...")
		return
	} else if res.StatusCode != http.StatusOK {
		fmt.Println("[ Node", localNode.ID, "] Non-200 response received during reconciliation with predecessor at address", predecessorAddr)
		fmt.Println("[ Node", localNode.ID, "] Aborting...")
		return
	}
	defer res.Body.Close()
}

func HandleReconciliation(msg models.ReconcileMessage) {
	localNode := GetLocalNode()
	if slices.Contains(msg.Signed, localNode.ID) {
		fmt.Println("[ Node", localNode.ID, "] Reconciliation sequence completed.")
		fmt.Println("[ Node", localNode.ID, "] Initiator:", msg.Initiator)
		fmt.Println("[ Node", localNode.ID, "] Signed:", msg.Signed)
		return
	}
	fmt.Println("[ Node", localNode.ID, "] Forwarding reconciliation message...")

	sList := msg.SuccessorList[:len(msg.SuccessorList)-1]
	sList = append([]int{localNode.Successor}, sList...)
	localNode.SuccessorList = sList

	config.AllNodeID = msg.AllNodeId
	config.AllNodeMap = msg.AllNodeMap
	UpdatePopulateFingerTable()

	msg.SuccessorList = sList
	msg.Sign(localNode.ID)
	jsonMsg, jsonErr := json.Marshal(msg)
	if jsonErr != nil {
		fmt.Println("[ Node", localNode.ID, "] JSON error during reconciliation forwarding process.")
	}

	predecessorAddr := config.AllNodeMap[localNode.Predecessor]
	predecessorUrl := GenerateUrl(predecessorAddr, "reconcile")
	res, err := http.Post(predecessorUrl, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil || res.StatusCode/500 >= 1 {
		fmt.Println("[ Node", localNode.ID, "] Error forwarding reconciliation message to predecessor at address", predecessorAddr)
		if err != nil {
			fmt.Println("[ Node", localNode.ID, "]", err.Error())
		}
		fmt.Println("[ Node", localNode.ID, "] Aborting...")
		HandleInvoluntaryDeadNode(localNode.Predecessor)
		return
	} else if res.StatusCode != http.StatusOK {
		fmt.Println("[ Node", localNode.ID, "] Non-200 response received during reconciliation with predecessor at address", predecessorAddr)
		fmt.Println("[ Node", localNode.ID, "] Aborting...")
		return
	}
	defer res.Body.Close()
}

func HandleContactDistantNode(nodeId int) {
	addr := config.AllNodeMap[nodeId]
	url := GenerateUrl(addr, "cycle_check")

	res, err := http.Get(url)
	if err != nil {
		fmt.Println("Dead node detected by /contact_distant.")
		HandleInvoluntaryDeadNode(nodeId)
		return
	} else if res.StatusCode != http.StatusOK {
		fmt.Println("Non-200 response during /contact_distant.")
		return
	}
	defer res.Body.Close()

}

func HandleRetrieveDataWithDuplication(key int, c *gin.Context) {
	fmt.Printf("Attempting retrieval for primary key: %d\n", key)

	// Attempt retrieval from the primary node
	primarySuccess := attemptRetrieveData(key, c)

	if primarySuccess {
		fmt.Println("Primary retrieval succeeded.")
		return
	}

	// If primary retrieval fails, attempt retrieval from the replica key
	fmt.Println("Primary node retrieval failed. Attempting replica retrieval...")

	// Retrieve the replica key (this should be the pre-computed replica key from the storage phase)
	replicaKey := retrieveReplicaKeyFromStoragePhase(key)

	fmt.Printf("Attempting retrieval for replica key: %d\n", replicaKey)
	replicaSuccess := attemptRetrieveData(replicaKey, c)

	if replicaSuccess {
		fmt.Println("Replica retrieval succeeded.")
		return
	}

	// If all attempts fail, respond with an error
	fmt.Println("All retrieval attempts failed.")
	c.JSON(http.StatusNotFound, gin.H{
		"message": "Data not found on any available node.",
	})
}

var ReplicaMap = make(map[int]int) // maps primaryKey -> replicaKey

func retrieveReplicaKeyFromStoragePhase(primaryKey int) int {
	return ReplicaMap[primaryKey]
}

func attemptRetrieveData(key int, c *gin.Context) bool {
	find_successor_request := models.FindSuccessorRequest{Key: key}
	jsonData, err := json.Marshal(find_successor_request)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return false
	}

	// Attempt multiple nodes if needed
	client := &http.Client{}
	var successorNode int
	var foundSuccessor bool

	// Try to find the successor starting from multiple known nodes (not just closestPrecedingNode)
	for _, nodeID := range config.AllNodeID {
		address := config.AllNodeMap[nodeID]
		if address == "" {
			continue
		}

		url := fmt.Sprintf("http://%s/find_successor", address)
		newReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			continue
		}

		response, err := client.Do(newReq)
		if err != nil {
			fmt.Printf("Error contacting node %s for key %d: %v\n", address, key, err)
			// Node might be down; consider removing it from AllNodeMap/AllNodeID
			removeNodeFromRing(nodeID)
			continue
		}

		// If not 200, node can't help us
		if response.StatusCode != http.StatusOK {
			fmt.Printf("Error: received non-200 response status: %d from %s\n", response.StatusCode, address)
			response.Body.Close()
			removeNodeFromRing(nodeID)
			continue
		}

		var responseBody models.FindSuccessorSuccessResponse
		if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
			fmt.Println("Error decoding response body:", err)
			response.Body.Close()
			continue
		}
		response.Body.Close()

		successorNode = responseBody.Successor
		foundSuccessor = true
		break
	}

	if !foundSuccessor {
		fmt.Println("Could not find any successor due to node failures.")
		return false
	}

	successorAddress := config.AllNodeMap[successorNode]
	if successorAddress == "" {
		fmt.Println("Successor address not found in AllNodeMap.")
		return false
	}

	// Attempt to retrieve data from the successor node
	internalRetrieveURL := fmt.Sprintf("http://%s/internal_retrieve_data/%d", successorAddress, key)
	retrieveResp, err := client.Get(internalRetrieveURL)
	if err != nil {
		fmt.Printf("Error retrieving data from node %s: %v\n", successorAddress, err)
		// Node might be down
		removeNodeFromRing(successorNode)
		return false
	}
	defer retrieveResp.Body.Close()

	if retrieveResp.StatusCode != http.StatusOK {
		fmt.Printf("Node %s could not retrieve data for key %d\n", successorAddress, key)
		return false
	}

	var retrieveDataResponse models.InternalRetrieveDataResponse
	if err := json.NewDecoder(retrieveResp.Body).Decode(&retrieveDataResponse); err != nil {
		fmt.Println("Error decoding internal retrieve response:", err)
		return false
	}

	// Return the retrieved data
	fmt.Printf("Successfully retrieved data from node %s: %s\n", successorAddress, retrieveDataResponse.Data)
	c.JSON(http.StatusOK, retrieveDataResponse)
	return true
}

func removeNodeFromRing(nodeID int) {
	// Safely remove the node from AllNodeMap and AllNodeID
	delete(config.AllNodeMap, nodeID)
	for i, id := range config.AllNodeID {
		if id == nodeID {
			config.AllNodeID = append(config.AllNodeID[:i], config.AllNodeID[i+1:]...)
			break
		}
	}
	fmt.Printf("Node %d removed from ring due to failure.\n", nodeID)
}
