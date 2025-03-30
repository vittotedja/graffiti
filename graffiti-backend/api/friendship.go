package api

import (
	"net/http"

	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
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
	var req createFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validate user IDs
	var fromUserID, toUserID pgtype.UUID
	if err := fromUserID.Scan(req.FromUserID); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := toUserID.Scan(req.ToUserID); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Prevent self-friending
	if fromUserID == toUserID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot send friend request to yourself"})
		return
	}

	// Use transaction method from hub
	friendship, err := server.hub.CreateFriendRequestTx(ctx, fromUserID, toUserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}

// ListFriendshipsByUserId retrieves all friendships for a specific user
func (server *Server) listFriendshipsByUserId(ctx *gin.Context) {
	userIDStr := ctx.Param("id")

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	friendships, err := server.hub.ListFriendshipsByUserId(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, friendships)
}

// GetNumberOfFriends retrieves the number of friends for a specific user
func (server *Server) getNumberOfFriends(ctx *gin.Context) {
	userIDStr := ctx.Param("id")

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	count, err := server.hub.GetNumberOfFriends(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// AcceptFriendRequest handles accepting a pending friend request
func (server *Server) acceptFriendRequest(ctx *gin.Context) {
	var req acceptFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var friendshipID pgtype.UUID
	if err := friendshipID.Scan(req.FriendshipID); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := server.hub.AcceptFriendRequestTx(ctx, friendshipID); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
}

// BlockUser handles blocking a user
func (server *Server) blockUser(ctx *gin.Context) {
	var req blockUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validate user IDs
	var fromUserID, toUserID pgtype.UUID
	if err := fromUserID.Scan(req.FromUserID); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := toUserID.Scan(req.ToUserID); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Prevent self-blocking
	if fromUserID == toUserID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot block yourself"})
		return
	}

	if err := server.hub.BlockUserTx(ctx, fromUserID, toUserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User blocked successfully"})
}

// UnblockUser handles unblocking a previously blocked user
func (server *Server) unblockUser(ctx *gin.Context) {
	var req unblockUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validate user IDs
	var fromUserID, toUserID pgtype.UUID
	if err := fromUserID.Scan(req.FromUserID); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := toUserID.Scan(req.ToUserID); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := server.hub.UnblockUserTx(ctx, fromUserID, toUserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User unblocked successfully"})
}

// GetFriends retrieves all friends for a specific user
func (server *Server) getFriends(ctx *gin.Context) {
	userIDStr := ctx.Param("id")

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	friends, err := server.hub.GetFriendsTx(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, friends)
}

func (server *Server) getFriendsByStatus(ctx *gin.Context) {
	token, err := ctx.Cookie("token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	payload, err := server.tokenMaker.VerifyToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unathorized"})
		return
	}

	user, err := server.hub.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

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
	userIDStr := ctx.Param("id")

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	pendingRequests, err := server.hub.GetPendingFriendRequestsTx(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

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
	userIDStr := ctx.Param("id")

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	count, err := server.hub.GetNumberOfPendingFriendRequests(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// GetSentFriendRequests retrieves friend requests sent by a user
func (server *Server) getSentFriendRequests(ctx *gin.Context) {
	userIDStr := ctx.Param("id")

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	sentRequests, err := server.hub.GetSentFriendRequestsTx(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

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
	var req createFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validate user IDs
	var fromUserID, toUserID pgtype.UUID
	if err := fromUserID.Scan(req.FromUserID); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := toUserID.Scan(req.ToUserID); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	params := db.ListFriendshipByUserPairsParams{
		FromUser: fromUserID,
		ToUser:   toUserID,
	}

	friendship, err := server.hub.Queries.ListFriendshipByUserPairs(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}
