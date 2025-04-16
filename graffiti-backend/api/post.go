package api

import (
	"context"
	"errors"
	"net/http"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pingcap/log"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
)

// Post request/response types
type createPostRequest struct {
	WallID   string `json:"wall_id" binding:"required,uuid"`
	MediaURL string `json:"media_url" binding:"required"`
	PostType string `json:"post_type" binding:"required,oneof=media embed_link"`
}

type postResponse struct {
	ID            string    `json:"id"`
	WallID        string    `json:"wall_id"`
	Author        string    `json:"author"`
	MediaURL      string    `json:"media_url"`
	PostType      string    `json:"post_type"`
	IsHighlighted bool      `json:"is_highlighted"`
	LikesCount    int32     `json:"likes_count"`
	IsDeleted     bool      `json:"is_deleted"`
	CreatedAt     time.Time `json:"created_at"`
}

type updatePostRequest struct {
	MediaURL *string `json:"media_url"`
	PostType *string `json:"post_type" binding:"omitempty,oneof=media embed_link"`
}

// Convert DB post to API response
func newPostResponse(post db.Post) postResponse {
	return postResponse{
		ID:            post.ID.String(),
		WallID:        post.WallID.String(),
		Author:        post.Author.String(),
		MediaURL:      post.MediaUrl.String,
		PostType:      string(post.PostType.PostType),
		IsHighlighted: post.IsHighlighted.Bool,
		LikesCount:    post.LikesCount.Int32,
		IsDeleted:     post.IsDeleted.Bool,
		CreatedAt:     post.CreatedAt.Time,
	}
}

type PostResponseWithAuthor struct {
	ID             string      `json:"id"`
	WallID         string      `json:"wall_id"`
	MediaURL       string      `json:"media_url"`
	PostType       string      `json:"post_type"`
	IsHighlighted  bool        `json:"is_highlighted"`
	LikesCount     int32       `json:"likes_count"`
	IsDeleted      bool        `json:"is_deleted"`
	CreatedAt      time.Time   `json:"created_at"`
	Username       string      `json:"username"`
	ProfilePicture pgtype.Text `json:"profile_picture"`
	Fullname       pgtype.Text `json:"fullname"`
}

func newPostResponseWithAuthor(post db.ListPostsByWallWithAuthorsDetailsRow) PostResponseWithAuthor {
	return PostResponseWithAuthor{
		ID:             post.ID.String(),
		WallID:         post.WallID.String(),
		MediaURL:       post.MediaUrl.String,
		PostType:       string(post.PostType.PostType),
		IsHighlighted:  post.IsHighlighted.Bool,
		LikesCount:     post.LikesCount.Int32,
		IsDeleted:      post.IsDeleted.Bool,
		CreatedAt:      post.CreatedAt.Time,
		Username:       post.Username,
		ProfilePicture: post.ProfilePicture,
		Fullname:       post.Fullname,
	}
}

