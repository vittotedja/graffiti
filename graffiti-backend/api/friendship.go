package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
)

// Friendship Request Structs

type createFriendRequestRequest struct {
	FromUserID string `json:"from_user_id" binding:"required"`
	ToUserID   string `json:"to_user_id" binding:"required"`
}

type acceptFriendRequestRequest struct {
	FriendshipID string `json:"friendship_id" binding:"required"`
}

type blockUserRequest struct {
	FromUserID string `json:"from_user_id" binding:"required"`
	ToUserID   string `json:"to_user_id" binding:"required"`
}

type unblockUserRequest struct {
	FromUserID string `json:"from_user_id" binding:"required"`
	ToUserID   string `json:"to_user_id" binding:"required"`
}

// CreateFriendRequest handles creating a new friend request
func (server *Server) createFriendRequest(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received create friend request")

	var req createFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validate user IDs
	var fromUserID, toUserID pgtype.UUID
	if err := fromUserID.Scan(req.FromUserID); err != nil {
		log.Error("Invalid from_user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := toUserID.Scan(req.ToUserID); err != nil {
		log.Error("Invalid to_user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Prevent self-friending
	if fromUserID == toUserID {
		log.Errorf("Attempt to send friend request to self for user %s", fromUserID)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot send friend request to yourself"})
		return
	}

	// Use transaction method from hub
	friendship, err := server.hub.CreateFriendRequestTx(ctx, fromUserID, toUserID)
	if err != nil {
		log.Error("Failed to create friend request", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Friend request created successfully")
	ctx.JSON(http.StatusOK, friendship)
}

// ListFriendshipsByUserId retrieves all friendships for a specific user
func (server *Server) listFriendshipsByUserId(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list friendships by user ID request")

	userIDStr := ctx.Param("id")
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		log.Error("Invalid user ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	friendships, err := server.hub.ListFriendshipsByUserId(ctx, userID)
	if err != nil {
		log.Error("Failed to list friendships", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Listed friendships successfully")
	ctx.JSON(http.StatusOK, friendships)
}

// GetNumberOfFriends retrieves the number of friends for a specific user
func (server *Server) getNumberOfFriends(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get number of friends request")

	userIDStr := ctx.Param("id")
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		log.Error("Invalid user ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	count, err := server.hub.GetNumberOfFriends(ctx, userID)
	if err != nil {
		log.Error("Failed to get number of friends", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Retrieved number of friends successfully")
	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// AcceptFriendRequest handles accepting a pending friend request
func (server *Server) acceptFriendRequest(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received accept friend request")

	var req acceptFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var friendshipID pgtype.UUID
	if err := friendshipID.Scan(req.FriendshipID); err != nil {
		log.Error("Invalid friendship ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := server.hub.AcceptFriendRequestTx(ctx, friendshipID); err != nil {
		log.Error("Failed to accept friend request", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Friend request accepted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
}

// BlockUser handles blocking a user
func (server *Server) blockUser(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received block user request")

	var req blockUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validate user IDs
	var fromUserID, toUserID pgtype.UUID
	if err := fromUserID.Scan(req.FromUserID); err != nil {
		log.Error("Invalid from_user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := toUserID.Scan(req.ToUserID); err != nil {
		log.Error("Invalid to_user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Prevent self-blocking
	if fromUserID == toUserID {
		log.Info("Attempt to block self")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot block yourself"})
		return
	}

	if err := server.hub.BlockUserTx(ctx, fromUserID, toUserID); err != nil {
		log.Error("Failed to block user", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("User blocked successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "User blocked successfully"})
}

// UnblockUser handles unblocking a previously blocked user
func (server *Server) unblockUser(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received unblock user request")

	var req unblockUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validate user IDs
	var fromUserID, toUserID pgtype.UUID
	if err := fromUserID.Scan(req.FromUserID); err != nil {
		log.Error("Invalid from_user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := toUserID.Scan(req.ToUserID); err != nil {
		log.Error("Invalid to_user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := server.hub.UnblockUserTx(ctx, fromUserID, toUserID); err != nil {
		log.Error("Failed to unblock user", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("User unblocked successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "User unblocked successfully"})
}

// GetFriends retrieves all friends for a specific user
func (server *Server) getFriends(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get friends request")

	userIDStr := ctx.Param("id")
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		log.Error("Invalid user ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	friends, err := server.hub.GetFriendsTx(ctx, userID)
	if err != nil {
		log.Error("Failed to get friends", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Retrieved friends successfully")
	ctx.JSON(http.StatusOK, friends)
}

func (server *Server) getFriendsByStatus(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(db.User)

	queryType := ctx.Query("type")

	arg := db.ListFriendsDetailsByStatusParams{
		FromUser: user.ID,
		Column2:  queryType,
	}

	switch queryType {
	case "friends", "sent", "requested":
		friends, err := server.hub.ListFriendsDetailsByStatus(ctx, arg)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		ctx.JSON(http.StatusOK, friends)
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
	}
}

// GetPendingFriendRequests retrieves pending friend requests for a user
func (server *Server) getPendingFriendRequests(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get pending friend requests request")

	userIDStr := ctx.Param("id")
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		log.Error("Invalid user ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	pendingRequests, err := server.hub.GetPendingFriendRequestsTx(ctx, userID)
	if err != nil {
		log.Error("Failed to get pending friend requests", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Retrieved pending friend requests successfully")
	ctx.JSON(http.StatusOK, pendingRequests)
}

func (server *Server) getReceivedPendingFriendRequests(ctx *gin.Context) {
	userIDStr := ctx.Param("id")

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	pendingRequests, err := server.hub.ListReceivedPendingFriendRequests(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, pendingRequests)
}

// GetNumberOfPendingFriendRequests retrieves the number of pending friend requests for a user
func (server *Server) getNumberOfPendingFriendRequests(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get number of pending friend requests request")

	userIDStr := ctx.Param("id")
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		log.Error("Invalid user ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	count, err := server.hub.GetNumberOfPendingFriendRequests(ctx, userID)
	if err != nil {
		log.Error("Failed to get number of pending friend requests", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Retrieved number of pending friend requests successfully")
	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// GetSentFriendRequests retrieves friend requests sent by a user
func (server *Server) getSentFriendRequests(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get sent friend requests request")

	userIDStr := ctx.Param("id")
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		log.Error("Invalid user ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	sentRequests, err := server.hub.GetSentFriendRequestsTx(ctx, userID)
	if err != nil {
		log.Error("Failed to get sent friend requests", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Retrieved sent friend requests successfully")
	ctx.JSON(http.StatusOK, sentRequests)
}

func (server *Server) getSentPendingFriendRequests(ctx *gin.Context) {
	userIDStr := ctx.Param("id")

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	pendingRequests, err := server.hub.ListSentPendingFriendRequests(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, pendingRequests)
}

// ListFriendshipByUserPairs retrieves a friendship by user pairs
func (server *Server) listFriendshipByUserPairs(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list friendship by user pairs request")

	var req createFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validate user IDs
	var fromUserID, toUserID pgtype.UUID
	if err := fromUserID.Scan(req.FromUserID); err != nil {
		log.Error("Invalid from_user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := toUserID.Scan(req.ToUserID); err != nil {
		log.Error("Invalid to_user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	params := db.ListFriendshipByUserPairsParams{
		FromUser: fromUserID,
		ToUser:   toUserID,
	}

	friendship, err := server.hub.Queries.ListFriendshipByUserPairs(ctx, params)
	if err != nil {
		log.Error("Failed to list friendship by user pairs", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Listed friendship by user pairs successfully")
	ctx.JSON(http.StatusOK, friendship)
}
