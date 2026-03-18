package controller

import (
	"airline-tracker/internal/middleware"
	"airline-tracker/internal/user/service"

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
	r.GET("/user/flights", middleware.AuthMiddleware("user"), c.Flights)
}

func (c *UserController) Flights(ctx *gin.Context) {}
