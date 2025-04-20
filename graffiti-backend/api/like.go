package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
)

// Like handlers
type likeRequest struct {
	PostID string `json:"post_id" binding:"required"`
}

func (s *Server) updateLike(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received create like request")

	var req likeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	currentUser, ok := ctx.MustGet("currentUser").(db.User)
	if !ok {
		log.Error("Failed to get current user from context", errors.New("unauthorized"))
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized")))
		return
	}

	var postID pgtype.UUID
	if err := postID.Scan(req.PostID); err != nil {
		log.Error("Invalid post_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	liked, err := s.hub.CreateOrDeleteLikeTx(ctx, postID, currentUser.ID)
	if err != nil {
		log.Error("Failed to toggle like", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	action := "unliked"
	if liked {
		action = "liked"

		post, err := s.hub.GetPost(ctx, postID)
		if err != nil {
			log.Error("Failed to get post details for notification", err)
		} else {
			if post.Author.Bytes != currentUser.ID.Bytes {
				err = s.SendNotification(
					ctx,
					post.Author.String(),           // recipient (post owner)
					currentUser.ID.String(),        // sender (user who liked)
					"post_like",                    // notification type
					post.WallID.String(),           // entity ID (wall ID)
					fmt.Sprintf("%s liked your post", currentUser.Username), // message
				)

				if err != nil {
					log.Error("Failed to send like notification", err)
				} else {
					log.Info("Like notification sent to user %s", post.Author.String())
				}
			}
		}
	}

	log.Info("Post %s successfully", action)
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Post %s successfully", action)})
}

type getLikeRequest struct {
	PostID string `uri:"post_id" binding:"required"`
}

func (s *Server) getLike(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get like request")

	var req getLikeRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	currentUser, ok := ctx.MustGet("currentUser").(db.User)

	if !ok {
		log.Error("Failed to get current user from context", errors.New("unauthorized"))
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized")))
		return
	}

	var postID pgtype.UUID
	if err := postID.Scan(req.PostID); err != nil {
		log.Error("Invalid post_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetLikeParams{
		PostID: postID,
		UserID: currentUser.ID,
	}

	_, err := s.hub.GetLike(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusOK, gin.H{"liked": false})
			return
		}
		log.Error("Failed to get like", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Like retrieved successfully")
	ctx.JSON(http.StatusOK, gin.H{"liked": true})
}

type deleteLikeRequest struct {
	PostID string `uri:"post_id" binding:"required"`
	UserID string `uri:"user_id" binding:"required"`
}

func (s *Server) deleteLike(ctx *gin.Context) {
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

	if err := s.hub.DeleteLike(ctx, arg); err != nil {
		log.Error("Failed to delete like", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Like deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Like deleted successfully"})
}

func (s *Server) listLikes(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list likes request")

	likes, err := s.hub.ListLikes(ctx)
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

func (s *Server) listLikesByPost(ctx *gin.Context) {
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

	likes, err := s.hub.ListLikesByPost(ctx, postID)
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

func (s *Server) listLikesByUser(ctx *gin.Context) {
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

	likes, err := s.hub.ListLikesByUser(ctx, userID)
	if err != nil {
		log.Error("Failed to list likes by user", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Likes by user listed successfully")
	ctx.JSON(http.StatusOK, likes)
}
