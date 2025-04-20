package api

import (
	"log"
	"time"
	"errors"

	cron "github.com/vittotedja/graffiti/graffiti-backend/util/cron"

	"context"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/google/uuid"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/token"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
)

// Server serves HTTP requests
type Server struct {
	hub        db.Hub
	db         *pgxpool.Pool
	config     util.Config
	router     *gin.Engine
	tokenMaker token.Maker
	httpServer *http.Server
}

func NewServer(config util.Config) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		log.Fatal("cannot create token maker", err)
	}
	server := &Server{config: config, router: gin.Default(), tokenMaker: tokenMaker}
	server.router.Use(logger.Middleware())
	server.registerRoutes("server")

    if config.SQSQueueURL != "" {
        ctx := context.Background()
        server.StartNotificationWorker(ctx)
    }

	return server, nil
}

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

	cron.ScheduleMaterializedViewRefresh(s.db)

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

func (s *Server) StartNotificationWorker(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(30 * time.Second) // Poll every 30 seconds
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                err := s.ProcessNotifications(ctx)
                if err != nil {
                    log.Printf("Error processing notifications: %v\n", err)
                }
            }
        }
    }()
}

func (s *Server) registerRoutes(env string) {

	frontendURL := s.config.FrontendURL
	// Apply CORS middleware
	if (env == "unit-test") {
		s.router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"}, 
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
	} else {
		s.router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{frontendURL}, 
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	// ALB Health check endpoint
	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Healthy")
	})

	s.router.POST("/api/v1/auth/register", s.Register)
	s.router.POST("/api/v1/auth/login", s.Login)
	s.router.POST("/api/v1/auth/logout", s.Logout)

	protected := s.router.Group("/api")
	if env != "unit-test" {
		protected.Use(s.AuthMiddleware())
	}
	{
		// auth
		protected.POST("/v1/auth/me", s.Me)
		// users
		protected.GET("/v1/users/:id", s.getUser)
		protected.POST("/v2/users", s.updateUserNew) // no test

		// Protected Walls Endpoint
		protected.GET("/v1/walls/:id", s.getWall) // working
		protected.GET("/v2/walls", s.getOwnWall)
		protected.GET("/v1/users/:id/walls", s.listWallsByUser)
		protected.POST("/v2/walls", s.createNewWall)              
		protected.PUT("/v1/walls/:id", s.updateWall)              
		protected.PUT("/v1/walls/:id/publicize", s.publicizeWall) 
		protected.PUT("/v1/walls/:id/privatize", s.privatizeWall) 
		protected.PUT("/v1/walls/:id/pin", s.pinWall)           
		protected.DELETE("/v1/walls/:id", s.deleteWall)

		protected.GET("/v1/walls/archived", s.getArchivedWalls)
		protected.PUT("/v1/walls/:id/archive", s.archiveWall)
		protected.PUT("/v1/walls/:id/unarchive", s.unarchiveWall)

		// search
		protected.POST("/v1/users/search", s.searchUsers)

		//uploads
		protected.POST("/v1/presign", s.presignHandler)

		//friends
		protected.POST("/v1/friend-requests", s.createFriendRequest) 
		protected.POST("/v1/friendships", s.listFriendshipByUserPairs)
		protected.GET("/v1/friends", s.getFriendsByStatus)                
		protected.PUT("/v1/friend-requests/accept", s.acceptFriendRequest) 
		protected.DELETE("/v1/friendships", s.deleteFriendship)

		//posts
		protected.GET("/v2/walls/:id/posts", s.listPostsByWallWithAuthorsDetails) 
		protected.DELETE("/v1/posts/:id", s.deletePost)
		protected.POST("/v1/posts", s.createPost)

		//likes
		protected.POST("/v1/likes", s.updateLike)
		protected.GET("/v1/likes/:post_id", s.getLike)

		//discover
		protected.POST("/v1/friends/discover", s.discoverFriendsByMutuals)
		protected.POST("/v1/friends/mutual", s.getMutualFriends)

		//notifications
		protected.GET("/v1/notifications", s.getNotifications)
		protected.PUT("/v1/notifications/:id/read", s.markNotificationAsRead)
		protected.PUT("/v1/notifications/read-all", s.markAllNotificationsAsRead)
		protected.GET("/v1/notifications/unread/count", s.getUnreadNotificationsCount)
	}

	s.router.GET("/api/v1/users", s.listUsers)
	s.router.DELETE("/api/v1/users/:id", s.deleteUser)
	s.router.PUT("/api/v1/users/:id/onboarding", s.finishOnboarding)

	s.router.GET("/api/v1/walls", s.listWalls)

	s.router.GET("/api/v1/posts/:id", s.getPost)                                    
	s.router.GET("/api/v1/posts", s.listPosts)                                      
	s.router.GET("/api/v1/walls/:id/posts", s.listPostsByWall)                       
	s.router.GET("/api/v1/posts/highlighted", s.getHighlightedPosts)                 
	s.router.GET("/api/v1/walls/:id/posts/highlighted", s.getHighlightedPostsByWall) 
	s.router.PUT("/api/v1/posts/:id", s.updatePost)                                  
	s.router.PUT("/api/v1/posts/:id/highlight", s.highlightPost)                    
	s.router.PUT("/api/v1/posts/:id/unhighlight", s.unhighlightPost)                

	// Friend Requests
	s.router.POST("/api/v1/friends/mutual/count", s.getNumberOfMutualFriends)

	// User Blocking
	s.router.PUT("/api/v1/users/block", s.blockUser)
	s.router.PUT("/api/v1/users/unblock", s.unblockUser)

	// Friends Retrieval
	s.router.GET("/api/v1/users/:id/accepted-friends", s.getFriends)                      
	s.router.GET("/api/v1/users/:id/friend-requests/pending", s.getPendingFriendRequests)
	s.router.GET("/api/v1/users/:id/friend-requests/sent", s.getSentFriendRequests)      

	// Existing friendship-related routes
	s.router.GET("/api/v1/users/:id/friendships", s.listFriendshipsByUserId)                            
	s.router.GET("/api/v1/users/:id/accepted-friends/count", s.getNumberOfFriends)                      
	s.router.GET("/api/v1/users/:id/friend-requests/pending/count", s.getNumberOfPendingFriendRequests) 
	// router for listFriendshipByUserPairs

	// Set up routes for the like API
	s.router.GET("/api/v1/likes", s.listLikes)
	s.router.GET("/api/v1/posts/:id/likes", s.listLikesByPost)
	s.router.GET("/api/v1/users/:id/likes", s.listLikesByUser)
	s.router.DELETE("/api/v1/likes", s.deleteLike)
}

