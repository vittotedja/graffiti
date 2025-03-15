package api

import (
	"graffiti-backend/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Define the request structure according to the FRIENDS table schema
type befriendRequest struct {
    UserID1   int       `json:"user_id_1" binding:"required"`
    UserID2   int       `json:"user_id_2" binding:"required"`
    Status    string    `json:"status" binding:"required,oneof=pending accepted rejected"`
    CreatedAt time.Time `json:"created_at" binding:"required"`
    UpdatedAt time.Time `json:"updated_at" binding:"required"`
}

func (server *Server) befriend(ctx *gin.Context) {
    var req befriendRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    // Validate that both users exist
    user1Exists, err := server.CheckUserExists(ctx, req.UserID1)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }
    if !user1Exists {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID 1 does not exist"})
        return
    }

    user2Exists, err := server.CheckUserExists(ctx, req.UserID2)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }
    if !user2Exists {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID 2 does not exist"})
        return
    }

    // Ensure the friendship is not already created
    exists, err := server.CheckFriendshipExists(ctx, db.CheckFriendshipExistsParams{
        UserID1: req.UserID1,
        UserID2: req.UserID2,
    })
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }
    if exists {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Friendship already exists"})
        return
    }

    // Create the friendship
    arg := db.CreateFriendParams{
        UserID1:   req.UserID1,
        UserID2:   req.UserID2,
        Status:    req.Status,
        CreatedAt: req.CreatedAt,
        UpdatedAt: req.UpdatedAt,
    }

    friendship, err := server.hub.Befriend(ctx, arg)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }
    ctx.JSON(http.StatusOK, friendship)
}