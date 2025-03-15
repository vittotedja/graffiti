package api

import (
	"graffiti-backend/db"

	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests
type Server struct {
	hub *db.Hub
	router *gin.Engine // helps us send each API request to the correct handler for processing
}

// NewServer creates a new HTTP server and sets up all routes
func NewServer (hub *db.Hub) *Server {
	server := &Server{hub: hub}
	router := gin.Default()

	// add routes to router
	router.POST("/api/v1/user", server.createUser)

	server.router = router // save the router object to the server object
	return server
}

// Start runs the HTTP server on a specific address
// and returns an error if the server fails to start
func (server *Server) Start(address string) error { // public
	return server.router.Run(address) // the router field is private, so it cannot be accessed from outside the api package
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}