package routes

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/models"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/pkg/chord"
	"github.com/gin-gonic/gin"
)

// SetupRoutes initializes the API routes
func SetupRoutes(router *gin.Engine) {
	router.GET("/node_identifier", getNodeAddressAndIdentifier)
	router.GET("/health_check", health_check)
	router.POST("/find_successor", find_successor)
	router.POST("/store_data", store_data)
	router.POST("/internal_store_data", internal_store_data)
	router.GET("/contact_distant", contactDistant)
	router.GET("/retrieve_data/:id", retrieve_data)
	router.GET("/internal_retrieve_data/:id", internal_retrieve_data)
	router.GET("/leave", leave)
	router.GET("/kill", kill)
	router.GET("/cycle_check", cycleCheckStart)
	router.POST("/cycle_check", cycleCheck)
	router.POST("/notify_leave", relink)
	router.POST("/leave_data", updateDataOnLeave)
	router.POST("/reconcile", reconcile)
	router.POST("/notify", notify)
	router.POST("/start_stablization", start_stablization)
	router.POST("/update_metadata", update_metadata)
	router.POST("/notify_new_successor", handleNewSuccessorNotification)
	router.POST("/receive_broadcast_dead_node", handleReceiveBroadcast)
	router.GET("/retrieve_data_with_duplication/:key", retrieve_data_with_duplication)
}

func handleReceiveBroadcast(c *gin.Context) {
	var msg models.BroadcastMessage
	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, models.NewHTTPErrorMessage("Invalid JSON body", err.Error()))
		return
	}

	c.JSON(http.StatusOK, "")
	chord.HandleReceiveBroadcast(msg)
}

func handleNewSuccessorNotification(c *gin.Context) {
	var msg models.InvoluntaryLeaveMessage
	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, models.NewHTTPErrorMessage("Invalid JSON body", err.Error()))
		return
	}

	c.JSON(http.StatusOK, "")
	chord.HandleNodeInvoluntaryLeave(msg)
}

func reconcile(c *gin.Context) {
	var msg models.ReconcileMessage
	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, models.NewHTTPErrorMessage("Invalid JSON body", err.Error()))
		return
	}

	chord.HandleReconciliation(msg)
}

func kill(c *gin.Context) {
	defer os.Exit(0)

	// See ya.
	c.JSON(http.StatusOK, models.NewHTTPErrorMessage("See ya.", ""))

	fmt.Println("[ Node", chord.GetLocalNode().ID, "] I died")
}

func leave(c *gin.Context) {
	fmt.Println("[ Node", chord.GetLocalNode().ID, "] Requested to leave ring...")
	chord.HandleLeaveSequence()
}

func updateDataOnLeave(c *gin.Context) {
	var msg models.DataUpdateMessage
	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, models.NewHTTPErrorMessage("Invalid JSON body", err.Error()))
		return
	}

	c.JSON(http.StatusOK, "")
	fmt.Println("[ Node", chord.GetLocalNode().ID, "] Received data of Node", msg.DepartingNodeID, "from Node", msg.SenderID)
	fmt.Println("[ Node", chord.GetLocalNode().ID, "] Data:", msg.Data)
}

func relink(c *gin.Context) {
	var msg models.LeaveRingMessage

	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, models.NewHTTPErrorMessage("Invalid JSON body", err.Error()))
		return
	}

	fmt.Println("[ Node", chord.GetLocalNode().ID, "] Handling leave sequence for departing node", msg.DepartingNodeID)
	chord.HandleNodeLeave(msg)
}

func cycleCheck(c *gin.Context) {
	var cycleCheckMessage models.CycleCheckMessage

	if err := c.BindJSON(&cycleCheckMessage); err != nil {
		c.JSON(http.StatusBadRequest, models.NewHTTPErrorMessage("Invalid JSON body", err.Error()))
		return
	}
	chord.HandleCycleCheck(cycleCheckMessage)
}

func cycleCheckStart(c *gin.Context) {
	fmt.Println("[ Node", chord.GetLocalNode().ID, "] Starting cycle check...")
	c.JSON(http.StatusOK, gin.H{"Message": "OK"})
	chord.HandleCycleCheckStart()
}

