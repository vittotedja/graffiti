package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
)

// Like handlers
type createLikeRequest struct {
	PostID string `json:"post_id" binding:"required"`
	UserID string `json:"user_id" binding:"required"`
}

func (server *Server) createLike(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received create like request")

	var req createLikeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var postID, userID pgtype.UUID
	if err := postID.Scan(req.PostID); err != nil {
		log.Error("Invalid post_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := userID.Scan(req.UserID); err != nil {
		log.Error("Invalid user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateLikeParams{
		PostID: postID,
		UserID: userID,
	}

	like, err := server.hub.CreateLike(ctx, arg)
	if err != nil {
		log.Error("Failed to create like", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Like created successfully")
	ctx.JSON(http.StatusOK, like)
}

type getLikeRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (server *Server) getLike(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get like request")

	var req getLikeRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	if err := id.Scan(req.ID); err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	like, err := server.hub.GetLike(ctx, id)
	if err != nil {
		log.Error("Failed to get like", err)
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	log.Info("Like retrieved successfully")
	ctx.JSON(http.StatusOK, like)
}

type deleteLikeRequest struct {
	PostID string `uri:"post_id" binding:"required"`
	UserID string `uri:"user_id" binding:"required"`
}

func (server *Server) deleteLike(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received delete like request")

	var req deleteLikeRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var postID, userID pgtype.UUID
	if err := postID.Scan(req.PostID); err != nil {
		log.Error("Invalid post_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := userID.Scan(req.UserID); err != nil {
		log.Error("Invalid user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.DeleteLikeParams{
		PostID: postID,
		UserID: userID,
	}

	if err := server.hub.DeleteLike(ctx, arg); err != nil {
		log.Error("Failed to delete like", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Like deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Like deleted successfully"})
}

func (server *Server) listLikes(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list likes request")

	likes, err := server.hub.ListLikes(ctx)
	if err != nil {
		log.Error("Failed to list likes", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Likes listed successfully")
	ctx.JSON(http.StatusOK, likes)
}

type listLikesByPostRequest struct {
	PostID string `uri:"post_id" binding:"required"`
}

func (server *Server) listLikesByPost(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list likes by post request")

	var req listLikesByPostRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var postID pgtype.UUID
	if err := postID.Scan(req.PostID); err != nil {
		log.Error("Invalid post_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	likes, err := server.hub.ListLikesByPost(ctx, postID)
	if err != nil {
		log.Error("Failed to list likes by post", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Likes by post listed successfully")
	ctx.JSON(http.StatusOK, likes)
}

type listLikesByUserRequest struct {
	UserID string `uri:"user_id" binding:"required"`
}

func (server *Server) listLikesByUser(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list likes by user request")

	var req listLikesByUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var userID pgtype.UUID
	if err := userID.Scan(req.UserID); err != nil {
		log.Error("Invalid user_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	likes, err := server.hub.ListLikesByUser(ctx, userID)
	if err != nil {
		log.Error("Failed to list likes by user", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Likes by user listed successfully")
	ctx.JSON(http.StatusOK, likes)
}
