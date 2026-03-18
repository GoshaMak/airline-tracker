package controller

import (
	"airline-tracker/internal/auth/service"
	"airline-tracker/internal/user/dto"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type AuthController struct {
	service *service.AuthService
}

func NewAuthController(i do.Injector) (*AuthController, error) {
	return &AuthController{
		service: do.MustInvoke[*service.AuthService](i),
	}, nil
}

func RegisterAuthRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*AuthController](i)
	r.POST("/auth/signup", c.Signup)
	r.POST("/auth/login", c.Login)
}

func (c *AuthController) Signup(ctx *gin.Context) {
	var req dto.UserDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	u := req.UserFromDTO()
	if err := c.service.CreateUser(u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to create a user"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "user created"})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.UserDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "incorrect args"})
		return
	}
	slog.Debug("Auth", "email", req.Email, "phone", req.Phone, "pswd", req.Password, "role", req.Role)
	u, err := c.service.GetUser(req.Email, req.Phone, req.Password)
	if err != nil {
		slog.Warn("Can't find user", "err", err)
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "user was not found"})
		return
	}
	token, err := service.GenerateJWT(u)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"token": token})
}
