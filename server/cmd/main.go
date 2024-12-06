package main

import (
	"log"

	"github.com/phzeng0726/go-server-template/internal/config"
	"github.com/phzeng0726/go-server-template/internal/database"
	"github.com/phzeng0726/go-server-template/internal/repository"
	"github.com/phzeng0726/go-server-template/internal/service"
	"github.com/phzeng0726/go-server-template/pkg/logger"

	delivery "github.com/phzeng0726/go-server-template/internal/delivery/http"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

var (
	repos    *repository.Repositories
	services *service.Services
	handlers *delivery.Handler
	router   *gin.Engine
)

func initLogger(configEnv config.AppConfig) logger.LoggerManager {
	loggerManager, err := logger.NewManager("your_server_name", configEnv.Env, configEnv.DataPath.LogFolder)
	if err != nil {
		log.Fatalln(err)
	}

	if err := loggerManager.InitLogger(); err != nil {
		log.Fatalln(err)
	}

	return loggerManager
}

func initServiceDeps(repos *repository.Repositories, configEnv config.AppConfig) service.Deps {
	// // Necessary if using token manager
	// tokenManager, err := auth.NewManager(nil, "./public.pem")
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	deps := service.Deps{
		Repos: repos,
		// TokenManager: tokenManager,
	}

	log.Printf("Current environment: %v", configEnv.Env)

	return deps
}

func init() {
	config.InitConfig()
	if config.Env.Env == "" {
		log.Fatalln("Failed to load env")
	}

	log.Printf("Disable log: %v", config.Env.DisableLog)
}

func main() {
	// Database
	conn := database.Connect()
	database.SyncDatabase(conn)

	// Others
	var loggerManager logger.LoggerManager
	if config.Env.DisableLog {
		loggerManager = nil
	} else {
		loggerManager = initLogger(*config.Env)
	}

	// 3-Layers
	repos = repository.NewRepositories(conn)
	deps := initServiceDeps(repos, *config.Env)
	services = service.NewServices(deps)
	handlers = delivery.NewHandler(services, deps.TokenManager, loggerManager)
	router = handlers.Init()

	router.Run(config.Env.Host + ":" + config.Env.Port)

}
