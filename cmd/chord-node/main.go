package main

import (
	// "fmt"
    // "os"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/internal/routes"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/pkg/chord"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
 
)

func main() {
    color.Green("This is green text!")
    logrus.Info("This is a log message using Logrus")
    chord.Test_pack()

    // Compute my own hash digest and send to all other nodes in the system
    go chord.InitChordRingStructure()

    // Set up routes
    router := gin.Default()
    routes.SetupRoutes(router)

    // Run the server
    router.Run(":8080")

}
