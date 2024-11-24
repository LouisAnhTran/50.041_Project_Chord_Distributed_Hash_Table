package main

import (
	// "fmt"
	// "os"
	"os"
	"time"

	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/config"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/internal/routes"
	"github.com/LouisAnhTran/50.041_Project_Chord_Distributed_Hash_Table/pkg/chord"
	"github.com/fatih/color"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
    color.Green("This is green text!")
    logrus.Info("This is a log message using Logrus")
    chord.Test_pack()

    // Compute my own hash digest and send to all other nodes in the system
    // the init chord ring structure function is only applicable to all nodes in initial network 
    my_address:=os.Getenv("NODE_ADDRESS")

    if my_address!=config.NewJoinNodeAddress {
        // if i am not a newly joined node, then i will join ring construction process
        go chord.InitChordRingStructure()
    } else {
        // I am not newly joint node, I will need to do a few steps to join the network
        go chord.NewNodeJoinNetwork()
    }


    // Set up routes
    router := gin.Default()


    // CORS middleware configuration to allow everything
    router.Use(cors.New(cors.Config{
        AllowAllOrigins: true, // Allow all origins
        AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
        AllowHeaders:    []string{"*"}, // Allow all headers
        ExposeHeaders:   []string{"*"}, // Expose all headers
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
    routes.SetupRoutes(router)

    // Run the server
    router.Run(":8080")

}
