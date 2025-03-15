package api

import (
	"graffiti-backend/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Define the request structure according to the LIKES table schema
type createLikeRequest struct {
    PostID   int       `json:"post_id" binding:"required"`
    UserID   int       `json:"user_id" binding:"required"`
    LikedAt  time.Time `json:"liked_at" binding:"required"`
}

func (server *Server) createLike(ctx *gin.Context) {
    var req createLikeRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    arg := db.CreateLikeParams{
        PostID:   req.PostID,
        UserID:   req.UserID,
        LikedAt:  req.LikedAt,
    }

    like, err := server.hub.CreateLike(ctx, arg)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }
    ctx.JSON(http.StatusOK, like)
}