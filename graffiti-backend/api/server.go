package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
)

// Server serves HTTP requests
type Server struct {
	hub    *db.Hub
	router *gin.Engine // helps us send each API request to the correct handler for processing
}

// NewServer creates a new HTTP server and sets up all routes
func NewServer(hub *db.Hub) *Server {
	server := &Server{hub: hub}
	router := gin.Default()
	router.Use(logger.Middleware())

	// add routes to router

	// Set up routes for the user API
	router.POST("/api/v1/users", server.createUser) // working
	router.GET("/api/v1/users/:id", server.getUser)
	router.GET("/api/v1/users", server.listUsers)      // working
	router.PUT("/api/v1/users/:id", server.updateUser) // working
	router.DELETE("/api/v1/users/:id", server.deleteUser)
	router.PUT("/api/v1/users/:id/profile", server.updateProfile)
	router.PUT("/api/v1/users/:id/onboarding", server.finishOnboarding) // working

	// Set up routes for the wall API
	router.POST("/api/v1/walls", server.createWall)
	router.GET("/api/v1/walls/:id", server.getWall)                 // working
	router.GET("/api/v1/walls", server.listWalls)                   // working
	router.GET("/api/v1/users/:id/walls", server.listWallsByUser)   // working
	router.PUT("/api/v1/walls/:id", server.updateWall)              // working
	router.PUT("/api/v1/walls/:id/publicize", server.publicizeWall) // working
	router.PUT("/api/v1/walls/:id/privatize", server.privatizeWall) // working
	router.PUT("/api/v1/walls/:id/archive", server.archiveWall)
	router.PUT("/api/v1/walls/:id/unarchive", server.unarchiveWall)
	router.DELETE("/api/v1/walls/:id", server.deleteWall) // working

	// Set up routes for the post API
	router.POST("/api/v1/posts", server.createPost)                                     // working
	router.GET("/api/v1/posts/:id", server.getPost)                                     // working
	router.GET("/api/v1/posts", server.listPosts)                                       // working
	router.GET("/api/v1/walls/:id/posts", server.listPostsByWall)                       // working
	router.GET("/api/v1/posts/highlighted", server.getHighlightedPosts)                 // working
	router.GET("/api/v1/walls/:id/posts/highlighted", server.getHighlightedPostsByWall) // working
	router.PUT("/api/v1/posts/:id", server.updatePost)                                  // working
	router.PUT("/api/v1/posts/:id/highlight", server.highlightPost)                     // working
	router.PUT("/api/v1/posts/:id/unhighlight", server.unhighlightPost)                 // working
	router.DELETE("/api/v1/posts/:id", server.deletePost)                               // working

	// Updated Friendship API routes
	// Friend Requests
	router.POST("/api/v1/friend-requests", server.createFriendRequest)       // working
	router.PUT("/api/v1/friend-requests/accept", server.acceptFriendRequest) // working

	// User Blocking
	router.PUT("/api/v1/users/block", server.blockUser)
	router.PUT("/api/v1/users/unblock", server.unblockUser)

	// Friends Retrieval
	router.GET("/api/v1/users/:id/accepted-friends", server.getFriends)                      // user_id; working
	router.GET("/api/v1/users/:id/friend-requests/pending", server.getPendingFriendRequests) // user_id; working
	router.GET("/api/v1/users/:id/friend-requests/sent", server.getSentFriendRequests)       // user_id; working

	// Existing friendship-related routes
	router.GET("/api/v1/users/:id/friendships", server.listFriendshipsByUserId)                            // working
	router.GET("/api/v1/users/:id/accepted-friends/count", server.getNumberOfFriends)                      // working
	router.GET("/api/v1/users/:id/friend-requests/pending/count", server.getNumberOfPendingFriendRequests) // working
	// router for listFriendshipByUserPairs
	// router.GET("/api/v1/friendships", server.listFriendshipByUserPairs)

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
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
