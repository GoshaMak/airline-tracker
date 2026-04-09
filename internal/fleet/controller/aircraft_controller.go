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
		g.POST("/add_aircraft", c.AddAircraft)
	}
}

// @Summary add aircraft
// @Description create new aircraft
// @Tags aircraft
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param aircraft body dto.CreateAircraftRequest true "aircraft info"
// @Success 201 "created"
// @Failure 400
// @Failure 401
// @Router /admin/add_aircraft [post]
func (c *AircraftController) AddAircraft(ctx *gin.Context) {
	req := &dto.CreateAircraftRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	cmd, err := command.NewCreateAircraftCommand(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := c.service.AddAircraft(cmd); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusCreated, "created")
}
