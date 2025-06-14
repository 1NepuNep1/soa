package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func ProxyHandler(c *gin.Context) {
	var targetURL string

	if strings.HasPrefix(c.Request.URL.Path, "/auth") ||
		strings.HasPrefix(c.Request.URL.Path, "/register") ||
		strings.HasPrefix(c.Request.URL.Path, "/profile") {
		targetURL = os.Getenv("USER_SERVICE_URL")
	} else if strings.HasPrefix(c.Request.URL.Path, "/posts") {
		targetURL = os.Getenv("POST_SERVICE_URL_HTTP")
	} else {
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
	proxy.ServeHTTP(c.Writer, c.Request)
}
