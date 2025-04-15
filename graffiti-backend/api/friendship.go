package api

import (
	"database/sql"
	"errors"
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
)

// Friendship Request Structs

type createFriendRequestRequest struct {
	ToUserID string `json:"to_user_id" binding:"required"`
}

type FriendshipIDRequest struct {
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

type getNumberOfMutualFriendsRequest struct {
	UserID1 string `json:"user_id_1" binding:"required"`
	UserID2 string `json:"user_id_2" binding:"required"`
}

type getMutualFriendsRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type MutualFriendsResponse struct {
	ID                string `json:"id"`
	Username          string `json:"username"`
	FullName          string `json:"fullname"`
	ProfilePicture    string `json:"profile_picture"`
	BackgroundImage   string `json:"background_image"`
	MutualFriendCount int    `json:"mutual_friend_count"`
}

// CreateFriendRequest handles creating a new friend request
func (s *Server) createFriendRequest(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received create friend request")

	user := ctx.MustGet("currentUser").(db.User)

	fromUserID := user.ID

	var req createFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validate user IDs
	var toUserID pgtype.UUID
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
	friendship, err := s.hub.CreateFriendRequestTx(ctx, fromUserID, toUserID)
	if err != nil {
		log.Error("Failed to create friend request", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	
	// Send notification to the recipient
	err = s.SendNotification(
        ctx,
        toUserID.String(),
        user.ID.String(),
        "friend_request",
		friendship.ID.String(),
        fmt.Sprintf("%s sent you a friend request", user.Username),
    )
    if err != nil {
        // Just log the error, don't fail the request
        log.Error("Failed to send notification", err)
    }


	log.Info("Friend request created successfully")
	ctx.JSON(http.StatusOK, friendship)
}

// ListFriendshipsByUserId retrieves all friendships for a specific user
func (s *Server) listFriendshipsByUserId(ctx *gin.Context) {
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

	friendships, err := s.hub.ListFriendshipsByUserId(ctx, userID)
	if err != nil {
		log.Error("Failed to list friendships", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Listed friendships successfully")
	ctx.JSON(http.StatusOK, friendships)
}

// GetNumberOfFriends retrieves the number of friends for a specific user
func (s *Server) getNumberOfFriends(ctx *gin.Context) {
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

	count, err := s.hub.GetNumberOfFriends(ctx, userID)
	if err != nil {
		log.Error("Failed to get number of friends", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Retrieved number of friends successfully")
	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// AcceptFriendRequest handles accepting a pending friend request
func (s *Server) acceptFriendRequest(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received accept friend request")

	currentUser, ok := ctx.MustGet("currentUser").(db.User)
	if !ok {
		log.Error("Failed to get current user from context", errors.New("unauthorized"))
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized")))
		return
	}

	var req FriendshipIDRequest
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

	friendships, err := s.hub.GetFriendship(ctx, friendshipID)
	if err != nil {
		log.Error("Failed to get friendship", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if currentUser.ID != friendships.FromUser && currentUser.ID != friendships.ToUser {
		log.Error("Unauthorized to delete friendship", errors.New("unauthorized"))
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized")))
		return
	}

	if err := s.hub.AcceptFriendRequestTx(ctx, friendshipID); err != nil {
		log.Error("Failed to accept friend request", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Friend request accepted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
}

// BlockUser handles blocking a user
func (s *Server) blockUser(ctx *gin.Context) {
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

	if err := s.hub.BlockUserTx(ctx, fromUserID, toUserID); err != nil {
		log.Error("Failed to block user", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("User blocked successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "User blocked successfully"})
}

// UnblockUser handles unblocking a previously blocked user
func (s *Server) unblockUser(ctx *gin.Context) {
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

	if err := s.hub.UnblockUserTx(ctx, fromUserID, toUserID); err != nil {
		log.Error("Failed to unblock user", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("User unblocked successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "User unblocked successfully"})
}

// GetFriends retrieves all friends for a specific user
func (s *Server) getFriends(ctx *gin.Context) {
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

	friends, err := s.hub.GetFriendsTx(ctx, userID)
	if err != nil {
		log.Error("Failed to get friends", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Retrieved friends successfully")
	ctx.JSON(http.StatusOK, friends)
}

func (s *Server) getFriendsByStatus(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(db.User)

	queryType := ctx.Query("type")

	arg := db.ListFriendsDetailsByStatusParams{
		FromUser: user.ID,
		Column2:  queryType,
	}

	switch queryType {
	case "friends", "sent", "requested":
		friends, err := s.hub.ListFriendsDetailsByStatus(ctx, arg)
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
func (s *Server) getPendingFriendRequests(ctx *gin.Context) {
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

	pendingRequests, err := s.hub.GetPendingFriendRequestsTx(ctx, userID)
	if err != nil {
		log.Error("Failed to get pending friend requests", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Retrieved pending friend requests successfully")
	ctx.JSON(http.StatusOK, pendingRequests)
}

// GetNumberOfPendingFriendRequests retrieves the number of pending friend requests for a user
func (s *Server) getNumberOfPendingFriendRequests(ctx *gin.Context) {
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

	count, err := s.hub.GetNumberOfPendingFriendRequests(ctx, userID)
	if err != nil {
		log.Error("Failed to get number of pending friend requests", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Retrieved number of pending friend requests successfully")
	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// GetSentFriendRequests retrieves friend requests sent by a user
func (s *Server) getSentFriendRequests(ctx *gin.Context) {
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

	sentRequests, err := s.hub.GetSentFriendRequestsTx(ctx, userID)
	if err != nil {
		log.Error("Failed to get sent friend requests", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Retrieved sent friend requests successfully")
	ctx.JSON(http.StatusOK, sentRequests)
}

// ListFriendshipByUserPairs retrieves a friendship by user pairs
func (s *Server) listFriendshipByUserPairs(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list friendship by user pairs request")

	var req createFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user := ctx.MustGet("currentUser").(db.User)

	fromUserID := user.ID

	// Validate user IDs
	var toUserID pgtype.UUID
	if err := toUserID.Scan(req.ToUserID); err != nil {
		log.Error("Invalid to_user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if fromUserID == toUserID {
		log.Errorf("User pairs could not be the same user for user %s", fromUserID)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "You are not able to befriend yourself"})
		return
	}

	params := db.ListFriendshipByUserPairsParams{
		FromUser: fromUserID,
		ToUser:   toUserID,
	}

	friendship, err := s.hub.ListFriendshipByUserPairs(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No friendship found, but not an error you want to report
			ctx.JSON(http.StatusOK, gin.H{
				"ID":     "00000000",
				"Status": gin.H{"Status": "none", "Valid": true},
			})
			return
		}
		log.Error("Failed to list friendship by user pairs", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Listed friendship by user pairs successfully")
	ctx.JSON(http.StatusOK, friendship)
}

func (s *Server) deleteFriendship(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received delete friendship request")

	_, ok := ctx.MustGet("currentUser").(db.User)
	if !ok {
		log.Error("User not found in context", nil)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req FriendshipIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var friendshipID pgtype.UUID
	if err := friendshipID.Scan(req.FriendshipID); err != nil {
		log.Error("Invalid FriendshipID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := s.hub.DeleteFriendship(ctx, friendshipID)
	if err != nil {
		log.Error("Failed to delete friendship", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Friendship deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Friendship deleted successfully"})
}

func (s *Server) getNumberOfMutualFriends(ctx *gin.Context) {
	var req getNumberOfMutualFriendsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var userID1, userID2 pgtype.UUID
	if err := userID1.Scan(req.UserID1); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := userID2.Scan(req.UserID2); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	count, err := s.hub.GetNumberOfMutualFriends(ctx, db.GetNumberOfMutualFriendsParams{
		UserID:   userID1,
		UserID_2: userID2,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"mutual_friends_count": count})
}

func (s *Server) discoverFriendsByMutuals(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received discover friends by mutual request")

	currentUser, ok := ctx.MustGet("currentUser").(db.User)
	if !ok {
		log.Error("User not found in context", nil)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	results, err := s.hub.DiscoverFriendsByMutuals(ctx, currentUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var resp []MutualFriendsResponse

	for _, u := range results {
		resp = append(resp, MutualFriendsResponse{
			ID:                u.ID.String(),
			Username:          u.Username,
			FullName:          u.Fullname.String,
			ProfilePicture:    u.ProfilePicture.String,
			BackgroundImage:   u.BackgroundImage.String,
			MutualFriendCount: int(u.MutualFriendCount),
		})
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *Server) getMutualFriends(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get mutual friends")

	currentUser, ok := ctx.MustGet("currentUser").(db.User)
	if !ok {
		log.Error("User not found in context", nil)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req getMutualFriendsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var userID pgtype.UUID
	if err := userID.Scan(req.UserID); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	mutuals, err := s.hub.ListMutualFriends(ctx, db.ListMutualFriendsParams{
		UserID:   userID,
		UserID_2: currentUser.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var resp []UserSearchResponse
	for _, u := range mutuals {
		resp = append(resp, UserSearchResponse{
			ID:             u.ID.String(),
			Username:       u.Username,
			FullName:       u.Fullname.String,
			ProfilePicture: u.ProfilePicture.String,
		})
	}

	ctx.JSON(http.StatusOK, resp)
}
