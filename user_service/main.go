package main

import (
	"log"
	"userservice/database"
	"userservice/handlers"
	"userservice/models"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()
	database.DB.AutoMigrate(&models.User{}, &models.UserProfile{}, &models.Session{})

	router := gin.Default()

	router.POST("/register", handlers.RegisterHandler)
	router.POST("/auth", handlers.AuthHandler)

	protected := router.Group("/")
	protected.Use(handlers.AuthMiddleware())
	{
		protected.GET("/profile", handlers.ProfileHandler)
		protected.GET("/profile/:id", handlers.GetUserProfileByID)
		protected.PUT("/profile", handlers.UpdateProfileHandler)
	}

	log.Println("Server started on :8000")
	router.Run(":8000")
}
