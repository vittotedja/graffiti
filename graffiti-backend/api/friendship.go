package api

import (
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// Friendship handlers

type createFriendshipRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	FriendID string `json:"friend_id" binding:"required"`
	Status   string `json:"status" binding:"required"`
}

func (server *Server) createFriendship(ctx *gin.Context) {
	var req createFriendshipRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var userID pgtype.UUID
	err := userID.Scan(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var friendID pgtype.UUID
	err = friendID.Scan(req.FriendID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateFriendshipParams{
		UserID:   userID,
		FriendID: friendID,
		Status:   db.NullStatus{Status: db.Status(req.Status), Valid: true},
	}

	friendship, err := server.hub.CreateFriendship(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}

type getFriendshipRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (server *Server) getFriendship(ctx *gin.Context) {
	var req getFriendshipRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	friendship, err := server.hub.GetFriendship(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}

type updateFriendshipRequest struct {
	ID     string `json:"id" binding:"required"`
	Status string `json:"status" binding:"required"`
}

func (server *Server) updateFriendship(ctx *gin.Context) {
	var req updateFriendshipRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateFriendshipParams{
		ID:     id,
		Status: db.NullStatus{Status: db.Status(req.Status), Valid: true},
	}

	friendship, err := server.hub.UpdateFriendship(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}

type deleteFriendshipRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (server *Server) deleteFriendship(ctx *gin.Context) {
	var req deleteFriendshipRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.hub.DeleteFriendship(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Friendship deleted successfully"})
}

func (server *Server) listFriendships(ctx *gin.Context) {
	friendships, err := server.hub.ListFriendships(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, friendships)
}

type getNumberOfFriendsRequest struct {
	UserID string `uri:"user_id" binding:"required"`
}

func (server *Server) getNumberOfFriends(ctx *gin.Context) {
	var req getNumberOfFriendsRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var userID pgtype.UUID
	err := userID.Scan(req.UserID)
	if err != nil {
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

type getNumberOfPendingFriendRequestsRequest struct {
	FriendID string `uri:"friend_id" binding:"required"`
}

func (server *Server) getNumberOfPendingFriendRequests(ctx *gin.Context) {
	var req getNumberOfPendingFriendRequestsRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var friendID pgtype.UUID
	err := friendID.Scan(req.FriendID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	count, err := server.hub.GetNumberOfPendingFriendRequests(ctx, friendID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}