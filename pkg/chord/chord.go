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
	"sort"
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
	fmt.Println("my all node id: ", config.AllNodeID, " || my id is: ", local_node.ID, " || my finger table: ", local_node.FingerTable, " || my map: ", config.AllNodeMap, " || my successor: ", local_node.Successor, " || predecessor: ", local_node.Predecessor)
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