// CreatePost handler
func (s *Server) createPost(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received create post request")

	currentUser, ok := ctx.MustGet("currentUser").(db.User)
	if !ok {
		log.Error("Failed to get current user from context", errors.New("unauthorized"))
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized")))
		return
	}

	var req createPostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var wallID pgtype.UUID
	if err := wallID.Scan(req.WallID); err != nil {
		log.Error("Invalid wall_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Get the wall to check who owns it
	wall, err := s.hub.GetWall(ctx, wallID)
	if err != nil {
		log.Error("Failed to get wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	postType := db.PostType(req.PostType)
	arg := db.CreatePostParams{
		WallID:   wallID,
		Author:   currentUser.ID,
		MediaUrl: pgtype.Text{String: req.MediaURL, Valid: true},
		PostType: db.NullPostType{PostType: postType, Valid: true},
	}

	post, err := s.hub.CreatePost(ctx, arg)
	if err != nil {
		log.Error("Failed to create post", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Send notification if someone posts on another user's wall
	if wall.UserID.Bytes != currentUser.ID.Bytes {
		// Only send notification if the post author is not the wall owner
		err = s.SendNotification(
			ctx,
			wall.UserID.String(),           // recipient (wall owner)
			currentUser.ID.String(),        // sender (post author)
			"wall_post",                    // notification type
			wallID.String(),               // entity ID (post ID)
			fmt.Sprintf("%s posted on your wall", currentUser.Username), // message
		)

		if err != nil {
			// Just log the error, don't fail the post creation
			log.Error("Failed to send wall post notification", err)
		} else {
			log.Info("Wall post notification sent to user %s", wall.UserID.String())
		}
	}

	log.Info("Post created successfully")
	response := newPostResponse(post)
	ctx.JSON(http.StatusCreated, response)
}

// GetPost handler
func (s *Server) getPost(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get post request")

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

	post, err := s.hub.GetPost(ctx, id)
	if err != nil {
		log.Error("Failed to get post", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Post retrieved successfully")
	response := newPostResponse(post)
	ctx.JSON(http.StatusOK, response)
}

// ListPosts handler
func (s *Server) listPosts(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list posts request")

	posts, err := s.hub.ListPosts(ctx)
	if err != nil {
		log.Error("Failed to list posts", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Posts listed successfully")
	responses := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, newPostResponse(post))
	}

	ctx.JSON(http.StatusOK, responses)
}

// ListPostsByWall handler
func (s *Server) listPostsByWall(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list posts by wall request")

	var uri struct {
		WallID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var wallID pgtype.UUID
	if err := wallID.Scan(uri.WallID); err != nil {
		log.Error("Invalid wall_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	posts, err := s.hub.ListPostsByWall(ctx, wallID)
	if err != nil {
		log.Error("Failed to list posts by wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Posts by wall listed successfully")
	responses := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, newPostResponse(post))
	}

	ctx.JSON(http.StatusOK, responses)
}

func (s *Server) listPostsByWallWithAuthorsDetails(ctx *gin.Context) {
	var uri struct {
		WallID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	_, ok := ctx.MustGet("currentUser").(db.User)
	if !ok {
		log.Error("Failed to get current user from context")
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized")))
		return
	}

	var wallID pgtype.UUID
	err := wallID.Scan(uri.WallID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	posts, err := s.hub.ListPostsByWallWithAuthorsDetails(ctx, wallID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	responses := make([]PostResponseWithAuthor, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, newPostResponseWithAuthor(post))
	}
	ctx.JSON(http.StatusOK, responses)
}

// GetHighlightedPosts handler
func (s *Server) getHighlightedPosts(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get highlighted posts request")

	posts, err := s.hub.GetHighlightedPosts(ctx)
	if err != nil {
		log.Error("Failed to get highlighted posts", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Highlighted posts retrieved successfully")
	responses := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, newPostResponse(post))
	}

	ctx.JSON(http.StatusOK, responses)
}

// GetHighlightedPostsByWall handler
func (s *Server) getHighlightedPostsByWall(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get highlighted posts by wall request")

	var uri struct {
		WallID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var wallID pgtype.UUID
	if err := wallID.Scan(uri.WallID); err != nil {
		log.Error("Invalid wall_id", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	posts, err := s.hub.GetHighlightedPostsByWall(ctx, wallID)
	if err != nil {
		log.Error("Failed to get highlighted posts by wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Highlighted posts by wall retrieved successfully")
	responses := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, newPostResponse(post))
	}

	ctx.JSON(http.StatusOK, responses)
}

// UpdatePost handler
func (s *Server) updatePost(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received update post request")

	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updatePostRequest
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

	currentPost, err := s.hub.GetPost(ctx, id)
	if err != nil {
		log.Error("Failed to get current post", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.UpdatePostParams{
		ID:       id,
		MediaUrl: currentPost.MediaUrl,
		PostType: currentPost.PostType,
	}

	if req.MediaURL != nil {
		arg.MediaUrl = pgtype.Text{String: *req.MediaURL, Valid: true}
	}
	if req.PostType != nil {
		postType := db.PostType(*req.PostType)
		arg.PostType = db.NullPostType{PostType: postType, Valid: true}
	}

	post, err := s.hub.UpdatePost(ctx, arg)
	if err != nil {
		log.Error("Failed to update post", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Post updated successfully")
	response := newPostResponse(post)
	ctx.JSON(http.StatusOK, response)
}

// HighlightPost handler
func (s *Server) highlightPost(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received highlight post request")

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

	post, err := s.hub.HighlightPost(ctx, id)
	if err != nil {
		log.Error("Failed to highlight post", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Post highlighted successfully")
	response := newPostResponse(post)
	ctx.JSON(http.StatusOK, response)
}

// UnhighlightPost handler
func (s *Server) unhighlightPost(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received unhighlight post request")

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

	post, err := s.hub.UnhighlightPost(ctx, id)
	if err != nil {
		log.Error("Failed to unhighlight post", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Post unhighlighted successfully")
	response := newPostResponse(post)
	ctx.JSON(http.StatusOK, response)
}

// DeletePost handler
func (s *Server) deletePost(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received delete post request")

	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
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

	var id pgtype.UUID
	if err := id.Scan(uri.ID); err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	post, err := s.hub.GetPost(ctx, id)
	if err != nil {
		log.Error("Failed to get post", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	wall, err := s.hub.GetWall(ctx, post.WallID)
	if err != nil {
		log.Error("Failed to get wall", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if wall.UserID != currentUser.ID && post.Author != currentUser.ID {
		log.Error("Unauthorized to delete post", errors.New("unauthorized"))
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized")))
		return
	}

	if err := s.hub.DeletePost(ctx, id); err != nil {
		log.Error("Failed to delete post", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if post.PostType.Valid && post.PostType.PostType == db.PostTypeMedia && post.MediaUrl.Valid {
		// Delete media from S3
		key := util.ExtractKeyFromMediaURL(post.MediaUrl.String)
		go func(key string) {
			// Use a background context so it won't get canceled when request is done
			bgCtx := context.Background()

			if err := s.DeleteFile(bgCtx, key); err != nil {
				log.Error("Failed to delete media from S3", err)
			}
		}(key)

	}

	log.Info("Post deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
