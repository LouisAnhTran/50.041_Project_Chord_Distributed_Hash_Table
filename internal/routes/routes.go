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
    router.GET("/node_identifier",getNodeAddressAndIdentifier)
    router.GET("/health_check",health_check)
    router.POST("/find_successor",find_successor)
    router.POST("/store_data",store_data)
    router.POST("/internal_store_data",internal_store_data)
    router.GET("/retrieve_data/:id",retrieve_data)
    router.GET("/internal_retrieve_data/:id",internal_retrieve_data)
}

func retrieve_data(c *gin.Context){
    // to do
    key_str := c.Param("id")

    key,err:=strconv.Atoi(key_str)
    if err != nil {
        c.JSON(http.StatusBadRequest,models.RetrieveDataResponse{Message: "Invalid ID format"})
        return
    }

    fmt.Println("key of data to be retrieved: ",key)

    chord.HandleRetrieveData(key,c)
}

func internal_retrieve_data(c *gin.Context){
    // to do
    key_str := c.Param("id")

    key,err:=strconv.Atoi(key_str)
    if err != nil {
        c.JSON(http.StatusBadRequest,models.InternalRetrieveDataResponse{Message: "Invalid ID format"})
        return
    }

    fmt.Println("key of data to be retrieved: ",key)

    chord.HandleInternalRetrieveData(key,c)
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

    fmt.Println("data to be stored: ",req.Data)

    chord.HandleStoreData(req,c)

}

func internal_store_data(c *gin.Context){
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

    fmt.Println("data to be stored: ",req.Data)
    fmt.Println("key of data to be stored: ",req.Key)

    chord.HandleInternalStoreData(req,c)
}

func find_successor(c *gin.Context){
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
    fmt.Println("key from request ",req.Key)

    chord.HandleFindSuccessor(req,c)
}

func health_check(c *gin.Context){
    c.JSON(http.StatusOK, gin.H{"message":"good health"})
}

func getNodeAddressAndIdentifier(c *gin.Context) {
    node_identifier:=chord.HashToRange(os.Getenv("NODE_ADDRESS"))
    c.JSON(http.StatusOK, gin.H{"data":node_identifier})
}



