package tests

import (
	"fmt"
	"os"
	"testing"
	"time"
	"userservice/database"
	"userservice/handlers"
	"userservice/models"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestCheckUserCredentials(t *testing.T) {
	login := fmt.Sprintf("testuser%d", time.Now().UnixNano())
	email := fmt.Sprintf("user%d@example.com", time.Now().UnixNano())
	password := "mypassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := models.User{
		Login:        login,
		PasswordHash: string(hash),
		Email:        email,
		Status:       "active",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	foundUser, err := handlers.CheckUserCredentials(login, password)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)

	_, err = handlers.CheckUserCredentials(login, "wrongpassword")
	assert.Error(t, err)
}

func TestGenerateJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key")
	userID := uint(1)

	tokenString, err := handlers.GenerateJWT(userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte("test_secret_key"), nil
	})

	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, float64(userID), claims["userId"])
}
