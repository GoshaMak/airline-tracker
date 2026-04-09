package controller

import (
	"airline-tracker/internal/middleware"
	"airline-tracker/internal/user/dto"
	"airline-tracker/internal/user/service"
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

func RegisterRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*UserController](i)

	g := r.Group("/user", middleware.AuthMiddleware("user"))
	{
		g.GET("/list_flights", c.ListFlights)
	}
}

// @Summary list flights
// @Description get all flights in which user is subscribed
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param aircraft body dto.ListFlightsRequest true "aircraft info"
// @Success 200
// @Failure 400
// @Failure 401
// @Router /user/list_flights [get]
func (c *UserController) ListFlights(ctx *gin.Context) {
	req := &dto.ListFlightsRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, "")
		return
	}
	flights := dto.ListFlightsResponse{}
	ctx.JSON(http.StatusOK, gin.H{"flights": flights})
}
