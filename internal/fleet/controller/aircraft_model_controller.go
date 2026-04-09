package controller

import (
	"airline-tracker/internal/fleet/command"
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
		g.POST("/add_aircraft_model", c.AddAircraftModel)
	}
}

// @Summary add aircraft model
// @Description create new aircraft model
// @Tags aircraft
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param aircraft_model body dto.CreateAircraftModelRequest true "aircraft model info"
// @Success 201 "created"
// @Failure 400
// @Failure 401
// @Router /admin/add_aircraft_model [post]
func (c *AircraftModelController) AddAircraftModel(ctx *gin.Context) {
	req := &dto.CreateAircraftModelRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args", "err": err})
		return
	}
	cmd, err := command.NewCreateAircraftModelCommand(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid args", "err": err})
		return
	}
	if err := c.service.AddAircraftModel(cmd); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request", "err": err})
		return
	}
	ctx.JSON(http.StatusCreated, "created")
}
