package handlers

import (
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func ProxyHandler(c *gin.Context) {
	targetURL := os.Getenv("USER_SERVICE_URL")
	if targetURL == "" {
		c.JSON(500, gin.H{"error": "User service URL not set"})
		return
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "Invalid target URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.FlushInterval = -1

	c.Request.Host = target.Host
	proxy.ServeHTTP(c.Writer, c.Request)
}
