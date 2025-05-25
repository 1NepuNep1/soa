package main

import (
	"apigateway/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	router.RedirectTrailingSlash = false
	router.Use(gin.Logger(), gin.Recovery())

	public := router.Group("/")
	{
		public.POST("/auth", handlers.ProxyHandler)
		public.POST("/auth/*proxyPath", handlers.ProxyHandler)

		public.POST("/register", handlers.ProxyHandler)
		public.POST("/register/*proxyPath", handlers.ProxyHandler)
	}

	protected := router.Group("/", handlers.AuthMiddleware())
	{
		protected.Any("/profile", handlers.ProxyHandler)
		protected.Any("/profile/*proxyPath", handlers.ProxyHandler)

		protected.Any("/posts/*proxyPath", handlers.ProxyHandler)
		protected.Any("/posts", handlers.ProxyHandler)
	}

	log.Println("API Gateway running on :8080")
	router.Run(":8080")
}
