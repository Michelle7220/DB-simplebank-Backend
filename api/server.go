package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/db/util"
	"github.com/techschool/simplebank/token"
)

//server serves http requests for our banking services

type Server struct {
	config     util.Config
	store      db.Store //allow to interacte with database while processing api requests from clients
	tokenMaker token.Maker
	router     *gin.Engine //this router will help us send each api request to correct handler for processing

}

//takes a database as input, and output a server
//NewServer creates a new http server and setup routing
//this function creates a new server instance, and set up all api route for service on that server
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	} //create a new server with input store

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()
	return server, nil

}

func (server *Server) setupRouter() {
	router := gin.Default() //create a router by calling gin.default
	//add routes to router
	//add first api route to create a  account
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccounts)

	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router //set this object to server.router

}

//strat runs the http server on a specific address

func (server *Server) Start(address string) error { //takes an address as input and return an error
	return server.router.Run(address)
}

// in account.go we hace the errorResponse function, here in server.go we define it. gin.H is a mapping function which map it to key- value pair

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
