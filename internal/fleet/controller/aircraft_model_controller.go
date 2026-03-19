package controller

import (
	"airline-tracker/internal/fleet/dto"
	"airline-tracker/internal/fleet/service"
	"airline-tracker/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type AircraftModelController struct {
	service *service.AircraftModelService
}

func NewAircraftModelController(i do.Injector) (*AircraftModelController, error) {
	return &AircraftModelController{
		service: do.MustInvoke[*service.AircraftModelService](i),
	}, nil
}

func RegisterAircraftModelRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*AircraftModelController](i)
	g := r.Group("/admin", middleware.AuthMiddleware("admin"))
	{
		g.POST("/aircraft_model", c.AddAircraftModel)
	}
}

func (c *AircraftModelController) AddAircraftModel(ctx *gin.Context) {
	type Req struct {
		AircraftModel dto.AircraftModelDTO `json:"aircraft_model"`
	}
	req := &Req{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	if err := c.service.AddAircraftModel(req.AircraftModel.AircraftModelFromDTO()); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "aircraft model created"})
}
