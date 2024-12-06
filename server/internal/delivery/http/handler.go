package http

import (
	"net/http"

	v1 "github.com/phzeng0726/go-server-template/internal/delivery/http/v1"
	"github.com/phzeng0726/go-server-template/internal/service"
	"github.com/phzeng0726/go-server-template/pkg/auth"
	"github.com/phzeng0726/go-server-template/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	services      *service.Services
	tokenManager  auth.TokenManager
	loggerManager logger.LoggerManager
}

func NewHandler(
	services *service.Services,
	tokenManager auth.TokenManager,
	loggerManager logger.LoggerManager,
) *Handler {
	return &Handler{
		services:      services,
		tokenManager:  tokenManager,
		loggerManager: loggerManager,
	}
}

func (h *Handler) Init() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	corsConfig := corsMiddleware()
	if h.loggerManager == nil {
		router.Use(
			gin.Recovery(),
			cors.New(corsConfig),
		)
	} else {
		router.Use(
			gin.Recovery(),
			cors.New(corsConfig),
			loggingMiddleware(h),
		)
	}

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router *gin.Engine) {
	handlerV1 := v1.NewHandler(h.services, h.tokenManager)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}

	router.GET("/ping", h.ping)
}

func (h *Handler) ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "pong")
}
