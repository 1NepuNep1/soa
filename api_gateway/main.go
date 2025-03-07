package main

import (
	"apigateway/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.Any("/*proxyPath", handlers.ProxyHandler)

	log.Println("API Gateway running on :8080")
	router.Run(":8080")
}
