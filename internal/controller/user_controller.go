package controller

import (
	"airline-tracker/internal/domain"
	"airline-tracker/internal/middleware"
	"airline-tracker/internal/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type UserController struct {
	service *service.UserService
}

func NewUserController(i do.Injector) (*UserController, error) {
	return &UserController{
		service: do.MustInvoke[*service.UserService](i),
	}, nil
}

func RegisterUserRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*UserController](i)
	r.POST("/signup", c.Signup)
	r.GET("/login", c.Login)
	r.GET("/user/flights", middleware.AuthMiddleware(), c.Flights)
}

func (c *UserController) Signup(ctx *gin.Context) {
	type signupUserRequest struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	var req signupUserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	u := domain.NewUser(req.Email, req.Phone, req.Password, req.Role)
	if err := c.service.Create(u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to create a user"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "user created"})
}

func (c *UserController) Login(ctx *gin.Context) {
	type loginUserRequest struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}

	var req loginUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "incorrect args"})
		return
	}
	slog.Debug("User", "email", req.Email, "phone", req.Phone, "pswd", req.Password)
	u, err := c.service.Get(req.Email, req.Phone, req.Password)
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

func (c *UserController) Flights(ctx *gin.Context)
