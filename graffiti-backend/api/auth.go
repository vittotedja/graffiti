package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *Server) Register(ctx *gin.Context) {
	var req registerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, errorResponse(err))
		return
	}
	//check if user exists
	user, _ := s.hub.GetUserByEmail(ctx, req.Email)
	if user.Email == req.Email {
		ctx.JSON(400, errors.New("user already exists"))
		return
	}

	//create user
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(500, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		Fullname:       pgtype.Text{String: req.Fullname, Valid: true},
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	newUser, err := s.hub.CreateUser(ctx, arg)

	if err != nil {
		ctx.JSON(500, errorResponse(err))
		return
	}

	resp := getUserResponse{
		ID:           newUser.ID.String(),
		Username:     newUser.Username,
		Fullname:     newUser.Fullname.String,
		Email:        newUser.Email,
		HasOnboarded: newUser.HasOnboarded.Bool,
		CreatedAt:    newUser.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    newUser.UpdatedAt.Time.Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *Server) Login(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := s.hub.GetUserByEmail(ctx, req.Email)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := util.CheckPassword(req.Password, user.HashedPassword); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _, err := s.tokenMaker.CreateToken(user.Username, time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	secure := false
	sameSite := http.SameSiteDefaultMode
	domain := ""

	log.Println("Frontend URL:", s.config.FrontendURL)
	log.Println("Is Production:", s.config.IsProduction)

	if s.config.IsProduction {
		secure = true
		sameSite = http.SameSiteNoneMode
		// Extract domain from frontend URL to ensure proper cookie sharing
		// Assuming your FrontendURL is in the format https://yourdomain.com
		// This helps with cross-site cookie sharing between FE and BE
		frontendURL := s.config.FrontendURL
		if frontendURL != "" {
			domain = frontendURL
			// Remove protocol (http:// or https://)
			if i := len("https://"); len(domain) > i && domain[:i] == "https://" {
				domain = domain[i:]
			} else if i := len("http://"); len(domain) > i && domain[:i] == "http://" {
				domain = domain[i:]
			}
			// Remove path and query params if any
			if i := len(domain); i > 0 {
				for j := 0; j < i; j++ {
					if domain[j] == '/' || domain[j] == '?' {
						domain = domain[:j]
						break
					}
				}
			}
		}
	}

	log.Println("Domain for cookie:", domain)

	ctx.SetSameSite(sameSite)
	ctx.SetCookie(
		"token",
		token,
		3600*72,
		"/",
		domain, // Now properly set for production
		secure,
		true,
	)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"user":    user,
	})
}

func (s *Server) Me(ctx *gin.Context) {
	token, err := ctx.Cookie("token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	payload, err := s.tokenMaker.VerifyToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := s.hub.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	resp := getUserResponse{
		ID:              user.ID.String(),
		Username:        user.Username,
		Fullname:        user.Fullname.String,
		Email:           user.Email,
		HasOnboarded:    user.HasOnboarded.Bool,
		CreatedAt:       user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:       user.UpdatedAt.Time.Format(time.RFC3339),
		Bio:             user.Bio.String,
		ProfilePicture:  user.ProfilePicture.String,
		BackgroundImage: user.BackgroundImage.String,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": resp,
	})
}
