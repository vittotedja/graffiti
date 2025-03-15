package api

import (
	"graffiti-backend/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type createUserRequest struct {
	Username           string `json:"username" binding:"required"`
	FullName           string `json:"full_name" binding:"required"`
	Email              string `json:"email" binding:"required"`
	Password           string `json:"password" binding:"required"`
	ProfilePictureURL  string `json:"profile_picture_url"`
	Bio                string `json:"bio"`
	HasOnboarded       bool   `json:"has_onboarded" binding:"oneof=0 1"`
	BackgroundImageURL string `json:"background_image_url"`
	OnboardingTime     time.Time `json:"onboarding_time"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:           req.Username,
		FullName:           req.FullName,
		Email:              req.Email,
		Password:           req.Password,
		ProfilePictureURL:  req.ProfilePictureURL,
		Bio:                req.Bio,
		HasOnboarded:       req.HasOnboarded,
		BackgroundImageURL: req.BackgroundImageURL,
		OnboardingTime:     req.OnboardingTime,
		CreatedAt:          req.CreatedAt,
		UpdatedAt:          req.UpdatedAt,
	}

	user, err := server.hub.CreateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, user)
}