package main

import (
	"log"
	"time"

	"HelmyTask/config"
	"HelmyTask/repositories"
	"HelmyTask/routes"
	"HelmyTask/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1) Load config from file and||or env
	cfg := config.Load()  // Returns *config.Config with merged settings.
	log.Printf("[boot] %s starting in %s on :%s", cfg.AppName, cfg.Env, cfg.HTTPPort)

	// 2) Initialize infrastructure (DB and Redis).
	db := config.InitDB(cfg)  // Open DB based on cfg.DBDriver and run migrations.
	_ = config.InitRedis(cfg) // Create Redis client (available for future use).

	// 3) Construct repositories and services (dependency injection).
	userRepo := repositories.NewUserRepository(db) // Repo uses *gorm.DB to talk to chosen DB.
	userSvc := services.NewUserService(userRepo) // Service wraps business rules and JWT issuance.


	// 4) Create Gin engine and wire routes
	r := gin.New() // Create a new bare Gin engine (no default middleware).
	jwtExp, _ := time.ParseDuration(cfg.JWTExpires) // Convert "72h" to time.Duration (ignore parse err due to defaults).
	routes.Setup(r, userSvc, cfg.JWTSecret, jwtExp) // Attach middlewares and endpoints.

		// 5) Start HTTP server on configured port; fatal if it fails to bind.
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatal(err) // Stop the process if server fails to start.
	}
}
