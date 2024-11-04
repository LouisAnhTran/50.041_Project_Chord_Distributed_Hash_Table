package chord

import (
	"fmt"
    "bytes"
	"log"
	"net/http"
    "encoding/json"
    "io/ioutil"
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

func InitChordRingStructure(){
    // go routine sleep for 3 seconds, giving sufficient time for all nodes to set up routes
    time.Sleep(3 * time.Second) // Pauses for 2 seconds

    for i := range config.NodeAddresses {
        node_address:=config.NodeAddresses[i]

        if node_address != os.Getenv("NODE_ADDRESS") {
            url:=fmt.Sprintf("http://%s/node_identifier",node_address)

            resp,err :=  http.Get(url)
    
            if err != nil {
                log.Fatalf("Failed to send request to %s: %v",node_address,err)
                continue
            }
        
            // Read and print the response
            body, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                log.Fatalf("Failed to read response: %v", err)
                continue
            }
    
            fmt.Printf("Response from %s: %s \n",node_address,string(body))

            // Unmarshal JSON into the Response struct
            var response models.ResponseNodeIdentifier

            if err := json.Unmarshal(body, &response); err != nil {
                log.Fatalf("Failed to parse JSON: %v", err)
                continue
            }

            // Now we can access the Data field directly
            fmt.Printf("Response from %s: %d\n", node_address, response.Data)

            // Append data to slice of node map
            config.AllNodeMap[response.Data]=node_address
        }        
    }

    // Append data to slice of node map
    config.AllNodeMap[HashToRange(os.Getenv("NODE_ADDRESS"))]=os.Getenv("NODE_ADDRESS")

    // Populate all node ID
    populate_all_node_id_and_sort()

    // Set node property
    local_node.ID=HashToRange(os.Getenv("NODE_ADDRESS"))

    fmt.Println("print all sorted node ids: ",config.AllNodeID)


    // determine successor
    var successor=FindSuccessorForNode()
    local_node.Successor=successor
    fmt.Println("my id ",local_node.ID, " my successor is: ",local_node.Successor)

    // determine predecessor 
    var predecessor=FindPredecessorForNode()
    local_node.Predecessor=predecessor
    fmt.Println("my id ",local_node.ID, " my predecessor is: ",local_node.Predecessor)

    // Populate Finger Table
    PopulateFingerTable()
    fmt.Println("my all node id: ",config.AllNodeID," || my id is: ",local_node.ID," || my finger table: ",local_node.FingerTable," || my map: ",config.AllNodeMap," || my successor: ",local_node.Successor," || predecessor: ",local_node.Predecessor)
} 

func HandleFindSuccessor(req models.Request,c *gin.Context) int {
    if req.ID > local_node.ID && req.ID <= local_node.Successor {
        fmt.Println("I found successor: ",local_node.Successor)
        c.JSON(http.StatusOK,gin.H{
            "successor":local_node.Successor})
        return 1
    }

    closest_preceding_node:=find_closest_preceding_node(req.ID)

    fmt.Println("closest preceding node: ",closest_preceding_node)

    address_closest_proceed_node:=config.AllNodeMap[closest_preceding_node]

    // prepare request
    jsonData, err := json.Marshal(req)
    if err != nil {
        fmt.Println("Error marshaling JSON:", err)
        return -1
    }

    url:=fmt.Sprintf("http://%s/find_successor",address_closest_proceed_node)

     // Create a new request
    new_req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("Error creating request:", err)
        return -1
    }

    // Send the request using the http.Client
    client := &http.Client{}
    response, err := client.Do(new_req)
    if err != nil {
        fmt.Println("Error making POST request:", err)
        return -1
    }
    defer response.Body.Close()

    // Check the response status code
    if response.StatusCode != http.StatusOK {
        fmt.Println("Error: received non-200 response status:", response.Status)
        return -1
    }

    // Read and print the response body
    var responseBody map[string]interface{}
    if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
        fmt.Println("Error decoding response body:", err)
        return -1
    }

    fmt.Println("Response from server:", responseBody)

    c.JSON(http.StatusOK,gin.H{
        "successor":responseBody["successor"]})

    return 0
}

func find_closest_preceding_node(node_id int) int{
    for i:=len(local_node.FingerTable)-1;i>=0;i-- {
        for k,v := range local_node.FingerTable[i] {
            fmt.Println("k is: ",k," v: ",v)
            if v < node_id {
                return v
            }
        }
    }
    return local_node.ID
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