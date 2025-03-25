package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/token"
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

type loginResponse struct {
	Token string  `json:"token"`
	User  db.User `json:"user"`
}

func (server *Server) Register(ctx *gin.Context) {
	var req registerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, errorResponse(err))
		return
	}
	//check if user exists
	users, err := server.hub.ListUsers(ctx)
	if err != nil {
		ctx.JSON(500, errorResponse(err))
		return
	}

	for _, user := range users {
		if user.Email == req.Email {
			ctx.JSON(400, errors.New("user already exists"))
			return
		}
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

	user, err := server.hub.CreateUser(ctx, arg)

	if err != nil {
		ctx.JSON(500, errorResponse(err))
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

	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) Login(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, errorResponse(err))
		return
	}

	//check if user exists already
	user, err := server.hub.GetUserByEmail(ctx, req.Email)

	if err != nil {
		ctx.JSON(500, errorResponse(err))
		return
	}

	//check password
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(400, errors.New("invalid password"))
		return
	}

	// create token
	maker, err := token.NewJWTMaker("veryverysecretkey")
	if err != nil {
		ctx.JSON(500, errorResponse(err))
		return
	}

	token, _, err := maker.CreateToken(user.Username, time.Hour)

	if err != nil {
		ctx.JSON(500, errorResponse(err))
		return
	}

	resp := loginResponse{
		Token: token,
		User:  user,
	}

	ctx.JSON(http.StatusOK, resp)

}
