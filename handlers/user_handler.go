// translates HTTP to use-cases (register/login/me).
//  No DB or JWT parsing here.
//1
package handlers

import (
	"net/http"
	"time"

	"HelmyTask/global" //context lkey (uid) 
	"HelmyTask/models" // DTOs for binding and response 
	"HelmyTask/services" //bussines logic 

	"github.com/gin-gonic/gin"
)

//UserHnadler groups dependencies and exposes methods to handle user routes 
type UserHandler struct {
	svc        services.UserService //usee case layer for user actions 
	jwtSecret  string //secret used to sign tokens (injected from config)
	jwtExpires time.Duration //token life time (injected from congfig )
}

//NewUserHandler is A constructor for *userhandler withh all dependecies provided 
func NewUserHandler(svc services.UserService, jwtSecret string, jwtExp time.Duration) *UserHandler {
	return &UserHandler{svc: svc, jwtSecret: jwtSecret, jwtExpires: jwtExp}
}

// register handle POST / auth / register 
//it validates input , calls the service to create a user and returns created user 

func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	//bind and validate JSON body based on struct tags 
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) //400 on validations erros
		return
	}

	//call service to preform business logic (hashing , unqueness chack , save )
	u, err := h.svc.Register(req)
	if err != nil { //example : email already exist 
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, u) //201 with created user (password omitted "delted" via json:"-" )
}

//login handles Post / auth / login 
//it validates credentiaLS and returns a signed JWt token
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest //struct to hold login input
	
	//bind and validate json
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//call service to verify passwprd and get a token 
	tok, err := h.svc.Login(req, h.jwtSecret, h.jwtExpires)
	if err != nil { // hide details ; ex ; "invalid crednentials"
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.AuthResponse{Token: tok}) //return token as JSON 
}



//Me handle GET / me (protected)
//it reads the authenticated user ID from context (set by auth Middleware ) and returns user data
func (h *UserHandler) Me(c *gin.Context) {
	uidVal, ok := c.Get(global.CtxUserIDKey) //Grab the value stored in context by auth middleware
	if !ok { //if not set , token was middleware not applied 
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
		return
	}
	uid := uidVal.(uint) //cast to uint (we set it as such in middleware)
	u, err := h.svc.GetByID(uid) //load the user entity from db via service 
	if err != nil { //not found or db error
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u) //return user json 
}
