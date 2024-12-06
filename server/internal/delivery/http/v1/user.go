package v1

import (
	"net/http"

	"github.com/phzeng0726/go-server-template/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initUserRoutes(api *gin.RouterGroup) {
	userApi := api.Group("/users")
	{
		userApi.POST("", h.createUser)

		auth := api.Group("", userIdentityMiddleware(h))
		{
			auth.GET("", h.getUserByEmail)
		}
	}
}

type createUserInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

func (h *Handler) createUser(c *gin.Context) {
	var inp createUserInput
	if err := c.BindJSON(&inp); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.services.Users.CreateUser(c, service.CreateUserInput{
		Name:  inp.Name,
		Email: inp.Email,
	}); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, true)
}

type queryUsersInput struct {
	Email string `form:"email" binding:"required"`
}

func (h *Handler) getUserByEmail(c *gin.Context) {
	var inp queryUsersInput
	if err := c.BindQuery(&inp); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	user, err := h.services.Users.GetUserByEmail(c, service.QueryUsersInput{
		Email: inp.Email,
	})

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, user)
}
