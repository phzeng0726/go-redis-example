package http

import (
	"net"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/phzeng0726/go-server-template/internal/config"
)

func corsMiddleware() cors.Config {
	allowedOrigins := strings.Split(config.Env.AccessAllowOrigin, ",")
	for i, origin := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(origin)
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = allowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Authorization", "Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Cache-Control", "X-Requested-With"}
	corsConfig.AllowCredentials = true

	return corsConfig
}

func loggingMiddleware(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.loggerManager == nil {
			c.Next()
			return
		}

		startTime := time.Now()

		// Process the request
		c.Next()

		// Log the request after processing is completed
		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()

		var clientIP string = ""
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err == nil {
			clientIP = host
		}

		fields := map[string]interface{}{
			"method":      reqMethod,
			"uri":         reqUri,
			"status_code": statusCode,
			"ip":          clientIP,
		}

		go func() {
			if err := c.Errors.Last(); err != nil {
				h.loggerManager.ErrorWithElapsedTime(c, "API request failed", startTime, err.Err, fields)
			} else {
				h.loggerManager.InfoWithElapsedTime(c, "API request completed", startTime, fields)
			}
		}()
	}
}
