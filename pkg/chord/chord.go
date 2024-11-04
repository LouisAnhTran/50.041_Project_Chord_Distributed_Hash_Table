package chord

import (
	"fmt"
	"log"
	"net/http"
    "encoding/json"
    "io/ioutil"
	"os"
    "time"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/config"
    "github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/models"
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

            // Assuming NodeList is a map where you store this data
            map_node:=map[int]string{response.Data:node_address}
            config.SliceOfNodeMap=append(config.SliceOfNodeMap, map_node)
        }        
    }

    // Append data to slice of node map
    config.AllNodeMap[HashToRange(os.Getenv("NODE_ADDRESS"))]=os.Getenv("NODE_ADDRESS")

    // Assuming NodeList is a map where we store this data
    map_node:=map[int]string{HashToRange(os.Getenv("NODE_ADDRESS")):os.Getenv("NODE_ADDRESS")}
    config.SliceOfNodeMap=append(config.SliceOfNodeMap, map_node)

    // Sorting all the nodes by identifier to determine successor and predecessor
    fmt.Println("config node list before sorting: ",config.SliceOfNodeMap)

    SortByKey(&config.SliceOfNodeMap)

    fmt.Println("config node list after sorting: ",config.SliceOfNodeMap)

    fmt.Println("Total map: ",config.AllNodeMap)

    // Set node property
    local_node.ID=HashToRange(os.Getenv("NODE_ADDRESS"))
    
    // determine successor
    fmt.Println("index of successor for ",local_node.ID,": ",FindSuccessorForNode(&local_node.ID))


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