package api

import (
	"errors"
	"net/http"
	"time"

	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

type getUserResponse struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	Fullname        string `json:"fullname"`
	Email           string `json:"email"`
	ProfilePicture  string `json:"profile_picture,omitempty"`
	Bio             string `json:"bio,omitempty"`
	HasOnboarded    bool   `json:"has_onboarded"`
	BackgroundImage string `json:"background_image,omitempty"`
	OnboardingAt    string `json:"onboarding_at,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type updateUserNewRequest struct {
	Username        *string `json:"username"`
	Fullname        *string `json:"fullname"`
	Email           *string `json:"email"`
	Password        *string `json:"password"`
	ProfilePicture  *string `json:"profile_picture"`
	Bio             *string `json:"bio"`
	BackgroundImage *string `json:"background_image"`
}

type UserSearchRequest struct {
	SearchTerm string `json:"search_term" binding:"required"`
}

type UserSearchResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	FullName       string `json:"fullname"`
	ProfilePicture string `json:"profile_picture"`
}

// getUser handles retrieving a user by ID
func (s *Server) getUser(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received get user request")

	_, ok := ctx.MustGet("currentUser").(db.User)
	if !ok {
		log.Error("Failed to get current user from context", errors.New("unauthorized"))
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("unauthorized")))
		return
	}

	var idReq struct {
		ID string `uri:"id" binding:"required,uuid"`
	}
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(idReq.ID)
	if err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := s.hub.GetUser(ctx, id)
	if err != nil {
		log.Error("Failed to get user", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := getUserResponse{
		ID:           user.ID.String(),
		Username:     user.Username,
		Fullname:     user.Fullname.String,
		Email:        user.Email,
		HasOnboarded: user.HasOnboarded.Bool,
		CreatedAt:    user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Time.Format(time.RFC3339),
	}

	if user.ProfilePicture.Valid {
		resp.ProfilePicture = user.ProfilePicture.String
	}
	if user.Bio.Valid {
		resp.Bio = user.Bio.String
	}
	if user.BackgroundImage.Valid {
		resp.BackgroundImage = user.BackgroundImage.String
	}
	if user.OnboardingAt.Valid {
		resp.OnboardingAt = user.OnboardingAt.Time.Format(time.RFC3339)
	}

	log.Info("User retrieved successfully")
	ctx.JSON(http.StatusOK, resp)
}

// listUsers handles retrieving a list of users
func (s *Server) listUsers(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received list users request")

	users, err := s.hub.ListUsers(ctx)
	if err != nil {
		log.Error("Failed to list users", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var resp []getUserResponse
	for _, user := range users {
		item := getUserResponse{
			ID:           user.ID.String(),
			Username:     user.Username,
			Fullname:     user.Fullname.String,
			Email:        user.Email,
			HasOnboarded: user.HasOnboarded.Bool,
			CreatedAt:    user.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:    user.UpdatedAt.Time.Format(time.RFC3339),
		}

		if user.ProfilePicture.Valid {
			item.ProfilePicture = user.ProfilePicture.String
		}
		if user.Bio.Valid {
			item.Bio = user.Bio.String
		}
		if user.BackgroundImage.Valid {
			item.BackgroundImage = user.BackgroundImage.String
		}
		if user.OnboardingAt.Valid {
			item.OnboardingAt = user.OnboardingAt.Time.Format(time.RFC3339)
		}

		resp = append(resp, item)
	}

	log.Info("Users listed successfully")
	ctx.JSON(http.StatusOK, resp)
}

func (s *Server) updateUserNew(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received update user request")

	currentUser := ctx.MustGet("currentUser").(db.User)

	var req updateUserNewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	username := currentUser.Username
	if req.Username != nil && *req.Username != "" {
		username = *req.Username
	}

	fullname := pgtype.Text{String: currentUser.Fullname.String, Valid: true}
	if req.Fullname != nil && *req.Fullname != "" {
		fullname = pgtype.Text{String: *req.Fullname, Valid: true}
	}

	email := currentUser.Email
	if req.Email != nil && *req.Email != "" {
		email = *req.Email
	}

	hashedPassword := currentUser.HashedPassword
	if req.Password != nil && *req.Password != "" {
		hashedPassword = util.HashPassword(*req.Password)
	}

	profilePicture := currentUser.ProfilePicture
	if req.ProfilePicture != nil {
		profilePicture = pgtype.Text{String: *req.ProfilePicture, Valid: true}
	}

	bio := currentUser.Bio
	if req.Bio != nil {
		bio = pgtype.Text{String: *req.Bio, Valid: true}
	}

	backgroundImage := currentUser.BackgroundImage
	if req.BackgroundImage != nil {
		backgroundImage = pgtype.Text{String: *req.BackgroundImage, Valid: true}
	}

	// Call the unified update method
	arg := db.UpdateUserNewParams{
		ID:              currentUser.ID,
		Username:        username,
		Fullname:        fullname,
		Email:           email,
		HashedPassword:  hashedPassword,
		ProfilePicture:  profilePicture,
		Bio:             bio,
		BackgroundImage: backgroundImage,
	}

	user, err := s.hub.UpdateUserNew(ctx, arg)
	if err != nil {
		log.Error("Failed to update user", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := getUserResponse{
		ID:              user.ID.String(),
		Username:        user.Username,
		Fullname:        user.Fullname.String,
		Email:           user.Email,
		CreatedAt:       user.CreatedAt.Time.Format(time.RFC3339),
		ProfilePicture:  user.ProfilePicture.String,
		Bio:             user.Bio.String,
		BackgroundImage: user.BackgroundImage.String,
	}

	if user.OnboardingAt.Valid {
		resp.OnboardingAt = user.OnboardingAt.Time.Format(time.RFC3339)
	}

	log.Info("User updated successfully")
	ctx.JSON(http.StatusOK, resp)
}

// finishOnboarding handles marking a user as having completed onboarding
func (s *Server) finishOnboarding(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received finish onboarding request")

	var idReq struct {
		ID string `uri:"id" binding:"required,uuid"`
	}
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(idReq.ID)
	if err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = s.hub.FinishOnboarding(ctx, id)
	if err != nil {
		log.Error("Failed to finish onboarding", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("Onboarding completed successfully")
	ctx.JSON(http.StatusOK, gin.H{"status": "onboarding completed"})
}

// deleteUser handles deleting a user
func (s *Server) deleteUser(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received delete user request")

	var idReq struct {
		ID string `uri:"id" binding:"required,uuid"`
	}
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		log.Error("Failed to bind URI", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(idReq.ID)
	if err != nil {
		log.Error("Invalid ID", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = s.hub.DeleteUser(ctx, id)
	if err != nil {
		log.Error("Failed to delete user", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Info("User deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{
		"id":      id.String(),
		"message": "User deleted successfully!",
	})
}

func (s *Server) searchUsers(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()
	log.Info("Received search users request")

	user := ctx.MustGet("currentUser").(db.User)

	var req UserSearchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to bind JSON body", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "search_term is required"})
		return
	}

	var (
		rawUsers   any
		err        error
		searchTerm pgtype.Text
	)

	searchTerm.Valid = true
	searchTerm.String = req.SearchTerm

	if len(req.SearchTerm) < 3 {
		rawUsers, err = s.hub.SearchUsersILike(ctx, searchTerm)
	} else {
		rawUsers, err = s.hub.SearchUsersTrigram(ctx, req.SearchTerm)
	}
	if err != nil {
		log.Error("Search query failed", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var respList []UserSearchResponse

	switch v := rawUsers.(type) {
	case []db.SearchUsersTrigramRow:
		for _, u := range v {
			if u.ID == user.ID {
				continue
			}
			respList = append(respList, UserSearchResponse{
				ID:             u.ID.String(),
				Username:       u.Username,
				FullName:       u.Fullname.String,
				ProfilePicture: u.ProfilePicture.String,
			})
		}
	case []db.SearchUsersILikeRow:
		for _, u := range v {
			if u.ID == user.ID {
				continue
			}

			respList = append(respList, UserSearchResponse{
				ID:             u.ID.String(),
				Username:       u.Username,
				FullName:       u.Fullname.String,
				ProfilePicture: u.ProfilePicture.String,
			})
		}
	}

	log.Info("User search returned results successfully")
	ctx.JSON(http.StatusOK, respList)
}
