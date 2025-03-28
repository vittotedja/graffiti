package api

import (
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

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

type updateUserRequest struct {
	Username *string `json:"username"`
	Fullname *string `json:"fullname"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type updateProfileRequest struct {
	ProfilePicture  string `json:"profile_picture"`
	Bio             string `json:"bio"`
	BackgroundImage string `json:"background_image"`
}

// createUser handles the creation of a new user
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Hash the password (you should implement password hashing)
	hashedPassword := hashPassword(req.Password)

	arg := db.CreateUserParams{
		Username:       req.Username,
		Fullname:       pgtype.Text{String: req.Fullname, Valid: true},
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	user, err := server.hub.CreateUser(ctx, arg)
	if err != nil {
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

	ctx.JSON(http.StatusOK, resp)
}

// getUser handles retrieving a user by ID
func (server *Server) getUser(ctx *gin.Context) {
	var idReq struct {
		ID string `uri:"id" binding:"required,uuid"`
	}
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(idReq.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.hub.GetUser(ctx, id)
	if err != nil {
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

	ctx.JSON(http.StatusOK, resp)
}

// listUsers handles retrieving a list of users
func (server *Server) listUsers(ctx *gin.Context) {
	meta := logger.GetMetadata(ctx.Request.Context())
	log := meta.GetLogger()

	log.Info("Listing users")
	log.Debug("Request ID: %s", meta.RequestID)
	log.Error("Error message")

	users, err := server.hub.ListUsers(ctx)
	if err != nil {
		log.Error("Failed to list users", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	log.Debug("Retrieved %d users", len(users))

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

	log.Info("Successfully listed users")
	ctx.JSON(http.StatusOK, resp)
}

// updateUser handles updating a user's basic information
func (server *Server) updateUser(ctx *gin.Context) {
	// Extract the user ID from the URI
	var idReq struct {
		ID string `uri:"id" binding:"required,uuid"`
	}
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Parse the request body
	var req updateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Convert the ID to pgtype.UUID
	var id pgtype.UUID
	err := id.Scan(idReq.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Fetch the current user data to retain non-nullable fields
	currentUser, err := server.hub.GetUser(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Prepare the UpdateUserParams struct
	arg := db.UpdateUserParams{
		ID:             id,
		Username:       currentUser.Username,       // Default to current value
		Fullname:       currentUser.Fullname,       // Default to current value
		Email:          currentUser.Email,          // Default to current value
		HashedPassword: currentUser.HashedPassword, // Default to current value
	}

	// Update Username if provided
	if req.Username != nil && *req.Username != "" {
		arg.Username = *req.Username
	}

	// Update Fullname if provided
	if req.Fullname != nil && *req.Fullname != "" {
		arg.Fullname = pgtype.Text{String: *req.Fullname, Valid: true}
	}

	// Update Email if provided
	if req.Email != nil && *req.Email != "" {
		arg.Email = *req.Email
	}

	// Hash the password if it is provided
	if req.Password != nil {
		hashedPassword := hashPassword(*req.Password)
		arg.HashedPassword = hashedPassword // Set the hashed password
	} else {
		// If no password is provided, pass the existing hashed password
		arg.HashedPassword = currentUser.HashedPassword
	}

	user, err := server.hub.UpdateUser(ctx, arg)
	if err != nil {
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

	ctx.JSON(http.StatusOK, resp)
}

// updateProfile handles updating a user's profile information
func (server *Server) updateProfile(ctx *gin.Context) {
	var idReq struct {
		ID string `uri:"id" binding:"required,uuid"`
	}
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(idReq.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateProfileParams{
		ID:              id,
		ProfilePicture:  pgtype.Text{String: req.ProfilePicture, Valid: req.ProfilePicture != ""},
		Bio:             pgtype.Text{String: req.Bio, Valid: req.Bio != ""},
		BackgroundImage: pgtype.Text{String: req.BackgroundImage, Valid: req.BackgroundImage != ""},
	}

	user, err := server.hub.UpdateProfile(ctx, arg)
	if err != nil {
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

	ctx.JSON(http.StatusOK, resp)
}

// finishOnboarding handles marking a user as having completed onboarding
func (server *Server) finishOnboarding(ctx *gin.Context) {
	var idReq struct {
		ID string `uri:"id" binding:"required,uuid"`
	}
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(idReq.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.hub.FinishOnboarding(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "onboarding completed"})
}

// deleteUser handles deleting a user
func (server *Server) deleteUser(ctx *gin.Context) {
	var idReq struct {
		ID string `uri:"id" binding:"required,uuid"`
	}
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var id pgtype.UUID
	err := id.Scan(idReq.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.hub.DeleteUser(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":      id.String(),
		"message": "User deleted successfully!",
	})
}

// Helper function to hash passwords (you should implement a proper password hashing algorithm)
func hashPassword(password string) string {
	// TODO: Implement proper password hashing
	return password
}
