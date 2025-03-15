package api

import (
	"graffiti-backend/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Define the request structure according to the WALL table schema
type createWallRequest struct {
    UserID         		int       `json:"user_id" binding:"required"`
    Description    		string    `json:"description" binding:"required"`
    BackgroundImageURL  string    `json:"background" binding:"required"`
    IsPublic       		bool      `json:"is_public" binding:"required"`
    IsArchived     		bool      `json:"is_archived" binding:"required"`
    PopularityScore 	int      `json:"popularity_score"`
    CreatedAt      		time.Time `json:"created_at" binding:"required"`
    UpdatedAt      		time.Time `json:"updated_at"`
}

func (server *Server) createWall(ctx *gin.Context) {
    var req createWallRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    // Validate that the user exists

    // userExists, err := server.CheckUserExists(ctx, req.UserID)
    // if err != nil {
    //     ctx.JSON(http.StatusInternalServerError, errorResponse(err))
    //     return
    // }
    // if !userExists {
    //     ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID does not exist"})
    //     return
    // }

    arg := db.CreateWallParams{
        UserID:         req.UserID,
        Description:    req.Description,
        Background:     req.BackgroundImageURL,
        IsPublic:       req.IsPublic,
        IsArchived:     req.IsArchived,
        PopularityScore: req.PopularityScore,
        CreatedAt:      req.CreatedAt,
        UpdatedAt:      req.UpdatedAt,
    }

    wall, err := server.hub.CreateWall(ctx, arg)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }
    ctx.JSON(http.StatusOK, wall)
}