package tests

import (
	"apigateway/handlers"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closeNotify chan bool
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closeNotify
}

func newCloseNotifyingRecorder() *closeNotifyingRecorder {
	return &closeNotifyingRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closeNotify:      make(chan bool, 1),
	}
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Any("/*proxyPath", handlers.ProxyHandler)
	return router
}

func TestProxyHandler_RoutesCorrectly(t *testing.T) {
	router := setupRouter()

	os.Setenv("USER_SERVICE_URL", "http://mock-user-service:8000")

	req, _ := http.NewRequest("GET", "/profile", nil)
	w := newCloseNotifyingRecorder()

	router.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestProxyHandler_InvalidRoute(t *testing.T) {
	router := setupRouter()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	w := newCloseNotifyingRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadGateway, w.Code)
}
