package api

import (
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// Wall request/response types
type createWallRequest struct {
	UserID          string `json:"user_id" binding:"required,uuid"`
	Description     string `json:"description"`
	BackgroundImage string `json:"background_image"`
}

type wallResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Description     string    `json:"description,omitempty"`
	BackgroundImage string    `json:"background_image,omitempty"`
	IsPublic        bool      `json:"is_public"`
	IsArchived      bool      `json:"is_archived"`
	IsDeleted       bool      `json:"is_deleted"`
	PopularityScore float64   `json:"popularity_score"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type updateWallRequest struct {
	Description     string `json:"description"`
	BackgroundImage string `json:"background_image"`
}

// Convert DB wall to API response
func newWallResponse(wall db.Wall) wallResponse {
	return wallResponse{
		ID:              wall.ID.String(),
		UserID:          wall.UserID.String(),
		Description:     wall.Description.String,
		BackgroundImage: wall.BackgroundImage.String,
		IsPublic:        wall.IsPublic.Bool,
		IsArchived:      wall.IsArchived.Bool,
		IsDeleted:       wall.IsDeleted.Bool,
		PopularityScore: wall.PopularityScore.Float64,
		CreatedAt:       wall.CreatedAt.Time,
		UpdatedAt:       wall.UpdatedAt.Time,
	}
}

// CreateWall handler
func (server *Server) createWall(ctx *gin.Context) {
	var req createWallRequest
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

	arg := db.CreateWallParams{
		UserID:          userID,
		Description:     pgtype.Text{String: req.Description, Valid: req.Description != ""},
		BackgroundImage: pgtype.Text{String: req.BackgroundImage, Valid: req.BackgroundImage != ""},
	}

	wall, err := server.hub.CreateWall(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newWallResponse(wall)
	ctx.JSON(http.StatusCreated, response)
}

// GetWall handler
func (server *Server) getWall(ctx *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(uri.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	wall, err := server.hub.GetWall(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newWallResponse(wall)
	ctx.JSON(http.StatusOK, response)
}

// ListWalls handler
func (server *Server) listWalls(ctx *gin.Context) {
	walls, err := server.hub.ListWalls(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	responses := make([]wallResponse, 0, len(walls))
	for _, wall := range walls {
		responses = append(responses, newWallResponse(wall))
	}

	ctx.JSON(http.StatusOK, responses)
}

// ListWallsByUser handler
func (server *Server) listWallsByUser(ctx *gin.Context) {
	var uri struct {
		UserID string `uri:"user_id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var userID pgtype.UUID
	err := userID.Scan(uri.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	walls, err := server.hub.ListWallsByUser(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	responses := make([]wallResponse, 0, len(walls))
	for _, wall := range walls {
		responses = append(responses, newWallResponse(wall))
	}

    ctx.JSON(http.StatusOK, responses)
}

// UpdateWall handler
func (server *Server) updateWall(ctx *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateWallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(uri.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateWallParams{
		ID:              id,
		Description:     pgtype.Text{String: req.Description, Valid: req.Description != ""},
		BackgroundImage: pgtype.Text{String: req.BackgroundImage, Valid: req.BackgroundImage != ""},
	}

	wall, err := server.hub.UpdateWall(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newWallResponse(wall)
	ctx.JSON(http.StatusOK, response)
}

// PublicizeWall handler
func (server *Server) publicizeWall(ctx *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(uri.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	wall, err := server.hub.PublicizeWall(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newWallResponse(wall)
	ctx.JSON(http.StatusOK, response)
}

// PrivatizeWall handler
func (server *Server) privatizeWall(ctx *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(uri.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	wall, err := server.hub.PrivatizeWall(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newWallResponse(wall)
	ctx.JSON(http.StatusOK, response)
}

// ArchiveWall handler
func (server *Server) archiveWall(ctx *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(uri.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.hub.ArchiveWall(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Wall archived successfully"})
}

// UnarchiveWall handler
func (server *Server) unarchiveWall(ctx *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(uri.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.hub.UnarchiveWall(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Wall unarchived successfully"})
}

// DeleteWall handler
func (server *Server) deleteWall(ctx *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(uri.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.hub.DeleteWall(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Wall deleted successfully"})
}