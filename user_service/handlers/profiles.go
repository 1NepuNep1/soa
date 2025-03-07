package handlers

import (
	"net/http"
	"time"
	"userservice/database"
	"userservice/models"

	"github.com/gin-gonic/gin"
)

func ProfileHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := database.DB.Preload("Profile").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateProfileHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	var input struct {
		FirstName   string     `json:"firstName"`
		LastName    string     `json:"lastName"`
		BirthDate   *time.Time `json:"birthDate"`
		PhoneNumber string     `json:"phoneNumber"`
		Bio         string     `json:"bio"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var profile models.UserProfile
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	profile.FirstName = input.FirstName
	profile.LastName = input.LastName
	profile.BirthDate = input.BirthDate
	profile.PhoneNumber = input.PhoneNumber
	profile.Bio = input.Bio

	if err := database.DB.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func GetUserProfileByID(c *gin.Context) {
	id := c.Param("id")

	var profile models.UserProfile
	if err := database.DB.Where("user_id = ?", id).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}
