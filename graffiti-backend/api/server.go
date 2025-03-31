package api

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/token"
)

// Server serves HTTP requests
type Server struct {
	hub        *db.Hub
	router     *gin.Engine // helps us send each API request to the correct handler for processing
	tokenMaker token.Maker
}

// NewServer creates a new HTTP server and sets up all routes
func NewServer(hub *db.Hub) *Server {
	tokenMaker, err := token.NewJWTMaker("veryverysecretkey")
	if err != nil {
		log.Fatal("cannot create token maker", err)
	}
	server := &Server{hub: hub, tokenMaker: tokenMaker}
	router := gin.Default()

	// Apply CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	//Set up routes for Auth
	router.POST("/api/v1/auth/register", server.Register)
	router.POST("/api/v1/auth/login", server.Login)

	protected := router.Group("/api")
	protected.Use(server.AuthMiddleware())
	{
		protected.POST("/v1/auth/me", server.Me)

		// Protected Walls Endpoint
		protected.GET("/v2/walls", server.getOwnWall)
		protected.GET("/v1/users/:id/walls", server.listWallsByUser)
		protected.POST("/v2/walls", server.createNewWall) //no test yet
	}

	// Set up routes for the user API
	router.POST("/api/v1/users", server.createUser) // working
	router.GET("/api/v1/users/:id", server.getUser) // working
	router.GET("/api/v1/users", server.listUsers)   // working
	router.PUT("/api/v1/users/:id", server.updateUser)
	router.DELETE("/api/v1/users/:id", server.deleteUser)         // working, but maybe need to add id of the deleted item in the response
	router.PUT("/api/v1/users/:id/profile", server.updateProfile) // working, but need fixing so that i can just edit 1 field at a time, not having to fill all fields just to edit 1 field
	router.PUT("/api/v1/users/:id/onboarding", server.finishOnboarding)

	// Set up routes for the wall API
	router.POST("/api/v1/walls", server.createWall)
	router.GET("/api/v1/walls/:id", server.getWall)                 // working
	router.GET("/api/v1/walls", server.listWalls)                   // working
	router.PUT("/api/v1/walls/:id", server.updateWall)              // working
	router.PUT("/api/v1/walls/:id/publicize", server.publicizeWall) // working
	router.PUT("/api/v1/walls/:id/privatize", server.privatizeWall) // working
	router.PUT("/api/v1/walls/:id/archive", server.archiveWall)     // need to add wall id and archive status
	router.PUT("/api/v1/walls/:id/unarchive", server.unarchiveWall) // need to add wall id and archive status
	router.DELETE("/api/v1/walls/:id", server.deleteWall)           // working

	// Set up routes for the post API
	router.POST("/api/v1/posts", server.createPost) // working
	router.GET("/api/v1/posts/:id", server.getPost) //
	router.GET("/api/v1/posts", server.listPosts)
	router.GET("/api/v1/walls/:id/posts", server.listPostsByWall)
	router.GET("/api/v2/walls/:id/posts", server.listPostsByWallWithAuthorsDetails) // no test yet
	router.PUT("/api/v1/posts/:id", server.updatePost)
	router.PUT("/api/v1/posts/:id/highlight", server.highlightPost)
	router.DELETE("/api/v1/posts/:id", server.deletePost)
	router.GET("/api/v1/posts/highlighted", server.getHighlightedPosts)                 // working
	router.GET("/api/v1/walls/:id/posts/highlighted", server.getHighlightedPostsByWall) // working
	router.PUT("/api/v1/posts/:id/unhighlight", server.unhighlightPost)                 // working

	// Updated Friendship API routes
	// Friend Requests
	router.POST("/api/v1/friend-requests", server.createFriendRequest)       // working
	router.PUT("/api/v1/friend-requests/accept", server.acceptFriendRequest) // working

	// User Blocking
	router.PUT("/api/v1/users/block", server.blockUser)
	router.PUT("/api/v1/users/unblock", server.unblockUser)

	router.GET("/api/v1/friends", server.getFriendsByStatus) //status = friends, requested, sent

	// Friends Retrieval
	router.GET("/api/v1/users/:id/accepted-friends", server.getFriends)                              // user_id; working
	router.GET("/api/v1/users/:id/friend-requests/pending", server.getPendingFriendRequests)         // user_id; working
	router.GET("/api/v2/users/:id/friend-requests/pending", server.getReceivedPendingFriendRequests) // user_id; working
	router.GET("/api/v1/users/:id/friend-requests/sent", server.getSentFriendRequests)               // user_id; working
	router.GET("/api/v2/users/:id/friend-requests/sent", server.getSentPendingFriendRequests)        // user_id; working

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