// NotificationResponse represents the response for a notification
type NotificationResponse struct {
    ID          string    `json:"id"`
    RecipientID string    `json:"recipient_id"`
    SenderID    string    `json:"sender_id"`
    Type        string    `json:"type"`
    EntityID    string    `json:"entity_id"`
    Message     string    `json:"message"`
    IsRead      bool      `json:"is_read"`
    CreatedAt   time.Time `json:"created_at"`
}

// getNotifications handles requests to get a user's notifications
func (s *Server) getNotifications(ctx *gin.Context) {
    user := ctx.MustGet("currentUser").(db.User)
    
    notifications, err := s.hub.GetNotificationsByUser(ctx, pgtype.UUID{Bytes: user.ID.Bytes, Valid: true})
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
        return
    }
    
    var response []NotificationResponse
    for _, notification := range notifications {
        response = append(response, NotificationResponse{
            ID:          notification.ID.String(),
            RecipientID: notification.RecipientID.String(),
            SenderID:    notification.SenderID.String(),
            Type:        notification.Type,
            EntityID:    notification.EntityID.String(),
            Message:     notification.Message,
            IsRead:      notification.IsRead.Bool,
            CreatedAt:   notification.CreatedAt.Time,
        })
    }
    
    ctx.JSON(http.StatusOK, response)
}

// markNotificationAsRead handles requests to mark a notification as read
func (s *Server) markNotificationAsRead(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
        return
    }
    
    err = s.hub.MarkNotificationAsRead(ctx, pgtype.UUID{Bytes: id, Valid: true})
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
        return
    }
    
    ctx.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// markAllNotificationsAsRead handles requests to mark all notifications as read
func (s *Server) markAllNotificationsAsRead(ctx *gin.Context) {
    user := ctx.MustGet("currentUser").(db.User)
    
    err := s.hub.MarkAllNotificationsAsRead(ctx, pgtype.UUID{Bytes: user.ID.Bytes, Valid: true})
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all notifications as read"})
        return
    }
    
    ctx.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}


func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

// getUnreadNotificationsCount handles requests to get the count of unread notifications for the current user
func (s *Server) getUnreadNotificationsCount(ctx *gin.Context) {
    meta := logger.GetMetadata(ctx.Request.Context())
    log := meta.GetLogger()
    log.Info("Received get unread notifications count request")

    currentUser, ok := ctx.MustGet("currentUser").(db.User)
    if !ok {
        log.Error("Failed to get current user from context", errors.New("unauthorized"))
        ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized")))
        return
    }
    
    count, err := s.hub.CountUnreadNotifications(ctx, pgtype.UUID{Bytes: currentUser.ID.Bytes, Valid: true})
    if err != nil {
        log.Error("Failed to count unread notifications", err)
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }
    
    log.Info("Unread notifications count retrieved successfully: %d", count)
    ctx.JSON(http.StatusOK, gin.H{"count": count})
}

