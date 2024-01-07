package main

import (
	"FraudDetection/api"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Hello World")
	router := gin.Default()

	// Initialize routes
	api.InitRoutes(router)

	// Start server
	router.Run(":8080")
}
