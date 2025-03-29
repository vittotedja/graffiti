package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
	"net/http"
	"time"
)

// Server serves HTTP requests
type Server struct {
	hub        *db.Hub
	db         *pgxpool.Pool
	config     util.Config
	router     *gin.Engine
	httpServer *http.Server
}

// NewServer creates a new HTTP server and sets up all routes
func NewServer(config util.Config) *Server {
	s := &Server{
		config: config,
		router: gin.Default(),
	}
	s.registerRoutes()
	return s
}

// Start runs the HTTP server on a specific address
// and returns an error if the server fails to start
func (s *Server) Start() error {
	// Init DB
	ctx := context.Background()

	connPool, err := pgxpool.New(ctx, s.config.DBSource)
	if err != nil {
		logger.Global().Error("Cannot connect to DB", err)
		return fmt.Errorf("cannot connect to db: %w", err)
	}
	s.db = connPool
	s.hub = db.NewHub(connPool)

	// Set up HTTP server
	s.httpServer = &http.Server{
		Addr:    s.config.ServerAddress,
		Handler: s.router,
	}

	logger.Global().Info("Server listening on %s", s.config.ServerAddress)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Global().Info("Attempting graceful shutdown...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Global().Error("HTTP server shutdown error", err)
		return fmt.Errorf("server shutdown failed: %v", err)
	}

	if s.db != nil {
		s.db.Close()
		logger.Global().Info("Database connection closed.")
	}

	logger.Global().Info("Server shut down successfully.")
	return nil
}

func (s *Server) registerRoutes() {
	// Set up routes for the user API
	s.router.POST("/api/v1/users", s.createUser) // working
	s.router.GET("/api/v1/users/:id", s.getUser)
	s.router.GET("/api/v1/users", s.listUsers)      // working
	s.router.PUT("/api/v1/users/:id", s.updateUser) // working
	s.router.DELETE("/api/v1/users/:id", s.deleteUser)
	s.router.PUT("/api/v1/users/:id/profile", s.updateProfile)
	s.router.PUT("/api/v1/users/:id/onboarding", s.finishOnboarding) // working

	// Set up routes for the wall API
	s.router.POST("/api/v1/walls", s.createWall)
	s.router.GET("/api/v1/walls/:id", s.getWall)                 // working
	s.router.GET("/api/v1/walls", s.listWalls)                   // working
	s.router.GET("/api/v1/users/:id/walls", s.listWallsByUser)   // working
	s.router.PUT("/api/v1/walls/:id", s.updateWall)              // working
	s.router.PUT("/api/v1/walls/:id/publicize", s.publicizeWall) // working
	s.router.PUT("/api/v1/walls/:id/privatize", s.privatizeWall) // working
	s.router.PUT("/api/v1/walls/:id/archive", s.archiveWall)
	s.router.PUT("/api/v1/walls/:id/unarchive", s.unarchiveWall)
	s.router.DELETE("/api/v1/walls/:id", s.deleteWall) // working

	// Set up routes for the post API
	s.router.POST("/api/v1/posts", s.createPost)                                     // working
	s.router.GET("/api/v1/posts/:id", s.getPost)                                     // working
	s.router.GET("/api/v1/posts", s.listPosts)                                       // working
	s.router.GET("/api/v1/walls/:id/posts", s.listPostsByWall)                       // working
	s.router.GET("/api/v1/posts/highlighted", s.getHighlightedPosts)                 // working
	s.router.GET("/api/v1/walls/:id/posts/highlighted", s.getHighlightedPostsByWall) // working
	s.router.PUT("/api/v1/posts/:id", s.updatePost)                                  // working
	s.router.PUT("/api/v1/posts/:id/highlight", s.highlightPost)                     // working
	s.router.PUT("/api/v1/posts/:id/unhighlight", s.unhighlightPost)                 // working
	s.router.DELETE("/api/v1/posts/:id", s.deletePost)                               // working

	// Updated Friendship API routes
	// Friend Requests
	s.router.POST("/api/v1/friend-requests", s.createFriendRequest)       // working
	s.router.PUT("/api/v1/friend-requests/accept", s.acceptFriendRequest) // working

	// User Blocking
	s.router.PUT("/api/v1/users/block", s.blockUser)
	s.router.PUT("/api/v1/users/unblock", s.unblockUser)

	// Friends Retrieval
	s.router.GET("/api/v1/users/:id/accepted-friends", s.getFriends)                      // user_id; working
	s.router.GET("/api/v1/users/:id/friend-requests/pending", s.getPendingFriendRequests) // user_id; working
	s.router.GET("/api/v1/users/:id/friend-requests/sent", s.getSentFriendRequests)       // user_id; working

	// Existing friendship-related routes
	s.router.GET("/api/v1/users/:id/friendships", s.listFriendshipsByUserId)                            // working
	s.router.GET("/api/v1/users/:id/accepted-friends/count", s.getNumberOfFriends)                      // working
	s.router.GET("/api/v1/users/:id/friend-requests/pending/count", s.getNumberOfPendingFriendRequests) // working
	// router for listFriendshipByUserPairs
	// s.router.GET("/api/v1/friendships", s.listFriendshipByUserPairs)

	// Set up routes for the like API
	s.router.POST("/api/v1/likes", s.createLike)
	s.router.GET("/api/v1/likes/:id", s.getLike)
	s.router.GET("/api/v1/likes", s.listLikes)
	s.router.GET("/api/v1/posts/:id/likes", s.listLikesByPost)
	s.router.GET("/api/v1/users/:id/likes", s.listLikesByUser)
	s.router.DELETE("/api/v1/likes", s.deleteLike)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
