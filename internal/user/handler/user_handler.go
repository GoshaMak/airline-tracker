package handler

import (
	"airline-tracker/internal/middleware"
	"airline-tracker/internal/user/dto"
	"airline-tracker/internal/user/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type UserHandler struct {
	uc *usecase.UserUsecase
}

func NewUserHandler(i do.Injector) (*UserHandler, error) {
	return &UserHandler{
		uc: do.MustInvoke[*usecase.UserUsecase](i),
	}, nil
}

func RegisterRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*UserHandler](i)

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
func (h *UserHandler) ListFlights(ctx *gin.Context) {
	req := &dto.ListFlightsRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, "")
		return
	}
	flights := dto.ListFlightsResponse{}
	ctx.JSON(http.StatusOK, gin.H{"flights": flights})
}
