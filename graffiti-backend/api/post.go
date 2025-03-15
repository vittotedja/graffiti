package api

import (
	"graffiti-backend/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Define the request structure according to the POST table schema
type createPostRequest struct {
    WallID        int       `json:"wall_id" binding:"required"`
    Author        string    `json:"author" binding:"required"`
    MediaURL      string    `json:"media_url" binding:"required"`
    PostType      string    `json:"post_type" binding:"required"`
    IsHighlighted bool      `json:"is_highlighted"`
    NoOfLikes     int       `json:"no_of_likes"`
    IsDeleted     bool      `json:"is_deleted"`
    CreatedAt     time.Time `json:"created_at" binding:"required"`
}

func (server *Server) createPost(ctx *gin.Context) {
    var req createPostRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    arg := db.CreatePostParams{
        WallID:        req.WallID,
        Author:        req.Author,
        MediaURL:      req.MediaURL,
        PostType:      req.PostType,
        IsHighlighted: req.IsHighlighted,
        NoOfLikes:     req.NoOfLikes,
        IsDeleted:     req.IsDeleted,
        CreatedAt:     req.CreatedAt,
    }

    post, err := server.hub.CreatePost(ctx, arg)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }
    ctx.JSON(http.StatusOK, post)
}