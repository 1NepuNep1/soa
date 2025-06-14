package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func ProxyHandler(c *gin.Context) {
	var targetURL string

	switch {
	case strings.HasPrefix(c.Request.URL.Path, "/auth"),
		strings.HasPrefix(c.Request.URL.Path, "/register"),
		strings.HasPrefix(c.Request.URL.Path, "/profile"):
		targetURL = os.Getenv("USER_SERVICE_URL")

	case strings.HasPrefix(c.Request.URL.Path, "/posts"):
		targetURL = os.Getenv("POST_SERVICE_URL_HTTP")

	case strings.HasPrefix(c.Request.URL.Path, "/stats"):
		targetURL = os.Getenv("STATISTIC_SERVICE_URL")

	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": "No service configured for this path"})
		return
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid target URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.FlushInterval = -1

	c.Request.Host = target.Host

	if userID, exists := c.Get("userID"); exists {
		q := c.Request.URL.Query()
		q.Set("clientId", fmt.Sprintf("%v", userID))
		c.Request.URL.RawQuery = q.Encode()
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
