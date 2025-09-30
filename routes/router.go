package routes

import (
	"time"

	"HelmyTask/handlers" //http handler
	"HelmyTask/middlewares" //logging , recovery , auth
	"HelmyTask/services" //servic interface used by handlers 

	"github.com/gin-gonic/gin"
)


//setup attaches middlewares and routes to a given gin engine 
//it keeps main.go small and makes routes testbale 
func Setup(r *gin.Engine, svc services.UserService, jwtSecret string, jwtExp time.Duration) {
	// global middlewares run for every request
	r.Use(middlewares.RequestLogger(), middlewares.Recovery())

	// Swagger static (very simple)
	//serve swagger spec as a static file so the clients can import it easily 
	r.StaticFile("/swagger.yaml", "./docs/swagger.yaml")

	//versioned Api group fpr clarity and future compatibilty 
	api := r.Group("/api/v1")

	// auth endpoints
	uh := handlers.NewUserHandler(svc, jwtSecret, jwtExp)
	api.POST("/auth/register", uh.Register)
	api.POST("/auth/login", uh.Login)

	// protected endpoints group required valid jwt 
	protected := api.Group("/")
	protected.Use(middlewares.Auth(jwtSecret)) //attach auth middleware to this gruop 
	protected.GET("/me", uh.Me) //current user profile 
}
