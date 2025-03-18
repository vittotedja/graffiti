package api

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
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
	
	// Set up routes for the user API
	router.POST("/api/v1/users", server.createUser)
	router.GET("api/v1/users/:id", server.getUser)
	router.GET("/api/v1/users", server.listUsers)
	router.PUT("/users/:id", server.updateUser)
	router.DELETE("/users/:id", server.deleteUser)
	router.PUT("/users/:id/profile", server.updateProfile)
	router.PUT("/users/:id/onboarding", server.finishOnboarding)

	// Set up routes for the wall API
	router.POST("/api/v1/walls", server.createWall)
	router.GET("/api/v1/walls/:id", server.getWall)
	router.GET("/api/v1/walls", server.listWalls)
	router.GET("/api/v1/users/:id/walls", server.listWallsByUser)
	router.PUT("/api/v1/walls/:id", server.updateWall)
	router.PUT("/api/v1/walls/:id/publicize", server.publicizeWall)
	router.PUT("/api/v1/walls/:id/privatize", server.privatizeWall)
	router.PUT("/api/v1/walls/:id/archive", server.archiveWall)
	router.PUT("/api/v1/walls/:id/unarchive", server.unarchiveWall)
	router.DELETE("/api/v1/walls/:id", server.deleteWall)

	// Set up routes for the post API
	router.POST("/api/v1/posts", server.createPost)
	router.GET("/api/v1/posts/:id", server.getPost)
	router.GET("/api/v1/posts", server.listPosts)
	router.GET("/api/v1/walls/:id/posts", server.listPostsByWall)
	router.PUT("/api/v1/posts/:id", server.updatePost)
	router.PUT("/api/v1/posts/:id/highlight", server.highlightPost)
	router.DELETE("/api/v1/posts/:id", server.deletePost)

	// Set up routes for the friendship API
	router.POST("/api/v1/friendships", server.createFriendship)
	router.GET("/api/v1/friendships/:id", server.getFriendship)
	router.GET("/api/v1/friendships", server.listFriendships)
	router.PUT("/api/v1/friendships/:id", server.updateFriendship)
	router.DELETE("/api/v1/friendships/:id", server.deleteFriendship)
	router.GET("/api/v1/users/:id/friends/count", server.getNumberOfFriends)
	router.GET("/api/v1/users/:id/friend-requests/count", server.getNumberOfPendingFriendRequests)

	// Set up routes for the like API
	router.POST("/api/v1/likes", server.createLike)
	router.GET("/api/v1/likes/:id", server.getLike)
	router.GET("/api/v1/likes", server.listLikes)
	router.GET("/api/v1/posts/:id/likes", server.listLikesByPost)
	router.GET("/api/v1/users/:id/likes", server.listLikesByUser)
	router.DELETE("/api/v1/likes", server.deleteLike)

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

// parseUUID converts a string to UUID
func parseUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, errors.New("empty UUID string")
	}
	
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format: %w", err)
	}
	
	return id, nil
}