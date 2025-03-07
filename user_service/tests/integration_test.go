package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"userservice/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/register", handlers.RegisterHandler)
	router.POST("/auth", handlers.AuthHandler)

	protected := router.Group("/")
	protected.Use(handlers.AuthMiddleware())
	{
		protected.GET("/profile", handlers.ProfileHandler)
		protected.PUT("/profile", handlers.UpdateProfileHandler)
	}

	return router
}

func TestAuthHandler_Success(t *testing.T) {
	router := setupRouter()

	registerBody := `{"login":"testuser123","password":"password123","email":"user@example.com"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	authBody := `{"login":"testuser123","password":"password123"}`
	req, _ = http.NewRequest("POST", "/auth", bytes.NewBufferString(authBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_WrongPassword(t *testing.T) {
	router := setupRouter()

	authBody := `{"login":"testuser123","password":"wrongpass"}`
	req, _ := http.NewRequest("POST", "/auth", bytes.NewBufferString(authBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProfileHandler_Unauthorized(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegisterHandler_BadRequest(t *testing.T) {
	router := setupRouter()

	payload := `{"login":"","password":"123","email":"bademail"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProfileHandler_Unauthorized(t *testing.T) {
	router := setupRouter()

	updateBody, _ := json.Marshal(gin.H{"firstName": "Test", "lastName": "User"})
	req, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_InvalidJSON(t *testing.T) {
	router := setupRouter()

	invalidJSON := `{ "login": "testuser123", "password": }` // Некорректный JSON
	req, _ := http.NewRequest("POST", "/auth", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_ExistingUser(t *testing.T) {
	router := setupRouter()

	registerBody := `{"login":"duplicateuser","password":"password123","email":"duplicate@example.com"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	req, _ = http.NewRequest("POST", "/register", bytes.NewBufferString(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestProfileHandler_NotFound(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/profile/99999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
