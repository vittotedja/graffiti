package api

import (
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// Like handlers

type createLikeRequest struct {
	PostID string `json:"post_id" binding:"required"`
	UserID string `json:"user_id" binding:"required"`
}

func (server *Server) createLike(ctx *gin.Context) {
	var req createLikeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var postID pgtype.UUID
	err := postID.Scan(req.PostID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var userID pgtype.UUID
	err = userID.Scan(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateLikeParams{
		PostID: postID,
		UserID: userID,
	}

	like, err := server.hub.CreateLike(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, like)
}

type getLikeRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (server *Server) getLike(ctx *gin.Context) {
	var req getLikeRequest
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

	like, err := server.hub.GetLike(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, like)
}

type deleteLikeRequest struct {
	PostID string `uri:"post_id" binding:"required"`
	UserID string `uri:"user_id" binding:"required"`
}

func (server *Server) deleteLike(ctx *gin.Context) {
	var req deleteLikeRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var postID pgtype.UUID
	err := postID.Scan(req.PostID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var userID pgtype.UUID
	err = userID.Scan(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.DeleteLikeParams{
		PostID: postID,
		UserID: userID,
	}

	err = server.hub.DeleteLike(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Like deleted successfully"})
}

func (server *Server) listLikes(ctx *gin.Context) {
	likes, err := server.hub.ListLikes(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, likes)
}

type listLikesByPostRequest struct {
	PostID string `uri:"post_id" binding:"required"`
}

func (server *Server) listLikesByPost(ctx *gin.Context) {
	var req listLikesByPostRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var postID pgtype.UUID
	err := postID.Scan(req.PostID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	likes, err := server.hub.ListLikesByPost(ctx, postID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, likes)
}

type listLikesByUserRequest struct {
	UserID string `uri:"user_id" binding:"required"`
}

func (server *Server) listLikesByUser(ctx *gin.Context) {
	var req listLikesByUserRequest
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

	likes, err := server.hub.ListLikesByUser(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, likes)
}