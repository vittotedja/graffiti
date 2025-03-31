package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pingcap/log"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// Wall request/response types
type createWallRequest struct {
	UserID          string `json:"user_id" binding:"required,uuid"`
	Description     string `json:"description"`
	BackgroundImage string `json:"background_image"`
}

type createTestWallRequest struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	BackgroundImage string `json:"background_image"`
	IsPublic        bool   `json:"is_public"`
}
type wallResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Title           string    `json:"title"`
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
	Description     *string `json:"description"`
	BackgroundImage *string `json:"background_image"`
}

// Convert DB wall to API response
func newWallResponse(wall db.Wall) wallResponse {
	return wallResponse{
		ID:              wall.ID.String(),
		UserID:          wall.UserID.String(),
		Title:           wall.Title,
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
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received create wall request")

	var req createWallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var userID pgtype.UUID
	if err := userID.Scan(req.UserID); err != nil {
		log.Error("Invalid user_id", err)
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
		log.Error("Failed to create wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Wall created successfully")
	ctx.JSON(http.StatusCreated, newWallResponse(wall))
}

// CreateNewWall handler
func (server *Server) createNewWall(ctx *gin.Context) {
	var req createTestWallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user := ctx.MustGet("currentUser").(db.User)

	arg := db.CreateTestWallParams{
		UserID:      user.ID,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Title:       req.Title,
		IsPublic:    pgtype.Bool{Bool: req.IsPublic, Valid: true},
	}

	wall, err := server.hub.CreateTestWall(ctx, arg)
	if err != nil {
		log.Error("Failed to create wall", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Wall created successfully")
	ctx.JSON(http.StatusCreated, newWallResponse(wall))
}

// GetWall handler
func (server *Server) getWall(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get wall request")

	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	if err := id.Scan(uri.ID); err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	wall, err := server.hub.GetWall(ctx, id)
	if err != nil {
		log.Error("Failed to get wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Wall retrieved successfully")
	ctx.JSON(http.StatusOK, newWallResponse(wall))
}

func (server *Server) getOwnWall(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get own wall request")

	user := ctx.MustGet("currentUser").(db.User)

	walls, err := server.hub.ListWallsByUser(ctx, user.ID)

	if err != nil {
		log.Error("Failed to list own walls", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Walls listed successfully")
	responses := make([]wallResponse, 0, len(walls))
	for _, wall := range walls {
		responses = append(responses, newWallResponse(wall))
	}

	ctx.JSON(http.StatusOK, responses)

}

// ListWalls handler
func (server *Server) listWalls(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list walls request")
	walls, err := server.hub.ListWalls(ctx)
	if err != nil {
		log.Error("Failed to list walls", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Walls listed successfully")
	responses := make([]wallResponse, 0, len(walls))
	for _, wall := range walls {
		responses = append(responses, newWallResponse(wall))
	}

	ctx.JSON(http.StatusOK, responses)
}

// ListWallsByUser handler
func (server *Server) listWallsByUser(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list walls by user request")

	me := ctx.MustGet("currentUser").(db.User)

	var uri struct {
		UserID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var userID pgtype.UUID
	if err := userID.Scan(uri.UserID); err != nil {
		log.Error("Invalid user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	walls, err := server.hub.ListWallsByUser(ctx, userID)
	if err != nil {
		log.Error("Failed to list walls by user", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Walls by user listed successfully")

	params := db.ListFriendshipByUserPairsParams{
		FromUser: me.ID,
		ToUser:   userID,
	}

	_, err = server.hub.Queries.ListFriendshipByUserPairs(ctx, params)

	isFriend := true
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			isFriend = false
		} else {
			log.Error("Failed to list friendship by user pairs", err)
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	var filteredWalls []wallResponse
	if isFriend {
		// If friends, show both public and private walls
		for _, wall := range walls {
			filteredWalls = append(filteredWalls, newWallResponse(wall))
		}
	} else {
		// If not friends, only show public walls
		for _, wall := range walls {
			if wall.IsPublic.Bool {
				filteredWalls = append(filteredWalls, newWallResponse(wall))
			}
		}
	}

	ctx.JSON(http.StatusOK, filteredWalls)
}

// UpdateWall handler
func (server *Server) updateWall(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received update wall request")

	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateWallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	if err := id.Scan(uri.ID); err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	currentWall, err := server.hub.GetWall(ctx, id)
	if err != nil {
		log.Error("Failed to get current wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.UpdateWallParams{
		ID:              id,
		Description:     currentWall.Description,
		BackgroundImage: currentWall.BackgroundImage,
	}

	if req.Description != nil && *req.Description != "" {
		arg.Description = pgtype.Text{String: *req.Description, Valid: true}
	}
	if req.BackgroundImage != nil {
		arg.BackgroundImage = pgtype.Text{String: *req.BackgroundImage, Valid: true}
	}

	wall, err := server.hub.UpdateWall(ctx, arg)
	if err != nil {
		log.Error("Failed to update wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Wall updated successfully")
	ctx.JSON(http.StatusOK, newWallResponse(wall))
}

// PublicizeWall handler
func (server *Server) publicizeWall(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received publicize wall request")

	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	if err := id.Scan(uri.ID); err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	wall, err := server.hub.PublicizeWall(ctx, id)
	if err != nil {
		log.Error("Failed to publicize wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Wall publicized successfully")
	ctx.JSON(http.StatusOK, newWallResponse(wall))
}

// PrivatizeWall handler
func (server *Server) privatizeWall(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received privatize wall request")

	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	if err := id.Scan(uri.ID); err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	wall, err := server.hub.PrivatizeWall(ctx, id)
	if err != nil {
		log.Error("Failed to privatize wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Wall privatized successfully")
	ctx.JSON(http.StatusOK, newWallResponse(wall))
}

// ArchiveWall handler
func (server *Server) archiveWall(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received archive wall request")

	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	if err := id.Scan(uri.ID); err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.hub.ArchiveWall(ctx, id)
	if err != nil {
		log.Error("Failed to archive wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Wall archived successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Wall archived successfully"})
}

// UnarchiveWall handler
func (server *Server) unarchiveWall(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received unarchive wall request")

	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	if err := id.Scan(uri.ID); err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.hub.UnarchiveWall(ctx, id)
	if err != nil {
		log.Error("Failed to unarchive wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Wall unarchived successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Wall unarchived successfully"})
}

// DeleteWall handler
func (server *Server) deleteWall(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received delete wall request")

	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	if err := id.Scan(uri.ID); err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := server.hub.DeleteWall(ctx, id); err != nil {
		log.Error("Failed to delete wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Wall deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Wall deleted successfully"})
}
