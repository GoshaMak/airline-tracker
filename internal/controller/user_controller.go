package controller

import (
	"airline-ticketing-svc/internal/domain"
	services "airline-ticketing-svc/internal/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type UserController struct {
	service *services.UserService
}

func NewUserController(i do.Injector) (*UserController, error) {
	return &UserController{
		service: do.MustInvoke[*services.UserService](i),
	}, nil
}

type loginParams struct {
	Email    string `form:"email"`
	Phone    string `form:"phone"`
	Password string `form:"password"`
}

func (c *UserController) Login(ctx *gin.Context) {
	var req loginParams
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, "incorrect args")
		return
	}
	slog.Debug("User", "phone", req.Phone, "email", req.Email, "pswd", req.Password)
	u, err := c.service.GetUser(req.Email, req.Phone, req.Password)
	if err != nil {
		slog.Warn("Can't find user", "err", err)
		ctx.JSON(http.StatusOK, "user was not found")
		return
	}
	ctx.JSON(http.StatusOK, u)
}

func (c *UserController) SignUp(ctx *gin.Context) {
	var u domain.User
	if err := ctx.ShouldBind(&u); err != nil {
		ctx.JSON(http.StatusBadRequest, "failed to parse args")
		return
	}
	slog.Debug("About to create user", "user", u)
	if err := c.service.CreateUser(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, "failed to create a user")
		return
	}
	ctx.JSON(http.StatusOK, "user created")
}
