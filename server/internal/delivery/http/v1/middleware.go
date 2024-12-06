package v1

import (
	"errors"
	"net/http"
	"strings"

	"github.com/phzeng0726/go-server-template/pkg/auth"

	"github.com/gin-gonic/gin"
)

// NOTE: This file is for validation purposes. If it passes, it means it is validated.
const (
	authorizationHeader = "Authorization"
	userCtx             = "userId"
)

// Accepted format is Bearer jwt
func (h *Handler) parseAuthHeader(c *gin.Context) (auth.CustomMapClaims, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return auth.CustomMapClaims{}, errors.New("empty auth header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
		return auth.CustomMapClaims{}, errors.New("invalid auth header")
	}

	if len(headerParts[1]) == 0 {
		return auth.CustomMapClaims{}, errors.New("token is empty")
	}

	return h.tokenManager.Parse(headerParts[1])
}

// Authentication middleware, proceeds to the next layer if validation passes
func userIdentityMiddleware(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := h.parseAuthHeader(c)
		if err != nil {
			newResponse(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		c.Set(userCtx, claims.UserId)
		c.Next()
	}
}
