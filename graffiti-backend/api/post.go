package api

import (
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// Post request/response types
type createPostRequest struct {
	WallID   string `json:"wall_id" binding:"required,uuid"`
	Author   string `json:"author" binding:"required,uuid"`
	MediaURL string `json:"media_url" binding:"required"`
	PostType string `json:"post_type" binding:"required,oneof=image video text gif"`
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
	MediaURL string `json:"media_url" binding:"required"`
	PostType string `json:"post_type" binding:"required,oneof=image video text gif"`
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

// CreatePost handler
func (server *Server) createPost(ctx *gin.Context) {
	var req createPostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var wallID pgtype.UUID
	err := wallID.Scan(req.WallID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var author pgtype.UUID
	err = author.Scan(req.Author)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	postType := db.PostType(req.PostType)
	
	arg := db.CreatePostParams{
		WallID:   wallID,
		Author:   author,
		MediaUrl: pgtype.Text{String: req.MediaURL, Valid: true},
		PostType: db.NullPostType{PostType: postType, Valid: true},
	}

	post, err := server.hub.CreatePost(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newPostResponse(post)
	ctx.JSON(http.StatusCreated, response)
}

// GetPost handler
func (server *Server) getPost(ctx *gin.Context) {
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

	post, err := server.hub.GetPost(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newPostResponse(post)
	ctx.JSON(http.StatusOK, response)
}

// ListPosts handler
func (server *Server) listPosts(ctx *gin.Context) {
	posts, err := server.hub.ListPosts(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	responses := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, newPostResponse(post))
	}

	ctx.JSON(http.StatusOK, responses)
}

// ListPostsByWall handler
func (server *Server) listPostsByWall(ctx *gin.Context) {
	var uri struct {
		WallID string `uri:"wall_id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var wallID pgtype.UUID
	err := wallID.Scan(uri.WallID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	posts, err := server.hub.ListPostsByWall(ctx, wallID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	responses := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, newPostResponse(post))
	}

	ctx.JSON(http.StatusOK, responses)
}

// GetHighlightedPosts handler
func (server *Server) getHighlightedPosts(ctx *gin.Context) {
	posts, err := server.hub.GetHighlightedPosts(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	responses := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, newPostResponse(post))
	}

	ctx.JSON(http.StatusOK, responses)
}

// UpdatePost handler
func (server *Server) updatePost(ctx *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required,uuid"`
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updatePostRequest
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

	postType := db.PostType(req.PostType)
	
	arg := db.UpdatePostParams{
		ID:       id,
		MediaUrl: pgtype.Text{String: req.MediaURL, Valid: true},
		PostType: db.NullPostType{PostType: postType, Valid: true},
	}

	post, err := server.hub.UpdatePost(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newPostResponse(post)
	ctx.JSON(http.StatusOK, response)
}

// HighlightPost handler
func (server *Server) highlightPost(ctx *gin.Context) {
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

	post, err := server.hub.HighlightPost(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newPostResponse(post)
	ctx.JSON(http.StatusOK, response)
}

// DeletePost handler
func (server *Server) deletePost(ctx *gin.Context) {
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

	err = server.hub.DeletePost(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}