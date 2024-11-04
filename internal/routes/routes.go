package routes

import (
	"net/http"
	"os"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/models"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/pkg/chord"
	"github.com/gin-gonic/gin"
)

var users = []models.User{
    {ID: "1", Name: "John Doe", Age: 30},
    {ID: "2", Name: "Jane Doe", Age: 25},
}

// SetupRoutes initializes the API routes
func SetupRoutes(router *gin.Engine) {
    router.GET("/users", getUsers)
    router.GET("/users/:id", getUser)
    router.POST("/users", createUser)
    router.GET("/node_identifier",getNodeAddressAndIdentifier)
}

func getNodeAddressAndIdentifier(c *gin.Context) {
    node_identifier:=chord.HashToRange(os.Getenv("NODE_ADDRESS"))
    c.JSON(http.StatusOK, gin.H{"data":node_identifier})
}


func getUsers(c *gin.Context) {
    c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context) {
    id := c.Param("id")
    for _, user := range users {
        if user.ID == id {
            c.JSON(http.StatusOK, user)
            return
        }
    }
    c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
}

func createUser(c *gin.Context) {
    var newUser models.User
    if err := c.ShouldBindJSON(&newUser); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    users = append(users, newUser)
    c.JSON(http.StatusCreated, newUser)
}
