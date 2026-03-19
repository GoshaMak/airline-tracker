package controller

import (
	"airline-tracker/internal/fleet/dto"
	"airline-tracker/internal/fleet/service"
	"airline-tracker/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type AircraftController struct {
	service *service.AircraftService
}

func NewAircraftController(i do.Injector) (*AircraftController, error) {
	return &AircraftController{
		service: do.MustInvoke[*service.AircraftService](i),
	}, nil
}

func RegisterAircraftRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*AircraftController](i)
	g := r.Group("/admin", middleware.AuthMiddleware("admin"))
	{
		g.POST("/aircraft", c.AddAircraft)
	}
}

func (c *AircraftController) AddAircraft(ctx *gin.Context) {
	type Req struct {
		Aircraft dto.AircraftDTO `json:"aircraft"`
	}
	req := &Req{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	if err := c.service.AddAircraft(req.Aircraft.AircraftFromDTO()); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "aircraft created"})
}