func update_metadata(c *gin.Context) {
	// to do
	var req models.UpdateMetadataUponNewNodeJoinRequest

	// Bind JSON data to the request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.UpdateMetadataUponNewNodeJoinResponse{
			Error: err.Error(),
		})
		return
	}

	fmt.Println("Update metadata - joining node's succesor sent joining node's key: ", req.Key)
	fmt.Println("Update metadata - joining node's succesor sent joining node's address: ", req.NodeAddress)

	chord.HandleUpdateMetaData(req, c)
}

func start_stablization(c *gin.Context) {
	// to do
	var req models.StablizationSuccessorRequest

	// Bind JSON data to the request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.StablizationSuccessorResponse{
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	fmt.Println("message from succesor: ", req.Message)
	fmt.Println("succesor sent joining node's key: ", req.Key)
	fmt.Println("succesor sent joining node's address: ", req.NodeAddress)

	chord.HandleStartStablization(req, c)
}

func notify(c *gin.Context) {
	// to do
	var req models.NotifyRequest

	// Bind JSON data to the request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NotifyResponse{
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	fmt.Println("successor received: joining node key: ", req.Key)
	fmt.Println("successor received: joining node address: ", req.NodeAddress)

	chord.HandleSuccessorNotification(req, c)
}

func retrieve_data(c *gin.Context) {
	// to do
	key_str := c.Param("id")

	key, err := strconv.Atoi(key_str)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.RetrieveDataResponse{Message: "Invalid ID format"})
		return
	}

	fmt.Println("key of data to be retrieved: ", key)

	chord.HandleRetrieveData(key, c)
}

func internal_retrieve_data(c *gin.Context) {
	// to do
	key_str := c.Param("id")

	key, err := strconv.Atoi(key_str)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.InternalRetrieveDataResponse{Message: "Invalid ID format"})
		return
	}

	fmt.Println("key of data to be retrieved: ", key)

	chord.HandleInternalRetrieveData(key, c)
}

func store_data(c *gin.Context) {
	// to do
	var req models.StoreDataRequest

	// Bind JSON data to the request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.StoreDataResponsee{
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	fmt.Println("data to be stored: ", req.Data)

	chord.HandleStoreData(req, c)

}

func internal_store_data(c *gin.Context) {
	// to do
	var req models.InternalStoreDataRequest

	// Bind JSON data to the request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.StoreDataResponsee{
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	fmt.Println("data to be stored: ", req.Data)
	fmt.Println("key of data to be stored: ", req.Key)

	chord.HandleInternalStoreData(req, c)
}

func find_successor(c *gin.Context) {
	var req models.FindSuccessorRequest

	// Bind JSON data to the request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.FindSuccessorErrorResponse{
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	// handle request
	fmt.Println("key from request ", req.Key)

	chord.HandleFindSuccessor(req, c)
}

func health_check(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "good health"})
}

func contactDistant(c *gin.Context) {
	nodeId := c.Query("id")
	if len(nodeId) == 0 {
		c.JSON(http.StatusBadRequest, models.NewHTTPErrorMessage("Missing node ID to be messaged.", ""))
	}

	nodeIdInt, err := strconv.Atoi(nodeId)
	if err != nil {
		fmt.Println("Error converting id from string to integer.")
		c.JSON(http.StatusInternalServerError, models.NewHTTPErrorMessage("Error converting id from string to integer.", err.Error()))
	}

	chord.HandleContactDistantNode(nodeIdInt)
}

func getNodeAddressAndIdentifier(c *gin.Context) {
	node_identifier := chord.HashToRange(os.Getenv("NODE_ADDRESS"))
	c.JSON(http.StatusOK, gin.H{"data": node_identifier})
}

func retrieve_data_with_duplication(c *gin.Context) {
	// Parse key from the URL parameter
	keyStr := c.Param("key")

	key, err := strconv.Atoi(keyStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.RetrieveDataResponse{Message: "Invalid key format"})
		return
	}

	fmt.Println("Attempting to retrieve data with duplication for key: ", key)

	// Call the enhanced retrieval handler
	chord.HandleRetrieveDataWithDuplication(key, c)
}
